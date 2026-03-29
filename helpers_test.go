package main

import (
	"testing"
	"time"
)

func TestQuoteFields(t *testing.T) {
	q := Quote{
		Quote:     "test quote",
		Quotee:    "<@123>",
		Quoter:    "<@456>",
		CreatedAt: time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC),
	}

	fields := quoteFields(q)

	if len(fields) != 4 {
		t.Fatalf("expected 4 fields, got %d", len(fields))
	}

	cases := []struct{ name, value string }{
		{"Quote", "test quote"},
		{"Quotee", "<@123>"},
		{"Quoter", "<@456>"},
	}
	for i, c := range cases {
		if fields[i].Name != c.name {
			t.Errorf("field[%d].Name = %q, want %q", i, fields[i].Name, c.name)
		}
		if fields[i].Value != c.value {
			t.Errorf("field[%d].Value = %q, want %q", i, fields[i].Value, c.value)
		}
	}
}

func TestGenerateEmbed(t *testing.T) {
	e := generateEmbed("Test Title", nil)

	if e.Title != "Test Title" {
		t.Errorf("Title = %q, want %q", e.Title, "Test Title")
	}
	if e.Color != embedColor {
		t.Errorf("Color = %d, want %d", e.Color, embedColor)
	}
}

func TestCtxWithTimeout(t *testing.T) {
	ctx, cancel := ctxWithTimeout()
	defer cancel()

	deadline, ok := ctx.Deadline()
	if !ok {
		t.Fatal("expected context to have a deadline")
	}

	remaining := time.Until(deadline)
	if remaining <= 0 || remaining > dbTimeout {
		t.Errorf("unexpected deadline remaining: %v (want 0 < remaining <= %v)", remaining, dbTimeout)
	}
}
