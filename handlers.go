package main

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/bwmarrin/discordgo"
)

type HandlerConext struct {
	Session *discordgo.Session
	DB      *SQLConn
}

var quoteHandler = map[string]func(hctx *HandlerConext, i *discordgo.InteractionCreate, options []*discordgo.ApplicationCommandInteractionDataOption){
	"count": func(hctx *HandlerConext, i *discordgo.InteractionCreate, options []*discordgo.ApplicationCommandInteractionDataOption) {
		ctx, cancel := ctxWithTimeout(10)
		defer cancel()

		count, err := hctx.DB.quoteCount(ctx)
		if err != nil {
			sendErr(hctx.Session, i, err)
			return
		}

		sendEmbed(hctx.Session, i, "Quote Count", []*discordgo.MessageEmbedField{
			{Name: "Total Quotes", Value: fmt.Sprintf("%d", count)},
		})
	},
	"add": func(hctx *HandlerConext, i *discordgo.InteractionCreate, options []*discordgo.ApplicationCommandInteractionDataOption) {
		quote := options[0].Options[0].StringValue()
		quotee := options[0].Options[1].UserValue(hctx.Session)
		t, err := discordgo.SnowflakeTimestamp(i.ID)
		if err != nil {
			sendErr(hctx.Session, i, err)
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

		err = hctx.DB.createQuote(ctx, quoteSave)
		if err != nil {
			sendErr(handlerCtx.Session, i, err)
			return
		}
		sendEmbed(handlerCtx.Session, i, fmt.Sprintf("Added quote for %s", quotee.Username), quoteFields(quoteSave))
	},
	"leaderboard": func(hctx *HandlerConext, i *discordgo.InteractionCreate, options []*discordgo.ApplicationCommandInteractionDataOption) {
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
	"user": func(hctx *HandlerConext, i *discordgo.InteractionCreate, options []*discordgo.ApplicationCommandInteractionDataOption) {
		ctx, cancel := ctxWithTimeout(10)
		defer cancel()
		quote, err := hctx.DB.getRandUserQuote(ctx, options[0].Options[0].UserValue(hctx.Session).ID)
		if err != nil {
			sendErr(hctx.Session, i, err)
			log.Printf("Error getting random user quote: %v", err)
			return
		}

		sendEmbed(hctx.Session, i, fmt.Sprintf("Random Quote from %s", options[0].Options[0].UserValue(hctx.Session).Username), quoteFields(quote))
	},
	"latest": func(hctx *HandlerConext, i *discordgo.InteractionCreate, options []*discordgo.ApplicationCommandInteractionDataOption) {
		var quote Quote
		var err error
		ctx, cancel := ctxWithTimeout(10)
		defer cancel()

		// if the user is specified, get the latest quote for that user
		if len(options[0].Options) == 0 {
			quotee := options[0].Options[0].UserValue(hctx.Session)
			quoteeID := fmt.Sprintf("<@%v>", quotee.ID)

			quote, err = hctx.DB.getLatestUserQuote(ctx, quoteeID)
			if err == sql.ErrNoRows {
				sendMsg(hctx.Session, i, fmt.Sprintf("No quotes found for %s", quotee.Username))
				return
			}
			if err != nil {
				sendErr(hctx.Session, i, err)
				log.Printf("Error getting latest quote: %v", err)
				return
			}
		} else {
			quote, err = hctx.DB.getLatestQuote(ctx)
			if err != nil {
				sendErr(hctx.Session, i, err)
				log.Printf("Error getting latest quote: %v", err)
				return
			}
		}
		sendEmbed(hctx.Session, i, "Latest Quote", quoteFields(quote))
	},
	"random": func(hctx *HandlerConext, i *discordgo.InteractionCreate, options []*discordgo.ApplicationCommandInteractionDataOption) {
		ctx, cancel := ctxWithTimeout(10)
		defer cancel()

		quote, err := hctx.DB.getRandQuote(ctx)
		if err != nil {
			sendErr(hctx.Session, i, err)
			log.Printf("Error getting quote: %v", err)
			return
		}

		sendEmbed(hctx.Session, i, "Random Quote", quoteFields(quote))
	},
}

// commandHandlers is the entrypoint for application commands and maps to commands and subcommands
var commandHandlers = map[string]func(hctx *HandlerConext, i *discordgo.InteractionCreate){
	"quote": func(hctx *HandlerConext, i *discordgo.InteractionCreate) {
		options := i.ApplicationCommandData().Options
		subCommand := options[0].Name

		h, ok := quoteHandler[subCommand]
		if !ok {
			sendErr(hctx.Session, i, fmt.Errorf("unknown sub-command: %s", subCommand))
			return
		}

		h(hctx, i, options)
	},
}
