package main

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"errors"
	"time"
	"github.com/bwmarrin/discordgo"
	"strings"
)

var (
	remindersDB *sql.DB = OpenDB("./db/reminders.sqlite")
)

type Reminder struct {
	id				int64
	userID			string
	remindDateTime	string
	message			string
}

func CreateReminder(userID string, remindDateTime string, message string) error {
	stmt, err := remindersDB.Prepare(`INSERT INTO reminders (user_id, remind_datetime, message) VALUES (?,?,?)`)
	if err != nil {
		return err
	}

	tx, err := remindersDB.Begin()
	if err != nil {
		return err
	}

	_, err = tx.Stmt(stmt).Exec(userID, remindDateTime, message)
	if err != nil {
		tx.Rollback()
		return err
	}

	tx.Commit()
	return nil
}

func GetDueReminders() ([]*Reminder, error) {
	stmt, err := remindersDB.Prepare(`SELECT * FROM reminders WHERE remind_datetime < DATE('now')`)
	if err != nil {
		return nil, err
	}

	result, err := stmt.Query()
	if err != nil {
		return nil, err
	}

	var reminders []*Reminder

	defer result.Close()
	for result.Next() {
		var reminder Reminder
		err := result.Scan(&reminder.id, &reminder.userID, &reminder.remindDateTime, &reminder.message)
		if err != nil {
			return nil, err
		}
		reminders = append(reminders, &reminder)
	}
	err = result.Err()
	if err != nil {
		return nil, err
	}

	return reminders, nil
}

func ProcessReminders() error {
	for {
		reminders, err := GetDueReminders()
		if err != nil {
			return err
		}

		for _, reminder := range reminders {
			channel, channelErr := session.UserChannelCreate(reminder.userID)
			PanicIf(channelErr)

			session.ChannelMessageSend(channel.ID, "Reminder: \n"+reminder.message)
			RemoveReminder(reminder.id)
		}
		time.Sleep(time.Minute)
	}
}

func RemoveReminder(id int64) error {
	stmt, err := remindersDB.Prepare(`DELETE FROM reminders WHERE id=?`)
	if err != nil {
		return err
	}

	_, err = stmt.Exec(id)
	if err != nil {
		return errors.New("Reminder not found.")
	}

	return nil
}

func (reminder *Reminder) String() string {
	return reminder.message
}

func HandleReminderInput(s *discordgo.Session, msg *discordgo.MessageCreate) {
	splitIn := strings.SplitN(msg.Content, " ", 2)

	if len(splitIn) != 2 {
		return
	}

	splitCmd := strings.SplitN(splitIn[1], " ", 2)

	if len(splitCmd) != 2 {
		s.ChannelMessageSend(msg.ChannelID,
			"Usage: !remindme time message\n"+
			"    Valid time format: MM-DD-YYYY|hh:mm")
		return
	}

	loc, _ := time.LoadLocation("America/New_York")
	remindTime, err := time.ParseInLocation("01-02-06|03:04", splitCmd[0], loc)
	if err != nil {
		s.ChannelMessageSend(msg.ChannelID,
			"Usage: !remindme time message\n"+
			"    Valid time format: MM-DD-YYYY|hh:mm")
	}

	err = CreateReminder(msg.Author.ID, remindTime.String(), splitCmd[1])
	if err != nil {
		return
	}
}