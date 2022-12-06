package commands

import (
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/chocolatetilt/TiltbotGo/pkg/utils"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func discSnowflakeConvert(i string) string {
	timeStamp, err := discordgo.SnowflakeTimestamp(i)
	if err != nil {
		fmt.Print(err)
	}
	return timeStamp.Local().Format(time.RFC822)
}

var CommandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
	"addquote": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		options := i.ApplicationCommandData().Options
		quote := options[0].StringValue()
		quotee := options[1].UserValue(s)
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
							{Name: "Created At", Value: discSnowflakeConvert(i.ID)},
						},
					},
				},
			},
		})

	},
	"randomquote": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		doc := utils.GetQuote("rand", "")
		quoteTime := doc.CreatedAt.Local().Format(time.RFC822)

		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Embeds: []*discordgo.MessageEmbed{
					{
						Title: "Random Quote",
						Color: i.Member.User.AccentColor,
						Fields: []*discordgo.MessageEmbedField{
							{Name: "Quote", Value: doc.Quote},
							{Name: "Quotee", Value: doc.Quotee},
							{Name: "Quoter", Value: doc.Quoter},
							{Name: "Created At", Value: quoteTime},
						},
					},
				},
			},
		})
	},
	"latestquote": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		doc := utils.GetQuote("latest", "")
		quoteTime := doc.CreatedAt.Local().Format(time.RFC822)

		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Embeds: []*discordgo.MessageEmbed{
					{
						Title: "Latest Quote",
						Color: i.Member.User.AccentColor,
						Fields: []*discordgo.MessageEmbedField{
							{Name: "Quote", Value: doc.Quote},
							{Name: "Quotee", Value: doc.Quotee},
							{Name: "Quoter", Value: doc.Quoter},
							{Name: "Created At", Value: quoteTime},
						},
					},
				},
			},
		})
	},
	"countquote": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		count := utils.QuoteCount("full", "")

		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("There are %v quotes in the collection.", count),
			},
		})
	},
	"userquote": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		options := i.ApplicationCommandData().Options[0].UserValue(s)
		userID := fmt.Sprintf("<@%v>", options.ID)
		doc := utils.GetQuote("user", userID)
		quoteTime := doc.CreatedAt.Local().Format(time.RFC822)

		if doc.Quote != "" {
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Embeds: []*discordgo.MessageEmbed{
						{
							Title: "Quote",
							Color: i.Member.User.AccentColor,
							Fields: []*discordgo.MessageEmbedField{
								{Name: "Quote", Value: doc.Quote},
								{Name: "Quotee", Value: doc.Quotee},
								{Name: "Quoter", Value: doc.Quoter},
								{Name: "Created At", Value: quoteTime},
							},
						},
					},
				},
			})
		} else {
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{Content: "No quotes exist for this user."},
			})
		}

	},
}
