package main

import (
	"context"
	"fmt"
	"log"
	"time"
)

type Incident struct {
	CreatedAt   time.Time
	Name        string
	Attendees   []string
	Description string
}

var (
	iMax  int       // cached number of incidents in the database
	iMaxT time.Time // time the max number of incidents was cached
)

// createIncident creates an incident in the database
func (db *SQLConn) createIncident(ctx context.Context, i Incident) error {
	// reset the cache timer
	iMaxT = time.Time{}

	log.Printf("Creating incident: %v", i)

	query := fmt.Sprintf(`INSERT INTO %s (name,attendees,description,createdAt) VALUES (?, ?, ?,?)`, db.iTable)
	_, err := db.conn.ExecContext(ctx, query, i.Name, i.Attendees, i.Description)
	if err != nil {
		log.Printf("Error creating incident: %v", err)
		return err
	}
	return nil
}

// getRandIncident gets an incident from the database
func (db *SQLConn) getRandIncident(ctx context.Context) (Incident, error) {
	var incident Incident
	query := fmt.Sprintf(`SELECT name,attendees,description FROM %s ORDER BY RANDOM() LIMIT 1`, db.iTable)
	row := db.conn.QueryRowContext(ctx, query)
	if err := row.Scan(&incident.Name, &incident.Attendees, &incident.Description); err != nil {
		return incident, err
	}

	return incident, nil
}

// countIncidents counts the number of incidents in the database
func (db *SQLConn) countIncidents(ctx context.Context) (int, error) {
	// if the cache is less than an hour old, return the cached value
	if time.Since(iMaxT) < time.Hour {
		return iMax, nil
	}

	var count int
	query := fmt.Sprintf(`SELECT COUNT(*) FROM %s`, db.iTable)
	err := db.conn.QueryRowContext(ctx, query).Scan(&count)
	if err != nil {
		return count, err
	}

	// cache the max count
	iMax = count
	iMaxT = time.Now()
	log.Printf("Cached %d incidents", count)

	return count, nil
}
