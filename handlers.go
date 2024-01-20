package main

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/bwmarrin/discordgo"
)

const (
	QuoteTypeUser       = "user"
	QuoteTypeLatest     = "latest"
	QuoteTypeLatestUser = "latestUser"
	QuoteTypeRandom     = "rand"
)

var quoteHandler = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate, options []*discordgo.ApplicationCommandInteractionDataOption){
	"count": func(s *discordgo.Session, i *discordgo.InteractionCreate, options []*discordgo.ApplicationCommandInteractionDataOption) {
		ctx, cancel := ctxWithTimeout(10)
		defer cancel()

		count, err := quoteCount(ctx)
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
			Quote:     quote,
			Quotee:    fmt.Sprintf("<@%v>", quotee.ID),
			Quoter:    fmt.Sprintf("<@%v>", i.Member.User.ID),
			CreatedAt: t,
		}

		ctx, cancel := ctxWithTimeout(10)
		defer cancel()

		err = createQuote(ctx, quoteSave)
		if err != nil {
			sendErr(s, i, err)
			return
		}
		sendEmbed(s, i, fmt.Sprintf("Added quote for %s", quotee.Username), quoteFields(quoteSave))
	},
	"leaderboard": func(s *discordgo.Session, i *discordgo.InteractionCreate, options []*discordgo.ApplicationCommandInteractionDataOption) {
		// ctx, cancel := ctxWithTimeout(10)
		// defer cancel()

		// leaderboard, err := getLeaderboard(ctx)
		// if err != nil {
		// 	sendErr(s, i, err)
		// }
		// sendEmbed(s, i, "Quote Leaderboard", []*discordgo.MessageEmbedField{
		// 	{Name: "All-time", Value: leaderboard},
		// })
	},
	"user": func(s *discordgo.Session, i *discordgo.InteractionCreate, options []*discordgo.ApplicationCommandInteractionDataOption) {
		ctx, cancel := ctxWithTimeout(10)
		defer cancel()
		quote, err := getRandUserQuote(ctx, options[0].Options[0].UserValue(s).ID)
		if err != nil {
			sendErr(s, i, err)
			log.Printf("Error getting random user quote: %v", err)
			return
		}

		sendEmbed(s, i, fmt.Sprintf("Random Quote from %s", options[0].Options[0].UserValue(s).Username), quoteFields(quote))
	},
	"latest": func(s *discordgo.Session, i *discordgo.InteractionCreate, options []*discordgo.ApplicationCommandInteractionDataOption) {
		var quote Quote
		var err error
		ctx, cancel := ctxWithTimeout(10)
		defer cancel()

		// if the user is specified, get the latest quote for that user
		if len(options[0].Options) == 0 {
			quotee := options[0].Options[0].UserValue(s)
			quoteeID := fmt.Sprintf("<@%v>", quotee.ID)

			quote, err = getLatestUserQuote(ctx, quoteeID)
			if err == sql.ErrNoRows {
				sendMsg(s, i, fmt.Sprintf("No quotes found for %s", quotee.Username))
				return
			}
			if err != nil {
				sendErr(s, i, err)
				log.Printf("Error getting latest quote: %v", err)
				return
			}
		} else {
			quote, err = getLatestQuote(ctx)
			if err != nil {
				sendErr(s, i, err)
				log.Printf("Error getting latest quote: %v", err)
				return
			}
		}
		sendEmbed(s, i, "Latest Quote", quoteFields(quote))
	},
	"random": func(s *discordgo.Session, i *discordgo.InteractionCreate, options []*discordgo.ApplicationCommandInteractionDataOption) {
		ctx, cancel := ctxWithTimeout(10)
		defer cancel()

		quote, err := getRandQuote(ctx)
		if err != nil {
			sendErr(s, i, err)
			log.Printf("Error getting quote: %v", err)
			return
		}

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
