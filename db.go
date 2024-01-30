package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/bwmarrin/discordgo"
)

// SQLConn is a wrapper around the database connection
type SQLConn struct {
	conn   *sql.DB
	qTable string
	iTable string
}

type HandlerConext struct {
	Session *discordgo.Session
	DB      *SQLConn
}

// newSQLConn creates a new connection to the database
func newSQLConn() (*SQLConn, error) {
	sqliteFile := os.Getenv("SQLITE_DB")
	q := os.Getenv("SQLITE_QUOTE_TABLE_NAME")
	i := os.Getenv("SQLITE_INCIDENT_TABLE_NAME")

	if sqliteFile == "" {
		return nil, fmt.Errorf("sqlite file not found in .env")
	}

	db, err := sql.Open("sqlite3", sqliteFile)
	if err != nil {
		return nil, err
	}

	log.Printf("Connected to SQLite database %s", sqliteFile)

	return &SQLConn{conn: db, qTable: q, iTable: i}, nil
}
