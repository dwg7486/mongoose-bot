package main

import (
	"database/sql"
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
	event_time  string
	event_date  string
	creator     string
}

func CreateEvent(name, description, location, event_time, event_date, creator string) *Event {
	var event Event
	stmt, err := eventDB.Prepare(
		`INSERT INTO events (name, description, location, event_time, event_date, creator)
        VALUES (?, ?, ?, ?, ?, ?)`)

	PanicIf(err)

	tx, err := eventDB.Begin()

	PanicIf(err)

	result, err := tx.Stmt(stmt).Exec(name, description, location, event_time, event_date, creator)

	if err == nil {
		tx.Commit()
		id, err := result.LastInsertId()
		fmt.Println(id)
		PanicIf(err)
		event = Event{id,name,description,location,event_time,event_date, creator}
	} else {
		tx.Rollback()
	}
	return &event
}
