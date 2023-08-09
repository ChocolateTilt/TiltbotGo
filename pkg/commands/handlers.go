package commands

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/chocolatetilt/TiltbotGo/pkg/utils"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var CommandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
	"quote": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		options := i.ApplicationCommandData().Options
		subCommand := options[0].Name

		switch subCommand {
		case "count":
			count := utils.QuoteCount("full", "")
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Embeds: []*discordgo.MessageEmbed{
						{
							Title: "Quote Count",
							Color: i.Member.User.AccentColor,
							Fields: []*discordgo.MessageEmbedField{
								{Name: "Total Quotes", Value: fmt.Sprintf("%v", count)},
							},
						},
					},
				},
			})
		case "add":
			quote := options[0].Options[0].StringValue()
			quotee := options[0].Options[1].UserValue(s)
			time, _ := discordgo.SnowflakeTimestamp(i.ID)
			quoteSave := utils.Quote{
				ID:        primitive.NewObjectID(),
				Quote:     quote,
				Quotee:    fmt.Sprintf("<@%v>", quotee.ID),
				Quoter:    fmt.Sprintf("<@%v>", i.Member.User.ID),
				CreatedAt: time,
			}
			utils.CreateQuote(quoteSave)
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Embeds: []*discordgo.MessageEmbed{
						{
							Title: "New Quote",
							Color: quotee.AccentColor,
							Fields: []*discordgo.MessageEmbedField{
								{Name: "Quote", Value: quoteSave.Quote},
								{Name: "Quotee", Value: quoteSave.Quotee},
								{Name: "Quoter", Value: quoteSave.Quoter},
								{Name: "Created At", Value: utils.DiscSnowflakeConvert(i.ID)},
							},
						},
					},
				},
			})
		case "random":
			quote := utils.GetQuote("rand", "")
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Embeds: []*discordgo.MessageEmbed{
						{
							Title:  "Random Quote",
							Color:  i.Member.User.AccentColor,
							Fields: utils.QuoteFields(quote),
						},
					},
				},
			})
		case "latest":
			quote := utils.GetQuote("latest", "")
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Embeds: []*discordgo.MessageEmbed{
						{
							Title:  "Latest Quote",
							Color:  i.Member.User.AccentColor,
							Fields: utils.QuoteFields(quote),
						},
					},
				},
			})
		case "user":
			quotee := options[0].Options[0].UserValue(s)
			userID := fmt.Sprintf("<@%v>", quotee.ID)
			quote := utils.GetQuote("user", userID)
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Embeds: []*discordgo.MessageEmbed{
						{
							Title:  "Random Quote",
							Color:  quotee.AccentColor,
							Fields: utils.QuoteFields(quote),
						},
					},
				},
			})
		case "leaderboard":
			leaderboard := utils.GetLeaderboard()
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
