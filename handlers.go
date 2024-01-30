package main

import (
	"database/sql"
	"fmt"
	"log"
	"strings"

	"github.com/bwmarrin/discordgo"
)

var quoteHandler = map[string]func(c *HandlerConext, i *discordgo.InteractionCreate, options []*discordgo.ApplicationCommandInteractionDataOption){
	"count": func(c *HandlerConext, i *discordgo.InteractionCreate, options []*discordgo.ApplicationCommandInteractionDataOption) {
		ctx, cancel := ctxWithTimeout(10)
		defer cancel()

		count, err := c.DB.countQuotes(ctx)
		if err != nil {
			sendErr(c.Session, i, err)
			return
		}
		sendMsg(c.Session, i, fmt.Sprintf("There are %d quotes in the collection", count))
	},
	"add": func(c *HandlerConext, i *discordgo.InteractionCreate, options []*discordgo.ApplicationCommandInteractionDataOption) {
		quote := options[0].Options[0].StringValue()
		quotee := options[0].Options[1].UserValue(c.Session)
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
	"leaderboard": func(c *HandlerConext, i *discordgo.InteractionCreate, options []*discordgo.ApplicationCommandInteractionDataOption) {
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
	"latest": func(c *HandlerConext, i *discordgo.InteractionCreate, options []*discordgo.ApplicationCommandInteractionDataOption) {
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
	"random": func(c *HandlerConext, i *discordgo.InteractionCreate, options []*discordgo.ApplicationCommandInteractionDataOption) {
		var quote Quote
		var err error
		ctx, cancel := ctxWithTimeout(10)
		defer cancel()

		// if the user is specified, get a random quote for that user
		if len(options[0].Options) != 0 {
			quotee := options[0].Options[0].UserValue(c.Session)

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
	"search": func(c *HandlerConext, i *discordgo.InteractionCreate, options []*discordgo.ApplicationCommandInteractionDataOption) {
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

var incidentHandler = map[string]func(c *HandlerConext, i *discordgo.InteractionCreate, options []*discordgo.ApplicationCommandInteractionDataOption){
	"add": func(c *HandlerConext, i *discordgo.InteractionCreate, options []*discordgo.ApplicationCommandInteractionDataOption) {
		users := strings.Split(options[0].Options[1].StringValue(), ",")
		var fUser []string

		for _, u := range users {
			fUser = append(fUser, fmt.Sprintf("<@%s>", u))
		}

		incident := Incident{
			Name:        options[0].Options[0].StringValue(),
			Attendees:   fUser,
			Description: options[0].Options[3].StringValue(),
		}

		ctx, cancel := ctxWithTimeout(10)
		defer cancel()

		err := c.DB.createIncident(ctx, incident)
		if err != nil {
			sendErr(c.Session, i, err)
			return
		}
		e := generateEmbed("Added Incident", incidentFields(incident))
		sendEmbed(c.Session, i, []*discordgo.MessageEmbed{e})
	},

	"random": func(c *HandlerConext, i *discordgo.InteractionCreate, options []*discordgo.ApplicationCommandInteractionDataOption) {
		ctx, cancel := ctxWithTimeout(10)
		defer cancel()

		incident, err := c.DB.getRandIncident(ctx)
		if err != nil {
			sendErr(c.Session, i, err)
			return
		}
		e := generateEmbed("Random Incident", incidentFields(incident))
		sendEmbed(c.Session, i, []*discordgo.MessageEmbed{e})
	},

	"count": func(c *HandlerConext, i *discordgo.InteractionCreate, options []*discordgo.ApplicationCommandInteractionDataOption) {
		ctx, cancel := ctxWithTimeout(10)
		defer cancel()

		count, err := c.DB.countIncidents(ctx)
		if err != nil {
			sendErr(c.Session, i, err)
			return
		}
		sendMsg(c.Session, i, fmt.Sprintf("There are %d incidents in the collection", count))
	},
}

// commandHandlers is the entrypoint for application commands and maps to commands and subcommands
var commandHandlers = map[string]func(c *HandlerConext, i *discordgo.InteractionCreate){
	"quote": func(c *HandlerConext, i *discordgo.InteractionCreate) {
		options := i.ApplicationCommandData().Options
		subCommand := options[0].Name

		h, ok := quoteHandler[subCommand]
		if !ok {
			sendErr(c.Session, i, fmt.Errorf("unknown sub-command: %s", subCommand))
			return
		}

		h(c, i, options)
	},
	"incident": func(c *HandlerConext, i *discordgo.InteractionCreate) {
		options := i.ApplicationCommandData().Options
		subCommand := options[0].Name

		h, ok := incidentHandler[subCommand]
		if !ok {
			sendErr(c.Session, i, fmt.Errorf("unknown sub-command: %s", subCommand))
			return
		}

		h(c, i, options)
	},
}
