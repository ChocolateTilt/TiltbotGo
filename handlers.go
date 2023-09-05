package main

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// commandHandlers is a map of all available Discord slash command handlers
var commandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
	"quote": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		options := i.ApplicationCommandData().Options
		subCommand := options[0].Name
		var searchType QuoteType

		switch subCommand {
		case "count":
			var countType QuoteType = "full"
			count, err := countType.quoteCount("")
			if err != nil {
				sendErrToDiscord(s, i, err)
				return
			}

			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Embeds: []*discordgo.MessageEmbed{
						{
							Title: "Quote Count",
							Color: 3093151, // dark blue
							Fields: []*discordgo.MessageEmbedField{
								{Name: "Total Quotes", Value: fmt.Sprintf("%d", count)},
							},
						},
					},
				},
			})
		case "add":
			quote := options[0].Options[0].StringValue()
			quotee := options[0].Options[1].UserValue(s)
			time, err := discordgo.SnowflakeTimestamp(i.ID)
			if err != nil {
				sendErrToDiscord(s, i, err)
				return
			}
			quoteSave := Quote{
				ID:        primitive.NewObjectID(),
				Quote:     quote,
				Quotee:    fmt.Sprintf("<@%v>", quotee.ID),
				Quoter:    fmt.Sprintf("<@%v>", i.Member.User.ID),
				CreatedAt: time,
			}
			err = createQuote(quoteSave)
			if err != nil {
				sendErrToDiscord(s, i, err)
				return
			}

			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Embeds: []*discordgo.MessageEmbed{
						{
							Title:  "New Quote",
							Color:  3093151, // dark blue
							Fields: quoteFields(quoteSave),
						},
					},
				},
			})
		case "random":
			searchType = "rand"
			quote, err := searchType.getQuote("")
			if err != nil {
				sendErrToDiscord(s, i, err)
			}

			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Embeds: []*discordgo.MessageEmbed{
						{
							Title:  "Random Quote",
							Color:  3093151, // dark blue
							Fields: quoteFields(quote),
						},
					},
				},
			})
		case "latest":
			searchType = "latest"
			quote, err := searchType.getQuote("")
			if err != nil {
				sendErrToDiscord(s, i, err)
			}

			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Embeds: []*discordgo.MessageEmbed{
						{
							Title:  "Latest Quote",
							Color:  3093151, // dark blue
							Fields: quoteFields(quote),
						},
					},
				},
			})
		case "user":
			searchType = "user"
			quotee := options[0].Options[0].UserValue(s)
			userID := fmt.Sprintf("<@%v>", quotee.ID)
			quote, err := searchType.getQuote(userID)
			if err != nil {
				sendErrToDiscord(s, i, err)
			}

			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Embeds: []*discordgo.MessageEmbed{
						{
							Title:  "Random Quote",
							Color:  3093151, // dark blue
							Fields: quoteFields(quote),
						},
					},
				},
			})
		case "leaderboard":
			leaderboard, err := getLeaderboard()
			if err != nil {
				sendErrToDiscord(s, i, err)
			}

			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Embeds: []*discordgo.MessageEmbed{
						{
							Title: "Quote Leaderboard",
							Color: 3093151, // dark blue
							Fields: []*discordgo.MessageEmbedField{
								{Name: "All-time", Value: leaderboard},
							},
						},
					},
				},
			})
		}
	},
}
