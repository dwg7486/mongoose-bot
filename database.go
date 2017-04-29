package main

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
)

func OpenDB( dbPath string ) *sql.DB {
	db, err := sql.Open("sqlite3", dbPath)
	PanicIf( err )
	return db;
}