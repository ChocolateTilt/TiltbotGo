package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/ncruces/go-sqlite3"
)

// Quote is the field structure for the "Quote" table
type Quote struct {
	CreatedAt time.Time
	Quote     string
	Quotee    string
	Quoter    string
}

var (
	dbMax  int
	dbMaxT time.Time
	db     *sql.DB
	dbName = os.Getenv("SQLITE_TABLE_NAME")
)

func connectSQLite() error {
	sqliteFile := os.Getenv("SQLITE_DB")

	if sqliteFile == "" {
		return fmt.Errorf("sqlite file not found in .env")
	}

	var err error
	db, err = sql.Open("sqlite3", sqliteFile)
	if err != nil {
		return fmt.Errorf("error connecting to SQLite: %w", err)
	}

	return nil
}

func createQuote(ctx context.Context, quote Quote) error {
	// reset the cache timer
	dbMaxT = time.Time{}

	query := fmt.Sprintf(`INSERT INTO %s (quote, quotee, quoter, createdAt) VALUES (?, ?, ?)`, dbName)
	_, err := db.ExecContext(ctx, query, quote.Quote, quote.Quotee, quote.Quoter, quote.CreatedAt)
	if err != nil {
		return fmt.Errorf("problem while creating a quote in the collection: %v", err)
	}
	return nil
}

func getRandQuote(ctx context.Context) (Quote, error) {
	var quote Quote
	query := fmt.Sprintf(`SELECT quote,quotee,quoter,createdAt FROM %s ORDER BY RANDOM() LIMIT 1`, dbName)
	err := db.QueryRowContext(ctx, query).Scan(&quote.Quote, &quote.Quotee, &quote.Quoter, &quote.CreatedAt)
	if err != nil {
		return quote, fmt.Errorf("error getting quote: %w", err)
	}
	return quote, nil
}

func getRandUserQuote(ctx context.Context, quotee string) (Quote, error) {
	var quote Quote
	query := fmt.Sprintf(`SELECT quote,quotee,quoter,createdAt FROM %s WHERE quotee = ? ORDER BY RANDOM() LIMIT 1`, dbName)
	err := db.QueryRowContext(ctx, query, quotee).Scan(&quote.Quote, &quote.Quotee, &quote.Quoter, &quote.CreatedAt)
	if err != nil {
		return quote, fmt.Errorf("error getting quote: %w", err)
	}
	return quote, nil
}

func getLatestUserQuote(ctx context.Context, quotee string) (Quote, error) {
	var quote Quote
	query := fmt.Sprintf(`SELECT quote,quotee,quoter,createdAt FROM %s WHERE quotee = ? ORDER BY createdAt DESC LIMIT 1`, dbName)
	err := db.QueryRowContext(ctx, query, quotee).Scan(&quote.Quote, &quote.Quotee, &quote.Quoter, &quote.CreatedAt)
	if err != nil {
		return quote, fmt.Errorf("error getting latest quote: %w", err)
	}
	return quote, nil
}

func getLatestQuote(ctx context.Context) (Quote, error) {
	var quote Quote
	query := fmt.Sprintf(`SELECT quote,quotee,quoter,createdAt FROM %s ORDER BY id DESC LIMIT 1`, dbName)
	err := db.QueryRowContext(ctx, query).Scan(&quote.Quote, &quote.Quotee, &quote.Quoter, &quote.CreatedAt)
	if err != nil {
		return quote, fmt.Errorf("error getting latest quote: %w", err)
	}
	return quote, nil
}

// func getLeaderboard(ctx context.Context) ([]string, error) {
// 	var leaderboard []string
// 	query := fmt.Sprintf(`SELECT quotee, COUNT(*) as count FROM %s GROUP BY quotee ORDER BY count DESC LIMIT 10`, dbName)
// 	rows, err := db.QueryContext(ctx, query)
// 	if err != nil {
// 		return nil, fmt.Errorf("error getting leaderboard: %w", err)
// 	}
// 	defer rows.Close()

// 	for rows.Next() {
// 		var quotee string
// 		var count int
// 		err := rows.Scan(&quotee, &count)
// 		if err != nil {
// 			return nil, fmt.Errorf("error scanning leaderboard row: %w", err)
// 		}
// 		leaderboard = append(leaderboard, fmt.Sprintf("%s: %d", quotee, count))
// 	}

// 	if err = rows.Err(); err != nil {
// 		return nil, fmt.Errorf("error iterating over leaderboard rows: %w", err)
// 	}

// 	return leaderboard, nil
// }

func quoteCount(ctx context.Context) (int, error) {
	var count int
	query := fmt.Sprintf(`SELECT COUNT(*) FROM %s`, dbName)
	err := db.QueryRowContext(ctx, query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("error counting quotes for user: %w", err)
	}

	// cache the max count
	if count > dbMax {
		dbMax = count
		dbMaxT = time.Now()
		log.Printf("Cached total quotes at %v. Number of quotes: %d", dbMaxT, dbMax)
	}

	return count, nil
}
