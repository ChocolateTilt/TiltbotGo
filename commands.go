package main

import (
	"github.com/bwmarrin/discordgo"
)

var (
	// Slice of all application commands
	commands = []*discordgo.ApplicationCommand{
		{
			Name:        "quote",
			Description: "Commands for interacting with the quote collection",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:        "add",
					Description: "Add a quote to the collection",
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "quote",
							Description: "Quote to add",
							Required:    true,
						},
						{
							Type:        discordgo.ApplicationCommandOptionUser,
							Name:        "quotee",
							Description: "Person who spoke the cursed quote",
							Required:    true,
						},
					},
				},
				{
					Name:        "random",
					Description: "Get a random quote from the collection",
					Type:        discordgo.ApplicationCommandOptionSubCommand,
				},
				{
					Name:        "latest",
					Description: "Get the latest quote from the collection",
					Type:        discordgo.ApplicationCommandOptionSubCommand,
				},
				{
					Name:        "count",
					Description: "Get the current number of quotes in the collection",
					Type:        discordgo.ApplicationCommandOptionSubCommand,
				},
				{
					Name:        "user",
					Description: "Get a random quote from a specific user",
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionUser,
							Name:        "user",
							Description: "The user to search for",
							Required:    true,
						},
					},
				},
				{
					Name:        "leaderboard",
					Description: "Get the leaderboard of users with the most quotes",
					Type:        discordgo.ApplicationCommandOptionSubCommand,
				},
			},
		},
	}
)
