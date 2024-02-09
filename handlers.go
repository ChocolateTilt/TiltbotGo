package main

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/bwmarrin/discordgo"
)

type HandlerContext struct {
	Session *discordgo.Session
	DB      *SQLConn
}

var quoteHandler = map[string]func(c *HandlerContext, i *discordgo.InteractionCreate, o []*discordgo.ApplicationCommandInteractionDataOption){
	"count": func(c *HandlerContext, i *discordgo.InteractionCreate, o []*discordgo.ApplicationCommandInteractionDataOption) {
		ctx, cancel := ctxWithTimeout(10)
		defer cancel()

		count, err := c.DB.quoteCount(ctx)
		if err != nil {
			sendErr(c.Session, i, err)
			return
		}
		sendMsg(c.Session, i, fmt.Sprintf("There are %d quotes in the collection", count))
	},
	"add": func(c *HandlerContext, i *discordgo.InteractionCreate, o []*discordgo.ApplicationCommandInteractionDataOption) {
		quote := o[0].Options[0].StringValue()
		quotee := o[0].Options[1].UserValue(c.Session)
		t, err := discordgo.SnowflakeTimestamp(i.ID)
		if err != nil {
			sendErr(c.Session, i, err)
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

		err = c.DB.createQuote(ctx, quoteSave)
		if err != nil {
			sendErr(c.Session, i, err)
			return
		}
		e := generateEmbed("Added Quote", quoteFields(quoteSave))
		sendEmbed(c.Session, i, []*discordgo.MessageEmbed{e})
	},
	"leaderboard": func(c *HandlerContext, i *discordgo.InteractionCreate, o []*discordgo.ApplicationCommandInteractionDataOption) {
		ctx, cancel := ctxWithTimeout(10)
		defer cancel()

		leaderboard, err := c.DB.getLeaderboard(ctx)
		if err != nil {
			sendErr(c.Session, i, err)
		}

		e := generateEmbed("Quote Leaderboard", []*discordgo.MessageEmbedField{
			{Name: "All-time", Value: leaderboard},
		})
		sendEmbed(c.Session, i, []*discordgo.MessageEmbed{e})
	},
	"latest": func(c *HandlerContext, i *discordgo.InteractionCreate, options []*discordgo.ApplicationCommandInteractionDataOption) {
		var quote Quote
		var err error
		ctx, cancel := ctxWithTimeout(10)
		defer cancel()

		// if the user is specified, get the latest quote for that user
		if len(options[0].Options) != 0 {
			quotee := options[0].Options[0].UserValue(c.Session)

			quote, err = c.DB.getLatestUserQuote(ctx, quotee.ID)
			if err == sql.ErrNoRows {
				sendMsg(c.Session, i, fmt.Sprintf("No quotes found for %s", quotee.Username))
				return
			}
			if err != nil {
				sendErr(c.Session, i, err)
				log.Printf("Error getting latest user quote: %v", err)
				return
			}
		} else {
			quote, err = c.DB.getLatestQuote(ctx)
			if err != nil {
				sendErr(c.Session, i, err)
				log.Printf("Error getting latest quote: %v", err)
				return
			}
		}
		e := generateEmbed("Latest Quote", quoteFields(quote))
		sendEmbed(c.Session, i, []*discordgo.MessageEmbed{e})
	},
	"random": func(c *HandlerContext, i *discordgo.InteractionCreate, o []*discordgo.ApplicationCommandInteractionDataOption) {
		var quote Quote
		var err error
		ctx, cancel := ctxWithTimeout(10)
		defer cancel()

		// if the user is specified, get a random quote for that user
		if len(o[0].Options) != 0 {
			quotee := o[0].Options[0].UserValue(c.Session)

			quote, err = c.DB.getRandUserQuote(ctx, quotee.ID)
			if err == sql.ErrNoRows {
				sendMsg(c.Session, i, fmt.Sprintf("No quotes found for %s", quotee.Username))
				return
			}
			if err != nil {
				sendErr(c.Session, i, err)
				log.Printf("Error getting random user quote: %v", err)
				return
			}
		} else {
			quote, err = c.DB.getRandQuote(ctx)
			if err != nil {
				sendErr(c.Session, i, err)
				log.Printf("Error getting random quote: %v", err)
				return
			}
		}
		e := generateEmbed("Random Quote", quoteFields(quote))
		sendEmbed(c.Session, i, []*discordgo.MessageEmbed{e})
	},
	"search": func(c *HandlerContext, i *discordgo.InteractionCreate, options []*discordgo.ApplicationCommandInteractionDataOption) {
		ctx, cancel := ctxWithTimeout(10)
		defer cancel()

		searchTerm := options[0].Options[0].StringValue()
		quotes, err := c.DB.searchQuote(ctx, searchTerm)
		if err != nil {
			sendErr(c.Session, i, err)
			log.Printf("Error searching quotes: %v", err)
			return
		}

		var e []*discordgo.MessageEmbed
		for x, quote := range quotes {
			emb := generateEmbed(fmt.Sprintf("Search Result %d", x+1), quoteFields(quote))
			e = append(e, emb)
		}
		sendEmbed(c.Session, i, e)
	},
}

// commandHandlers is the entrypoint for application commands and maps to commands and subcommands
var commandHandlers = map[string]func(c *HandlerContext, i *discordgo.InteractionCreate){
	"quote": func(c *HandlerContext, i *discordgo.InteractionCreate) {
		o := i.ApplicationCommandData().Options
		subCommand := o[0].Name

		h, ok := quoteHandler[subCommand]
		if !ok {
			sendErr(c.Session, i, fmt.Errorf("unknown sub-command: %s", subCommand))
			return
		}
		h(c, i, o)
	},
}
