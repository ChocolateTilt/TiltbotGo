package main

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	QuoteTypeUser       = "user"
	QuoteTypeLatest     = "latest"
	QuoteTypeLatestUser = "latestUser"
	QuoteTypeRandom     = "rand"
)

var quoteHandler = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate, options []*discordgo.ApplicationCommandInteractionDataOption){
	"count": func(s *discordgo.Session, i *discordgo.InteractionCreate, options []*discordgo.ApplicationCommandInteractionDataOption) {
		ctx, cancel := ctxWithTimeout()
		defer cancel()

		count, err := quoteCount("", "full", ctx)
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
		t, err := discordgo.SnowflakeTimestamp(i.ID)
		if err != nil {
			sendErr(s, i, err)
			return
		}
		quoteSave := Quote{
			ID:        primitive.NewObjectID(),
			Quote:     quote,
			Quotee:    fmt.Sprintf("<@%v>", quotee.ID),
			Quoter:    fmt.Sprintf("<@%v>", i.Member.User.ID),
			CreatedAt: t,
		}

		ctx, cancel := ctxWithTimeout()
		defer cancel()

		err = createQuote(quoteSave, ctx)
		if err != nil {
			sendErr(s, i, err)
			return
		}
		sendEmbed(s, i, fmt.Sprintf("Added quote for %s", quotee.Username), quoteFields(quoteSave))
	},
	"leaderboard": func(s *discordgo.Session, i *discordgo.InteractionCreate, options []*discordgo.ApplicationCommandInteractionDataOption) {
		ctx, cancel := ctxWithTimeout()
		defer cancel()

		leaderboard, err := getLeaderboard(ctx)
		if err != nil {
			sendErr(s, i, err)
		}
		sendEmbed(s, i, "Quote Leaderboard", []*discordgo.MessageEmbedField{
			{Name: "All-time", Value: leaderboard},
		})
	},
	"user": func(s *discordgo.Session, i *discordgo.InteractionCreate, options []*discordgo.ApplicationCommandInteractionDataOption) {
		quote := handleQuoteSearch(s, i, options, QuoteTypeUser)

		sendEmbed(s, i, fmt.Sprintf("Random Quote from %s", options[0].Options[0].UserValue(s).Username), quoteFields(quote))
	},
	"latest": func(s *discordgo.Session, i *discordgo.InteractionCreate, options []*discordgo.ApplicationCommandInteractionDataOption) {
		var quote Quote

		if len(options[0].Options) == 0 {
			quote = handleQuoteSearch(s, i, options, QuoteTypeLatest)
		} else {
			quote = handleQuoteSearch(s, i, options, QuoteTypeLatestUser)
		}
		sendEmbed(s, i, "Latest Quote", quoteFields(quote))

	},
	"random": func(s *discordgo.Session, i *discordgo.InteractionCreate, options []*discordgo.ApplicationCommandInteractionDataOption) {
		quote := handleQuoteSearch(s, i, options, QuoteTypeRandom)

		sendEmbed(s, i, "Random Quote", quoteFields(quote))
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
