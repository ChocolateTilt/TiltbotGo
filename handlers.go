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
		sendMsg(hctx.Session, i, fmt.Sprintf("There are %d quotes in the collection", count))
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
		e := generateEmbed("Added Quote", quoteFields(quoteSave))
		sendEmbed(handlerCtx.Session, i, []*discordgo.MessageEmbed{e})
	},
	"leaderboard": func(hctx *HandlerConext, i *discordgo.InteractionCreate, options []*discordgo.ApplicationCommandInteractionDataOption) {
		ctx, cancel := ctxWithTimeout(10)
		defer cancel()

		leaderboard, err := hctx.DB.getLeaderboard(ctx)
		if err != nil {
			sendErr(handlerCtx.Session, i, err)
		}

		e := generateEmbed("Quote Leaderboard", []*discordgo.MessageEmbedField{
			{Name: "All-time", Value: leaderboard},
		})
		sendEmbed(handlerCtx.Session, i, []*discordgo.MessageEmbed{e})
	},
	"latest": func(hctx *HandlerConext, i *discordgo.InteractionCreate, options []*discordgo.ApplicationCommandInteractionDataOption) {
		var quote Quote
		var err error
		ctx, cancel := ctxWithTimeout(10)
		defer cancel()

		// if the user is specified, get the latest quote for that user
		if len(options[0].Options) != 0 {
			quotee := options[0].Options[0].UserValue(hctx.Session)

			quote, err = hctx.DB.getLatestUserQuote(ctx, quotee.ID)
			if err == sql.ErrNoRows {
				sendMsg(hctx.Session, i, fmt.Sprintf("No quotes found for %s", quotee.Username))
				return
			}
			if err != nil {
				sendErr(hctx.Session, i, err)
				log.Printf("Error getting latest user quote: %v", err)
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
		e := generateEmbed("Latest Quote", quoteFields(quote))
		sendEmbed(handlerCtx.Session, i, []*discordgo.MessageEmbed{e})
	},
	"random": func(hctx *HandlerConext, i *discordgo.InteractionCreate, options []*discordgo.ApplicationCommandInteractionDataOption) {
		var quote Quote
		var err error
		ctx, cancel := ctxWithTimeout(10)
		defer cancel()

		// if the user is specified, get a random quote for that user
		if len(options[0].Options) != 0 {
			quotee := options[0].Options[0].UserValue(hctx.Session)

			quote, err = hctx.DB.getRandUserQuote(ctx, quotee.ID)
			if err == sql.ErrNoRows {
				sendMsg(hctx.Session, i, fmt.Sprintf("No quotes found for %s", quotee.Username))
				return
			}
			if err != nil {
				sendErr(hctx.Session, i, err)
				log.Printf("Error getting random user quote: %v", err)
				return
			}
		} else {
			quote, err = hctx.DB.getRandQuote(ctx)
			if err != nil {
				sendErr(hctx.Session, i, err)
				log.Printf("Error getting random quote: %v", err)
				return
			}
		}
		e := generateEmbed("Random Quote", quoteFields(quote))
		sendEmbed(handlerCtx.Session, i, []*discordgo.MessageEmbed{e})
	},
	"search": func(hctx *HandlerConext, i *discordgo.InteractionCreate, options []*discordgo.ApplicationCommandInteractionDataOption) {
		ctx, cancel := ctxWithTimeout(10)
		defer cancel()

		searchTerm := options[0].Options[0].StringValue()
		quotes, err := hctx.DB.searchQuote(ctx, searchTerm)
		if err != nil {
			sendErr(hctx.Session, i, err)
			log.Printf("Error searching quotes: %v", err)
			return
		}

		var e []*discordgo.MessageEmbed
		for x, quote := range quotes {
			emb := generateEmbed(fmt.Sprintf("Search Result %d", x), quoteFields(quote))
			e = append(e, emb)
		}
		sendEmbed(handlerCtx.Session, i, e)
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
