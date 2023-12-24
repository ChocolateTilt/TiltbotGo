package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/bwmarrin/discordgo"
)

// quoteFields creates the embed fields for a quote
func quoteFields(quote Quote) []*discordgo.MessageEmbedField {
	quoteTime := quote.CreatedAt.Local().Format(time.RFC822)
	return []*discordgo.MessageEmbedField{
		{Name: "Quote", Value: quote.Quote},
		{Name: "Quotee", Value: quote.Quotee},
		{Name: "Quoter", Value: quote.Quoter},
		{Name: "Created At", Value: quoteTime},
	}
}

// sendErr sends an ephemeral message to the user who sent the command with the error message
func sendErr(s *discordgo.Session, i *discordgo.InteractionCreate, err error) {
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: fmt.Sprintf("Error executing command, please attempt it again. If this persists please contact <@%s> with the the error message.\nError message: %s",
				os.Getenv("DISC_BOT_OWNER_ID"), err),
			Flags: discordgo.MessageFlagsEphemeral,
		},
	})
}

// sendEmbed sends an InteractionRespond to the passed in session/interaction
func sendEmbed(s *discordgo.Session, i *discordgo.InteractionCreate, title string, fields []*discordgo.MessageEmbedField) {
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{
				{
					Title:  title,
					Color:  3093151, // dark blue
					Fields: fields,
				},
			},
		},
	})
}

// handleQuoteSearch is a helper function that gets a quote based on t
func handleQuoteSearch(s *discordgo.Session, i *discordgo.InteractionCreate, opts []*discordgo.ApplicationCommandInteractionDataOption, t QuoteType) Quote {
	var quoteeID string
	if t == QuoteTypeUser {
		quotee := opts[0].Options[0].UserValue(s)
		quoteeID = fmt.Sprintf("<@%v>", quotee.ID)
	}

	ctx, cancel := ctxWithTimeout()
	defer cancel()

	quote, err := t.getQuote(quoteeID, ctx)
	if err != nil {
		sendErr(s, i, err)
		log.Printf("Error getting quote: %v", err) // Log the error
	}
	return quote
}

// ctxWithTimeout creates a context with a ten second timeout
func ctxWithTimeout() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 10*time.Second)
}
