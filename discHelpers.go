package main

import (
	"time"

	"github.com/bwmarrin/discordgo"
)

// quoteFields returns a slice of MessageEmbedFields for a given Quote
func quoteFields(quote Quote) []*discordgo.MessageEmbedField {
	quoteTime := quote.CreatedAt.Local().Format(time.RFC822)
	return []*discordgo.MessageEmbedField{
		{Name: "Quote", Value: quote.Quote},
		{Name: "Quotee", Value: quote.Quotee},
		{Name: "Quoter", Value: quote.Quoter},
		{Name: "Created At", Value: quoteTime},
	}
}
