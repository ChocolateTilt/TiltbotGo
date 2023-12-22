package main

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	QuoteTypeUser   = "user"
	QuoteTypeLatest = "latest"
	QuoteTypeRandom = "rand"
)

var quoteHandler = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate, options []*discordgo.ApplicationCommandInteractionDataOption){
	"count": func(s *discordgo.Session, i *discordgo.InteractionCreate, options []*discordgo.ApplicationCommandInteractionDataOption) {
		var countType QuoteType = "full"
		count, err := countType.quoteCount("")
		if err != nil {
			sendErr(s, i, err)
			return
		}
		sendEmbed(s, i, "Quote Count", []*discordgo.MessageEmbedField{
			{Name: "Total Quotes", Value: fmt.Sprintf("%d", count)},
		})
	},
	"add": func(s *discordgo.Session, i *discordgo.InteractionCreate, options []*discordgo.ApplicationCommandInteractionDataOption) {
		quote := options[0].Options[0].StringValue()
		quotee := options[0].Options[1].UserValue(s)
		time, err := discordgo.SnowflakeTimestamp(i.ID)
		if err != nil {
			sendErr(s, i, err)
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
			sendErr(s, i, err)
			return
		}
	},
	"leaderboard": func(s *discordgo.Session, i *discordgo.InteractionCreate, options []*discordgo.ApplicationCommandInteractionDataOption) {
		leaderboard, err := getLeaderboard()
		if err != nil {
			sendErr(s, i, err)
		}
		sendEmbed(s, i, "Quote Leaderboard", []*discordgo.MessageEmbedField{
			{Name: "All-time", Value: leaderboard},
		})
	},
	"user": func(s *discordgo.Session, i *discordgo.InteractionCreate, options []*discordgo.ApplicationCommandInteractionDataOption) {
		handleQuoteSearch(s, i, options, QuoteTypeUser)
	},
	"latest": func(s *discordgo.Session, i *discordgo.InteractionCreate, options []*discordgo.ApplicationCommandInteractionDataOption) {
		handleQuoteSearch(s, i, options, QuoteTypeLatest)
	},
	"random": func(s *discordgo.Session, i *discordgo.InteractionCreate, options []*discordgo.ApplicationCommandInteractionDataOption) {
		handleQuoteSearch(s, i, options, QuoteTypeRandom)
	},
}

// commandHandlers is the entrypoint for application commands and maps to commands and subcommands
var commandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){

	"quote": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		options := i.ApplicationCommandData().Options
		subCommand := options[0].Name

		h, ok := quoteHandler[subCommand]
		if !ok {
			sendErr(s, i, fmt.Errorf("unknown sub-command: %s", subCommand))
			return
		}

		h(s, i, options)
	},
}
