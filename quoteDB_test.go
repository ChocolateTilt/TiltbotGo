package main

import (
	"context"
	"database/sql"
	"testing"
	"time"
)

// newTestDB creates an in-memory SQLite database with the quotes table.
func newTestDB(t *testing.T) *SQLConn {
	t.Helper()
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("open in-memory db: %v", err)
	}

	const table = "quotes"
	_, err = db.Exec(`CREATE TABLE quotes (
		id        INTEGER PRIMARY KEY AUTOINCREMENT,
		quote     TEXT    NOT NULL,
		quotee    TEXT    NOT NULL,
		quoter    TEXT    NOT NULL,
		createdAt TIMESTAMP NOT NULL
	)`)
	if err != nil {
		t.Fatalf("create table: %v", err)
	}

	conn := &SQLConn{Conn: db, Table: table, Cache: &QuoteCache{}}
	t.Cleanup(func() { db.Close() })
	return conn
}

func insertQuote(t *testing.T, conn *SQLConn, q Quote) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := conn.createQuote(ctx, q); err != nil {
		t.Fatalf("insertQuote: %v", err)
	}
}

func TestCreateAndCountQuotes(t *testing.T) {
	conn := newTestDB(t)
	ctx := context.Background()

	count, err := conn.quoteCount(ctx)
	if err != nil {
		t.Fatalf("quoteCount: %v", err)
	}
	if count != 0 {
		t.Errorf("expected 0 quotes, got %d", count)
	}

	insertQuote(t, conn, Quote{Quote: "hello", Quotee: "<@1>", Quoter: "<@2>", CreatedAt: time.Now()})

	// cache was reset on insert, so count should hit DB
	count, err = conn.quoteCount(ctx)
	if err != nil {
		t.Fatalf("quoteCount after insert: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1 quote, got %d", count)
	}
}

func TestQuoteCountCaching(t *testing.T) {
	conn := newTestDB(t)
	ctx := context.Background()

	insertQuote(t, conn, Quote{Quote: "cached", Quotee: "<@1>", Quoter: "<@2>", CreatedAt: time.Now()})

	// prime the cache
	if _, err := conn.quoteCount(ctx); err != nil {
		t.Fatalf("prime cache: %v", err)
	}

	// insert another quote without resetting cache (simulate stale cache)
	_, err := conn.Conn.Exec(`INSERT INTO quotes (quote, quotee, quoter, createdAt) VALUES (?, ?, ?, ?)`,
		"uncached", "<@1>", "<@2>", time.Now())
	if err != nil {
		t.Fatalf("raw insert: %v", err)
	}

	// count should return cached value (1), not 2
	count, err := conn.quoteCount(ctx)
	if err != nil {
		t.Fatalf("quoteCount from cache: %v", err)
	}
	if count != 1 {
		t.Errorf("expected cached count 1, got %d", count)
	}
}

func TestGetRandQuote(t *testing.T) {
	conn := newTestDB(t)
	ctx := context.Background()

	// empty table should return sql.ErrNoRows
	_, err := conn.getRandQuote(ctx)
	if err == nil {
		t.Fatal("expected error on empty table, got nil")
	}

	insertQuote(t, conn, Quote{Quote: "hi", Quotee: "<@1>", Quoter: "<@2>", CreatedAt: time.Now()})

	q, err := conn.getRandQuote(ctx)
	if err != nil {
		t.Fatalf("getRandQuote: %v", err)
	}
	if q.Quote != "hi" {
		t.Errorf("Quote = %q, want %q", q.Quote, "hi")
	}
}

func TestGetLatestQuote(t *testing.T) {
	conn := newTestDB(t)
	ctx := context.Background()

	insertQuote(t, conn, Quote{Quote: "first", Quotee: "<@1>", Quoter: "<@2>", CreatedAt: time.Now()})
	insertQuote(t, conn, Quote{Quote: "second", Quotee: "<@1>", Quoter: "<@2>", CreatedAt: time.Now()})

	q, err := conn.getLatestQuote(ctx)
	if err != nil {
		t.Fatalf("getLatestQuote: %v", err)
	}
	if q.Quote != "second" {
		t.Errorf("latest quote = %q, want %q", q.Quote, "second")
	}
}

func TestGetLatestUserQuote(t *testing.T) {
	conn := newTestDB(t)
	ctx := context.Background()

	insertQuote(t, conn, Quote{Quote: "user1 first", Quotee: "<@1>", Quoter: "<@2>", CreatedAt: time.Now()})
	insertQuote(t, conn, Quote{Quote: "user1 second", Quotee: "<@1>", Quoter: "<@2>", CreatedAt: time.Now()})
	insertQuote(t, conn, Quote{Quote: "user2 only", Quotee: "<@2>", Quoter: "<@1>", CreatedAt: time.Now()})

	q, err := conn.getLatestUserQuote(ctx, "1")
	if err != nil {
		t.Fatalf("getLatestUserQuote: %v", err)
	}
	if q.Quote != "user1 second" {
		t.Errorf("latest user quote = %q, want %q", q.Quote, "user1 second")
	}

	// unknown user should return sql.ErrNoRows
	_, err = conn.getLatestUserQuote(ctx, "999")
	if err == nil {
		t.Fatal("expected error for unknown user, got nil")
	}
}

func TestSearchQuote(t *testing.T) {
	conn := newTestDB(t)
	ctx := context.Background()

	insertQuote(t, conn, Quote{Quote: "hello world", Quotee: "<@1>", Quoter: "<@2>", CreatedAt: time.Now()})
	insertQuote(t, conn, Quote{Quote: "goodbye world", Quotee: "<@1>", Quoter: "<@2>", CreatedAt: time.Now()})
	insertQuote(t, conn, Quote{Quote: "nothing matches", Quotee: "<@1>", Quoter: "<@2>", CreatedAt: time.Now()})

	results, err := conn.searchQuote(ctx, "world")
	if err != nil {
		t.Fatalf("searchQuote: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("expected 2 results, got %d", len(results))
	}

	// no match
	results, err = conn.searchQuote(ctx, "zzznomatch")
	if err != nil {
		t.Fatalf("searchQuote no match: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("expected 0 results, got %d", len(results))
	}
}

func TestGetLeaderboard(t *testing.T) {
	conn := newTestDB(t)
	ctx := context.Background()

	insertQuote(t, conn, Quote{Quote: "a", Quotee: "<@1>", Quoter: "<@2>", CreatedAt: time.Now()})
	insertQuote(t, conn, Quote{Quote: "b", Quotee: "<@1>", Quoter: "<@2>", CreatedAt: time.Now()})
	insertQuote(t, conn, Quote{Quote: "c", Quotee: "<@2>", Quoter: "<@1>", CreatedAt: time.Now()})

	lb, err := conn.getLeaderboard(ctx)
	if err != nil {
		t.Fatalf("getLeaderboard: %v", err)
	}
	if lb == "" {
		t.Error("expected non-empty leaderboard")
	}
	// user 1 has 2 quotes and should appear first
	if lb[:5] != "`1:` " {
		t.Errorf("leaderboard doesn't start with position 1: %q", lb[:10])
	}
}
