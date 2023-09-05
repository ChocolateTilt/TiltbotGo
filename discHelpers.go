package main

import (
	"fmt"
	"os"
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

func sendErrToDiscord(s *discordgo.Session, i *discordgo.InteractionCreate, err error) {
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: fmt.Sprintf("Error executing command, please attempt it again. If this persists please contact <@%s> with the the error message.\nError message: %s",
				os.Getenv("DISC_BOT_OWNER_ID"), err),
			Flags: discordgo.MessageFlagsEphemeral,
		},
	})
}
