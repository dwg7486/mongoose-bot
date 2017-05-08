package main

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"

	"github.com/bwmarrin/discordgo"

	"bytes"
	"errors"
	"strconv"
	"strings"
)

var (
	eventDB *sql.DB = OpenDB("./db/events.sqlite")
)

// An Event represents a date and time when one or more people will convene at a certain location
type Event struct {
	id          int64
	name        string
	description string
	location    string
	date        string
	time        string
	creator     string
	creatorID   string
}

// An RSVP is a response from a person to a specific Event specifying if they are going, might be going, or not going
type RSVP struct {
	id       int64
	eventID  string
	username string
	userID   string
	status   string
}

// Create a new Event in the DB and return a pointer to it
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

// Get an Event from the DB using either the Event's ID or its name (or part of it)
// Returns a slice in case a name search returns more than one result
func RetrieveEvent(eventSearch string) ([]*Event, error) {
	if _, err := strconv.ParseInt(eventSearch, 10, 64); err != nil {
		// eventSearch is not numeric, use RetrieveEventByName
		return RetrieveEventByName(eventSearch)
	} else {
		event, err := RetrieveEventByID(eventSearch)
		return []*Event{event}, err
	}
}

// Get an Event from the DB using its unique ID
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
	err = result.Scan(&event.id, &event.name, &event.description, &event.location, &event.date, &event.time,
		&event.creator, &event.creatorID)
	if err != nil {
		return nil, err
	}

	return &event, nil
}

// Get an Event from the DB using a search by name or partial name
// Returns a slice in case there are multiple results returned by the search
func RetrieveEventByName(name string) ([]*Event, error) {
	stmt, err := eventDB.Prepare(`SELECT * FROM events WHERE name LIKE ?`)
	if err != nil {
		return nil, err
	}

	result, err := stmt.Query("%" + name + "%")
	if err != nil {
		return nil, err
	}

	var events []*Event

	defer result.Close()
	for result.Next() {
		var event Event
		err := result.Scan(&event.id, &event.name, &event.description, &event.location, &event.date,
			&event.time, &event.creator, &event.creatorID)
		if err != nil {
			return nil, err
		}
		events = append(events, &event)
	}
	err = result.Err()
	if err != nil {
		return nil, err
	}

	return events, nil
}

// Cancel and remove an Event from the DB
func CancelEvent(id string) error {
	stmt, err := eventDB.Prepare(`DELETE FROM events WHERE id=?`)
	if err != nil {
		return err
	}

	_, err = stmt.Exec(id)
	if err != nil {
		return errors.New("Event not found.")
	}

	return nil
}

func UpdateEventDescription(id string, newDescription string) error {
	return updateEventColumn(id, "description", newDescription)
}

func UpdateEventLocation(id string, newLocation string) error {
	return updateEventColumn(id, "location", newLocation)
}

func UpdateEventDate(id string, newDate string) error {
	return updateEventColumn(id, "event_date", newDate)
}

func UpdateEventTime(id string, newTime string) error {
	return updateEventColumn(id, "event_time", newTime)
}

func updateEventColumn(id string, columnName string, newValue string) error {
	stmt, err := eventDB.Prepare(`UPDATE events SET ` + columnName + `=? WHERE id=?`)
	PanicIf(err)

	tx, err := eventDB.Begin()
	PanicIf(err)

	_, err = tx.Stmt(stmt).Exec(newValue, id)

	if err == nil {
		tx.Commit()
	} else {
		tx.Rollback()
	}
	return err
}

func (event *Event) String() string {
	return "__**"+event.name +"**__\n"+
		"**Created by:** "+event.creator+"\n" +
		"**When:** "+event.date+" at "+event.time+"\n"+
		"**Where:** "+event.location+"\n"+
		"**Description:** "+event.description+"\n"
}

// Create an RSVP to the specified Event in the DB
func CreateRSVP(eventID string, username string, userID string, status string) (*RSVP, error) {
	stmt, err := eventDB.Prepare(
		`INSERT INTO rsvps (event_id, username, user_id, status)
        VALUES (?, ?, ?, ?)`)
	if err != nil {
		return nil, err
	}

	tx, err := eventDB.Begin()
	if err != nil {
		return nil, err
	}

	result, err := tx.Stmt(stmt).Exec(eventID, username, userID, status)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	tx.Commit()
	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	rsvp := RSVP{id, eventID, username, userID, status}
	return &rsvp, err
}

