package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/bwmarrin/discordgo"
)

// quoteFields creates the embed fields for a quote
func quoteFields(q Quote) []*discordgo.MessageEmbedField {
	quoteTime := q.CreatedAt.Local().Format(time.RFC822)
	return []*discordgo.MessageEmbedField{
		{Name: "Quote", Value: q.Quote},
		{Name: "Quotee", Value: q.Quotee},
		{Name: "Quoter", Value: q.Quoter},
		{Name: "Created At", Value: quoteTime},
	}
}

// incidentFields creates the embed fields for an incident
func incidentFields(i Incident) []*discordgo.MessageEmbedField {
	incidentTime := i.CreatedAt.Local().Format(time.RFC822)
	return []*discordgo.MessageEmbedField{
		{Name: "Name", Value: i.Name},
		{Name: "Attendees", Value: i.Attendees},
		{Name: "Description", Value: i.Description},
		{Name: "Created At", Value: incidentTime},
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

// generateEmbed creates an embed with the passed in title and fields
func generateEmbed(t string, f []*discordgo.MessageEmbedField) *discordgo.MessageEmbed {
	return &discordgo.MessageEmbed{
		Title:  t,
		Color:  3093151, // dark blue
		Fields: f,
	}
}

// sendEmbed sends an embeded interaction response to the user who sent the command
func sendEmbed(s *discordgo.Session, i *discordgo.InteractionCreate, e []*discordgo.MessageEmbed) {
	// Respond to the interaction with the first embed
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: e,
		},
	})
}

// sendMsg sends a message to the user who sent the command
func sendMsg(s *discordgo.Session, i *discordgo.InteractionCreate, m string) {
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: m,
		},
	})
}

// ctxWithTimeout creates a context with s seconds timeout
func ctxWithTimeout(s time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), s*time.Second)
}
