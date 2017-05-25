package main

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"strings"
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

func DetectRepost(message string) (bool, error) {
	stmt, err := messageDB.Prepare(`SELECT message FROM messages WHERE message = ?`)
	if err != nil {
		return false, err
	}

	result, err := stmt.Query(message)

	if err != nil {
		return false, err
	}

	defer result.Close()
	for result.Next() {
		var m string
		err := result.Scan(&m)
		if err != nil {
			return false, err
		}
		if strings.Compare(message, m) == 0 {
			return true, nil
		}
	}
	return false, nil
}