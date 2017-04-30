package main

import (
	"database/sql"
	"errors"
	_ "github.com/mattn/go-sqlite3"
	"strconv"
)

var (
	eventDB *sql.DB = OpenDB("./db/events.db")
)

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

type RSVP struct {
	id			int64
	eventID		string
	username	string
	userID		string
	status		string
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

func RetrieveEvent(eventSearch string) ([]*Event, error) {
	if _, err := strconv.ParseInt(eventSearch, 10, 64); err != nil {
		// eventSearch is not numeric, use RetrieveEventByName
		return RetrieveEventByName(eventSearch)
	} else {
		event, err := RetrieveEventByID(eventSearch)
		return []*Event{event}, err
	}
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
	err = result.Scan(&event.id, &event.name, &event.description, &event.location, &event.date, &event.time,
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
	stmt, err := eventDB.Prepare(`UPDATE events SET `+columnName+`=? WHERE id=?`)
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
	return "__**" + event.name + "**__\n" +
		"**Created by:** " + event.creator + "\n" +
		"**When:** " + event.date + " at " + event.time + "\n" +
		"**Where:** " + event.location + "\n" +
		"**Description:** " + event.description + "\n"
}

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