// Update an existing RSVP in the DB by setting a new status
func UpdateRSVP(id string, status string) error {
	stmt, err := eventDB.Prepare(`UPDATE rsvps SET status=? WHERE id=?`)
	PanicIf(err)

	tx, err := eventDB.Begin()
	PanicIf(err)

	_, err = tx.Stmt(stmt).Exec(status, id)

	if err == nil {
		tx.Commit()
	} else {
		tx.Rollback()
	}
	return err
}

// Get all RSVPs from the DB for the specified Event
func RetrieveRSVPs(eventID string) ([]*RSVP, error) {
	event, err := RetrieveEventByID(eventID)
	if err != nil {
		return nil, err
	}

	stmt, err := eventDB.Prepare(`SELECT * FROM rsvps WHERE event_id=?`)
	PanicIf(err)

	result, err := stmt.Query(event.id)

	var rsvps []*RSVP

	defer result.Close()
	for result.Next() {
		var rsvp RSVP
		err := result.Scan(&rsvp.id, &rsvp.eventID, &rsvp.username, &rsvp.userID, &rsvp.status)
		if err != nil {
			return nil, err
		}
		rsvps = append(rsvps, &rsvp)
	}
	err = result.Err()
	if err != nil {
		return nil, err
	}

	return rsvps, nil
}

