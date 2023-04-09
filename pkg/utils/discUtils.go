package utils

import (
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
)

func DiscSnowflakeConvert(i string) string {
	timeStamp, err := discordgo.SnowflakeTimestamp(i)
	if err != nil {
		fmt.Print(err)
	}
	return timeStamp.Local().Format(time.RFC822)
}

func QuoteFields(quote Quote) []*discordgo.MessageEmbedField {
	quoteTime := quote.CreatedAt.Local().Format(time.RFC822)
	return []*discordgo.MessageEmbedField{
		{Name: "Quote", Value: quote.Quote},
		{Name: "Quotee", Value: quote.Quotee},
		{Name: "Quoter", Value: quote.Quoter},
		{Name: "Created At", Value: quoteTime},
	}
}
