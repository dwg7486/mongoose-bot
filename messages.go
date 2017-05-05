package main

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
)

var (
	messageDB *sql.DB = OpenDB("./db/messages.sqlite")
)

func RecordMessage(authorID string, message string) error {
	stmt, err := messageDB.Prepare(`INSERT INTO messages (author_id, message) VALUES (?,?)`)
	if err != nil {
		return err
	}

	tx, err := messageDB.Begin()
	if err != nil {
		return err
	}

	_, err = tx.Stmt(stmt).Exec(authorID, message)
	if err != nil {
		tx.Rollback()
		return err
	}

	tx.Commit()
	return nil
}