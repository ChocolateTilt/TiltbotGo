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
	Attendees   string
	Description string
}

type IncidentCache struct {
	Total       int
	LastUpdated time.Time
}

var inC IncidentCache

// createIncident creates an incident in the database
func (db *SQLConn) createIncident(ctx context.Context, i Incident) error {
	// reset the cache timer
	inC.LastUpdated = time.Time{}

	log.Printf("Creating incident: %v", i)

	query := fmt.Sprintf(`INSERT INTO %s (name,attendees,description,createdAt) VALUES (?,?,?,?)`, db.iTable)
	_, err := db.conn.ExecContext(ctx, query, i.Name, i.Attendees, i.Description, i.CreatedAt)
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
	if time.Since(inC.LastUpdated) < time.Hour {
		return inC.Total, nil
	}

	var count int
	query := fmt.Sprintf(`SELECT COUNT(*) FROM %s`, db.iTable)
	err := db.conn.QueryRowContext(ctx, query).Scan(&count)
	if err != nil {
		return count, err
	}

	// cache the max count
	inC.Total = count
	inC.LastUpdated = time.Now()
	log.Printf("Cached %d incidents at %s", count, inC.LastUpdated)

	return count, nil
}
