package main

import (
	"database/sql"
	"errors"
	_ "github.com/mattn/go-sqlite3"
	"fmt"
)

var (
	eventDB *sql.DB = OpenDB("./db/events.db")
)

type Event struct {
	id          int64
	name        string
	description string
	location    string
	eventDate   string
	eventTime   string
	creator     string
	creatorID   string
}

func CreateEvent(name, description, location, event_date, event_time, creator, creator_id string) (*Event, error) {
	stmt, err := eventDB.Prepare(
		`INSERT INTO events (name, description, location, event_date, event_time, creator, creator_id)
        VALUES (?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return nil, err
	}

	tx, err := eventDB.Begin()
	if err != nil {
		return nil, err
	}

	result, err := tx.Stmt(stmt).Exec(name, description, location, event_date, event_time, creator, creator_id)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	tx.Commit()
	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	event := Event{id, name, description, location, event_date, event_time, creator, creator_id}
	return &event, err
}

func RetrieveEventByID(id string) (*Event, error) {
	stmt, err := eventDB.Prepare(`SELECT * FROM events WHERE id=?`)
	if err != nil {
		return nil, err
	}

	result := stmt.QueryRow(id)
	if result == nil {
		return nil, errors.New("Event not found.")
	}

	var event Event
	err = result.Scan(&event.id, &event.name, &event.description, &event.location, &event.eventDate, &event.eventTime,
		&event.creator, &event.creatorID)
	if err != nil {
		return nil, err
	}

	return &event, nil
}

func RetrieveEventByName(name string) ([]*Event, error) {
	stmt, err := eventDB.Prepare(`SELECT * FROM events WHERE name LIKE ?`)
	if err != nil {
		return nil, err
	}

	result, err := stmt.Query("%"+name+"%")
	if err != nil {
		return nil, err
	}

	var events []*Event

	defer result.Close()
	for result.Next() {
		var event Event
		err := result.Scan(&event.id, &event.name, &event.description, &event.location, &event.eventDate,
			&event.eventTime, &event.creator, &event.creatorID)
		if err != nil {
			return nil, err
		}
		fmt.Println(event)
		events = append(events, &event)
		fmt.Println(events)
	}
	err = result.Err()
	if err != nil {
		return nil, err
	}

	return events, nil
}

func UpdateEventDescription(id int, newDescription string) error {
	return UpdateEventColumn(id, "description", newDescription)
}

func UpdateEventLocation(id int, newLocation string) error {
	return UpdateEventColumn(id, "location", newLocation)
}

func UpdateEventDate(id int, newDate string) error {
	return UpdateEventColumn(id, "event_date", newDate)
}

func UpdateEventTime(id int, newTime string) error {
	return UpdateEventColumn(id, "event_time", newTime)
}

func UpdateEventColumn(id int, columnName string, newValue string) error {
	stmt, err := eventDB.Prepare(`UPDATE events SET ?=? WHERE id=?`)
	PanicIf(err)

	tx, err := eventDB.Begin()
	PanicIf(err)

	_, err = tx.Stmt(stmt).Exec(columnName, newValue, id)

	if err == nil {
		tx.Commit()
	} else {
		tx.Rollback()
	}
	return err
}

func (event *Event) String() string {
	return "__**" + event.name + "**__\n" +
		"**Created by:** " + event.creator + "\n" +
		"**When:** " + event.eventDate + " at " + event.eventTime + "\n" +
		"**Where:** " + event.location + "\n" +
		"**Description:** " + event.description + "\n"
}
