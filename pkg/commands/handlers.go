package commands

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/chocolatetilt/TiltbotGo/pkg/utils"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var CommandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
	"quote": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		options := i.ApplicationCommandData().Options
		subCommand := options[0].Name
		switch subCommand {
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
			quote := utils.GetQuote("user", quotee.ID)
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
			// 		case "count":
			// 			count := utils.QuoteCount("full", "")
			// 			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			// 				Type: discordgo.InteractionResponseChannelMessageWithSource,
			// 				Data: &discordgo.InteractionResponseData{
			// 					Content: fmt.Sprintf("There are %v quotes in the collection.", count),
			// 		},
			// 	}),
			// }
		}
	},
}
