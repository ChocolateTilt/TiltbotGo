package main

import (
	"fmt"
	"log"
	"strings"

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
			count := countType.quoteCount("")
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Embeds: []*discordgo.MessageEmbed{
						{
							Title: "Quote Count",
							Color: i.Member.User.AccentColor,
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
			time, _ := discordgo.SnowflakeTimestamp(i.ID)
			quoteSave := Quote{
				ID:        primitive.NewObjectID(),
				Quote:     quote,
				Quotee:    fmt.Sprintf("<@%v>", quotee.ID),
				Quoter:    fmt.Sprintf("<@%v>", i.Member.User.ID),
				CreatedAt: time,
			}
			err := createQuote(quoteSave)
			if err != nil {
				log.Println(err)
				return
			}
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Embeds: []*discordgo.MessageEmbed{
						{
							Title:  "New Quote",
							Color:  quotee.AccentColor,
							Fields: quoteFields(quoteSave),
						},
					},
				},
			})
		case "random":
			searchType = "rand"
			quote := searchType.getQuote("")
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Embeds: []*discordgo.MessageEmbed{
						{
							Title:  "Random Quote",
							Color:  i.Member.User.AccentColor,
							Fields: quoteFields(quote),
						},
					},
				},
			})
		case "latest":
			searchType = "latest"
			quote := searchType.getQuote("")
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Embeds: []*discordgo.MessageEmbed{
						{
							Title:  "Latest Quote",
							Color:  i.Member.User.AccentColor,
							Fields: quoteFields(quote),
						},
					},
				},
			})
		case "user":
			searchType = "user"
			quotee := options[0].Options[0].UserValue(s)
			userID := fmt.Sprintf("<@%v>", quotee.ID)
			quote := searchType.getQuote(userID)
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Embeds: []*discordgo.MessageEmbed{
						{
							Title:  "Random Quote",
							Color:  quotee.AccentColor,
							Fields: quoteFields(quote),
						},
					},
				},
			})
		case "leaderboard":
			leaderboard := getLeaderboard()
			var leaderboardVal []string

			for i, v := range leaderboard {
				leaderboardVal = append(leaderboardVal, fmt.Sprintf("`%v:`%v: %v\n", i+1, v["_id"], v["count"]))
			}

			cleanLB := strings.Join(leaderboardVal, "\n")

			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Embeds: []*discordgo.MessageEmbed{
						{
							Title: "Quote Leaderboard",
							Color: 3093151, // dark blue
							Fields: []*discordgo.MessageEmbedField{
								{Name: "All-time", Value: cleanLB},
							},
						},
					},
				},
			})
		}
	},
}