// Parse the command input from a Discord message and execute the command if valid
func HandleCommandInput(s *discordgo.Session, msg *discordgo.MessageCreate) {
	splitIn := strings.SplitN(msg.Content, " ", 2)

	if len(splitIn) != 2 {
		return
	}
	splitCmd := strings.SplitN(splitIn[1], " ", 2)
	cmdKey := splitCmd[0]
	var cmdBody string = ""
	if len(splitCmd) == 2 {
		cmdBody = splitCmd[1]
	}

	switch cmdKey {
	case "create":
		splitBody := strings.Split(cmdBody, "|")
		if len(splitBody) != 5 {
			s.ChannelMessageSend(msg.ChannelID, "Usage: !event create name|description|location|date|time")
			return
		}

		name := splitBody[0]
		description := splitBody[1]
		location := splitBody[2]
		date := splitBody[3]
		time := splitBody[4]

		event, err := CreateEvent(
			name,
			description,
			location,
			date,
			time,
			msg.Author.Username,
			msg.Author.ID)

		// Send a private message to the creator detailing the created Event
		channel, channelErr := s.UserChannelCreate(msg.Author.ID)
		PanicIf(channelErr)
		if err == nil {
			s.ChannelMessageSend(channel.ID,
				"**Created event:** "+name+"\n"+
					"**Description:** "+description+"\n"+
					"**When:** "+event.date+" at "+event.time+"\n"+
					"**Where:** "+location+"\n"+
					"Your event ID is "+strconv.FormatInt(event.id, 10)+".\n"+
					"Remember this ID if you wish to make changes to your event.")
		} else {
			s.ChannelMessageSend(channel.ID,
				"Event creation failed.  Please make sure you are using the command correctly.")
		}

	case "info":
		eventSearch := cmdBody
		events, err := RetrieveEvent(eventSearch)
		if err != nil {
			s.ChannelMessageSend(msg.ChannelID, err.Error())
			return
		}

		if len(events) > 1 {
			// Matched multiple events, display each one with its ID so the user can pick one and retry their command
			var buffer bytes.Buffer
			for _, event := range events {
				buffer.WriteString(
					"ID: " + strconv.FormatInt(event.id, 10) + " `**" + event.name + "** on " + event.date +
						" at " + event.time + "`\n")
			}
			s.ChannelMessageSend(msg.ChannelID, "Your query matched the following events:\n"+
				buffer.String()+"Select one by its ID with !event info <ID> for more information.")

		} else if len(events) == 1 {
			event := events[0]
			// Retrieve the RSVPs for this Event to display with the Event information
			rsvps, err := RetrieveRSVPs(strconv.FormatInt(event.id, 10))
			if err != nil {
				s.ChannelMessageSend(msg.ChannelID, err.Error())
				return
			}
			var buffer bytes.Buffer
			for i, rsvp := range rsvps {
				buffer.WriteString("__" + rsvp.username + "__: " + rsvp.status)
				if i < (len(rsvps) - 1) {
					buffer.WriteString("    ")
				}
			}
			s.ChannelMessageSend(msg.ChannelID, event.String()+"\n"+buffer.String())
		} else {
			s.ChannelMessageSend(msg.ChannelID, "No events found.  Try a different search.")
		}

	case "cancel":
		eventSearch := cmdBody
		events, err := RetrieveEvent(eventSearch)
		if err != nil {
			s.ChannelMessageSend(msg.ChannelID, err.Error())
			return
		}

		if len(events) > 1 {
			// Matched multiple events, display each one with its ID so the user can pick one and retry their command
			var buffer bytes.Buffer
			for _, event := range events {
				buffer.WriteString(
					"ID: " + strconv.FormatInt(event.id, 10) + " `**" + event.name + "** on " + event.date +
						" at " + event.time + "`\n")
			}
			s.ChannelMessageSend(msg.ChannelID, "Your query matched the following events:\n"+
				buffer.String()+"Select one by its ID with !event info <ID> for more information.")

		} else if len(events) == 1 {
			// Cancel the event
			event := events[0]
			if msg.Author.ID == event.creatorID {
				defer CancelEvent(strconv.FormatInt(event.id, 10)) // defer because we want to let RSVPs know first
				if err != nil {
					s.ChannelMessageSend(msg.ChannelID, err.Error())
				} else {
					s.ChannelMessageSend(msg.ChannelID, "Event cancelled.")
				}
			} else {
				s.ChannelMessageSend(msg.ChannelID, "You can't cancel an event you didn't create.")
			}

			// Let everyone who is or might be going to this Event know that it has been cancelled
			rsvps, err := RetrieveRSVPs(strconv.FormatInt(event.id, 10))
			if err != nil {
				s.ChannelMessageSend(msg.ChannelID, err.Error())
				return
			}
			for _, rsvp := range rsvps {
				channel, channelErr := s.UserChannelCreate(rsvp.userID)
				PanicIf(channelErr)
				if err == nil {
					if strings.Compare(rsvp.status, "Going") == 0 || strings.Compare(rsvp.status, "Maybe") == 0 {
						s.ChannelMessageSend(channel.ID, "**"+event.name+"** has been cancelled.\n")
					}
				}
			}
		} else {
			s.ChannelMessageSend(msg.ChannelID, "No events found.  Try a different search.")
		}


	case "edit":
		splitBody := strings.Split(cmdBody, "|")
		if len(splitBody) != 3 {
			s.ChannelMessageSend(msg.ChannelID, "Usage: !event edit eventID|fieldName|newValue\n"+
				"Editable field names are desc[ription], loc[ation], date, and time.  Event name is not editable.")
			return
		}

		eventSearch := splitBody[0]
		column := splitBody[1]
		newValue := splitBody[2]

		events, err := RetrieveEvent(eventSearch)
		if err != nil {
			s.ChannelMessageSend(msg.ChannelID, err.Error())
			return
		}

		if len(events) > 1 {
			// Matched multiple events, display each one with its ID so the user can pick one and retry their command
			var buffer bytes.Buffer
			for _, event := range events {
				buffer.WriteString(
					"ID: " + strconv.FormatInt(event.id, 10) + " `**" + event.name + "** on " + event.date +
						" at " + event.time + "`\n")
			}
			s.ChannelMessageSend(msg.ChannelID, "Your query matched the following events:\n"+
				buffer.String()+"Select one by its ID with !event info <ID> for more information.")

		} else if len(events) == 1 {
			// Update the Event with the new information
			event := events[0]

			if event.creatorID != msg.Author.ID {
				s.ChannelMessageSend(msg.ChannelID, "You can't edit an event you didn't create.")
				return
			}

			idStr := strconv.FormatInt(event.id, 10)

			var msgBuffer bytes.Buffer
			msgBuffer.WriteString("**" + event.name + "** has been updated.\n")

			var err error
			switch strings.ToLower(column) {
			case "description", "desc":
				err = UpdateEventDescription(idStr, newValue)
				msgBuffer.WriteString("New description: " + newValue)
			case "description+", "desc+": // for appending to the description instead of overwriting it
				err = UpdateEventDescription(idStr, event.description+"\n*Update:* "+newValue)
				msgBuffer.WriteString("Update: " + newValue)
			case "location", "loc":
				err = UpdateEventLocation(idStr, newValue)
				msgBuffer.WriteString("New location: " + newValue)
			case "date":
				err = UpdateEventDate(idStr, newValue)
				msgBuffer.WriteString("New date: " + newValue)
			case "time":
				err = UpdateEventTime(idStr, newValue)
				msgBuffer.WriteString("New time: " + newValue)
			default:
				s.ChannelMessageSend(msg.ChannelID,
					"Editable field names are desc[ription], loc[ation], date, and time.  Event name is not editable.")
				return
			}


			// Send a private message to the creator/editor with the status of the update
			channel, channelErr := s.UserChannelCreate(msg.Author.ID)
			PanicIf(channelErr)
			if err == nil {
				s.ChannelMessageSend(channel.ID, event.name+" updated successfully.")
			} else {
				s.ChannelMessageSend(channel.ID, "There was a problem updating "+event.name+
					".  Please make sure you are using the command correctly.")
			}


			// Let everyone who is or might be going to this Event know it has been updated
			rsvps, err := RetrieveRSVPs(strconv.FormatInt(event.id, 10))
			if err != nil {
				s.ChannelMessageSend(msg.ChannelID, err.Error())
				return
			}
			for _, rsvp := range rsvps {
				channel, channelErr := s.UserChannelCreate(rsvp.userID)
				PanicIf(channelErr)
				if err == nil {
					if strings.Compare(rsvp.status, "Going") == 0 || strings.Compare(rsvp.status, "Maybe") == 0 {
						s.ChannelMessageSend(channel.ID, msgBuffer.String())
					}
				}
			}

		} else {
			s.ChannelMessageSend(msg.ChannelID, "No events found.  Try a different search.")
		}

	case "rsvp":
		splitBody := strings.Split(cmdBody, "|")
		if len(splitBody) != 2 {
			s.ChannelMessageSend(msg.ChannelID, "Usage: !event rsvp eventID|choice")
			return
		}

		choice := splitBody[1]
		// Validate RSVP choice
		switch strings.ToLower(choice) {
		case "g", "going":
			choice = "Going"

		case "m", "maybe":
			choice = "Maybe"

		case "n", "not going":
			choice = "Not going"

		default:
			s.ChannelMessageSend(msg.ChannelID,
				"Valid RSVP choices: G[oing], M[aybe], N[ot going]")
			return
		}

		eventSearch := splitBody[0]
		events, err := RetrieveEvent(eventSearch)
		if err != nil {
			s.ChannelMessageSend(msg.ChannelID, err.Error())
			return
		}

		if len(events) > 1 {
			// Matched multiple events, display each one with its ID so the user can pick one and retry their command
			var buffer bytes.Buffer
			for _, event := range events {
				buffer.WriteString(
					"ID: " + strconv.FormatInt(event.id, 10) + " `**" + event.name + "** on " + event.date +
						" at " + event.time + "`\n")
			}
			s.ChannelMessageSend(msg.ChannelID, "Your query matched the following events:\n"+
				buffer.String()+"Select one by its ID with !event info <ID> for more information.")

		} else if len(events) == 1 {
			// Update the RSVP if it already exists
			event := events[0]
			rsvps, err := RetrieveRSVPs(strconv.FormatInt(event.id, 10))
			for _, rsvp := range rsvps {
				if rsvp.userID == msg.Author.ID {
					UpdateRSVP(strconv.FormatInt(rsvp.id, 10), choice)
					return
				}
			}

			// Create a RSVP because one didn't exist already
			rsvp, err := CreateRSVP(strconv.FormatInt(event.id, 10), msg.Author.Username, msg.Author.ID, choice)

			channel, channelErr := s.UserChannelCreate(msg.Author.ID)
			PanicIf(channelErr)
			if err == nil {
				s.ChannelMessageSend(channel.ID, "RSVP submitted - "+event.name+": "+rsvp.status)
			} else {
				s.ChannelMessageSend(msg.ChannelID,
					"Submit RSVP failed.  Please make sure you are using the command correctly")
				return
			}

		} else {
			s.ChannelMessageSend(msg.ChannelID, "No events found.  Try a different search.")
		}

	case "help":
		s.ChannelMessageSend(msg.ChannelID,
			"__Discord Event Planner created by Mongoose__"+"\n```"+
				"**Create event:** !event create name|description|location|date|time"+"\n"+
				"**Edit event:**   !event edit eventID|fieldName|newValue"+"\n"+
				"**Cancel event**  !event cancel eventID"+"\n"+
				"**Show event**    !event info eventID  OR  !event info eventName"+"\n"+
				"**RSVP**          !event rsvp eventID|choice OR  !rsvp eventName|choice"+"\n"+
				"**RSVP choices:** G[oing], M[aybe], N[ot going]"+"```")
	}
}
