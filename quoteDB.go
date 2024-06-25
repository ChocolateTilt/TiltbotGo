package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	_ "github.com/ncruces/go-sqlite3"
	_ "github.com/ncruces/go-sqlite3/driver"
	_ "github.com/ncruces/go-sqlite3/embed"
)

// Quote is the field structure for the "Quote" table
type Quote struct {
	CreatedAt time.Time
	Quote     string
	Quotee    string
	Quoter    string
}

// QuoteCache is a cache for the total number of quotes and the last time it was updated in the database
type QuoteCache struct {
	Total       int
	LastUpdated time.Time
}

// SQLConn is a wrapper around the database connection
type SQLConn struct {
	Conn  *sql.DB
	Table string
	Cache *QuoteCache
}

// newSQLConn creates a new connection to the database
func newSQLConn() (*SQLConn, error) {
	sqliteFile := os.Getenv("SQLITE_DB")
	table := os.Getenv("SQLITE_TABLE_NAME")

	if sqliteFile == "" {
		return nil, fmt.Errorf("sqlite file not found in .env")
	}

	db, err := sql.Open("sqlite3", sqliteFile)
	if err != nil {
		return nil, err
	}

	log.Printf("Connected to SQLite database %s", sqliteFile)

	return &SQLConn{Conn: db, Table: table, Cache: &QuoteCache{}}, nil
}

// createQuote creates a quote in the database
func (db *SQLConn) createQuote(ctx context.Context, quote Quote) error {
	// reset the cache timer
	db.Cache.LastUpdated = time.Time{}

	log.Printf("Creating quote: %v", quote)

	query := fmt.Sprintf(`INSERT INTO %s (quote, quotee, quoter, createdAt) VALUES (?, ?, ?, ?)`, db.Table)
	_, err := db.Conn.ExecContext(ctx, query, quote.Quote, quote.Quotee, quote.Quoter, quote.CreatedAt)
	if err != nil {
		log.Printf("Error creating quote: %v", err)
		return err
	}

	return nil
}

// getRandQuote gets a quote from the database
func (db *SQLConn) getRandQuote(ctx context.Context) (Quote, error) {
	var quote Quote
	query := fmt.Sprintf(`SELECT quote,quotee,quoter,createdAt FROM %s ORDER BY RANDOM() LIMIT 1`, db.Table)
	row := db.Conn.QueryRowContext(ctx, query)
	if err := row.Scan(&quote.Quote, &quote.Quotee, &quote.Quoter, &quote.CreatedAt); err != nil {
		return quote, err
	}

	return quote, nil
}

// getRandUserQuote gets a quote from the database for a specific user
func (db *SQLConn) getRandUserQuote(ctx context.Context, quotee string) (Quote, error) {
	var quote Quote
	id := fmt.Sprintf("<@%s>", quotee)
	query := fmt.Sprintf(`SELECT quote,quotee,quoter,createdAt FROM %s WHERE quotee = ? ORDER BY RANDOM() LIMIT 1`, db.Table)
	err := db.Conn.QueryRowContext(ctx, query, id).Scan(&quote.Quote, &quote.Quotee, &quote.Quoter, &quote.CreatedAt)
	if err != nil {
		return quote, err
	}

	return quote, nil
}

// getLatestUserQuote gets the latest quote from the database for a specific user
func (db *SQLConn) getLatestUserQuote(ctx context.Context, quotee string) (Quote, error) {
	var quote Quote
	id := fmt.Sprintf("<@%s>", quotee)
	query := fmt.Sprintf(`SELECT quote,quotee,quoter,createdAt FROM %s WHERE quotee = ? ORDER BY id DESC LIMIT 1`, db.Table)
	err := db.Conn.QueryRowContext(ctx, query, id).Scan(&quote.Quote, &quote.Quotee, &quote.Quoter, &quote.CreatedAt)
	if err != nil {
		return quote, err
	}

	return quote, nil
}

// getLatestQuote gets the latest quote from the database
func (db *SQLConn) getLatestQuote(ctx context.Context) (Quote, error) {
	var quote Quote
	query := fmt.Sprintf(`SELECT quote,quotee,quoter,createdAt FROM %s ORDER BY id DESC LIMIT 1`, db.Table)
	err := db.Conn.QueryRowContext(ctx, query).Scan(&quote.Quote, &quote.Quotee, &quote.Quoter, &quote.CreatedAt)
	if err != nil {
		return quote, err
	}

	return quote, nil
}

// searchQuote searches the database for string (s) and returns the top 10 results
func (db *SQLConn) searchQuote(ctx context.Context, s string) ([]Quote, error) {
	var quotes []Quote
	query := fmt.Sprintf(`SELECT quote,quotee,quoter,createdAt FROM %s WHERE quote LIKE ? ORDER BY id DESC LIMIT 10`, db.Table)
	rows, err := db.Conn.QueryContext(ctx, query, "%"+s+"%")
	if err != nil {
		return quotes, err
	}

	defer rows.Close()

	for rows.Next() {
		var quote Quote
		err := rows.Scan(&quote.Quote, &quote.Quotee, &quote.Quoter, &quote.CreatedAt)
		if err != nil {
			return quotes, err
		}
		quotes = append(quotes, quote)
	}

	if err = rows.Err(); err != nil {
		return quotes, err
	}

	return quotes, nil
}

// getLeaderboard generates a leaderboard of the top 10 quotees
func (db *SQLConn) getLeaderboard(ctx context.Context) (string, error) {
	var leaderboard []string
	var cleanLB string

	query := fmt.Sprintf(`SELECT quotee, COUNT(*) as count FROM %s GROUP BY quotee ORDER BY count DESC LIMIT 10`, db.Table)
	rows, err := db.Conn.QueryContext(ctx, query)
	if err != nil {
		return cleanLB, fmt.Errorf("error getting leaderboard: %w", err)
	}

	defer rows.Close()

	position := 0
	for rows.Next() {
		var quotee string
		var count int
		position++
		err := rows.Scan(&quotee, &count)
		if err != nil {
			return cleanLB, fmt.Errorf("error scanning leaderboard row: %w", err)
		}
		leaderboard = append(leaderboard, fmt.Sprintf("`%d:` %s: %d", position, quotee, count))
	}

	if err = rows.Err(); err != nil {
		return cleanLB, fmt.Errorf("error iterating over leaderboard rows: %w", err)
	}

	cleanLB = strings.Join(leaderboard, "\n")

	return cleanLB, nil
}

// quoteCount gets the number of quotes in the database. It caches the max count for one hour.
func (db *SQLConn) quoteCount(ctx context.Context) (int, error) {
	// if the cache is less than an hour old, return the cached value
	if time.Since(db.Cache.LastUpdated) < time.Hour {
		log.Println("Returning cached total quotes.")
		return db.Cache.Total, nil
	}

	log.Println("Cache is older than an hour. Fetching total quotes from database")

	var count int
	query := fmt.Sprintf(`SELECT COUNT(*) FROM %s`, db.Table)
	err := db.Conn.QueryRowContext(ctx, query).Scan(&count)
	if err != nil {
		return 0, err
	}

	// cache the max count
	db.Cache.Total = count
	db.Cache.LastUpdated = time.Now()
	log.Printf("Cached total quotes at %v. Number of quotes: %d", db.Cache.LastUpdated, db.Cache.Total)

	return count, nil
}
