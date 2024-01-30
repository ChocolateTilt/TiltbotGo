package main

import (
	"github.com/bwmarrin/discordgo"
)

var (
	// Slice of all application commands
	commands = []*discordgo.ApplicationCommand{
		{
			Name:        "quote",
			Description: "Commands for interacting with the collection of quotes",
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
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionUser,
							Name:        "user",
							Description: "Get a random quote for a specific user",
							Required:    false,
						},
					},
				},
				{
					Name:        "latest",
					Description: "Get the most recent quote from the collection",
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionUser,
							Name:        "user",
							Description: "Get the most recent quote for a specific user",
							Required:    false,
						},
					},
				},
				{
					Name:        "count",
					Description: "Get the current number of quotes in the collection",
					Type:        discordgo.ApplicationCommandOptionSubCommand,
				},
				{
					Name:        "leaderboard",
					Description: "Get the leaderboard of users with the most quotes",
					Type:        discordgo.ApplicationCommandOptionSubCommand,
				},
				{
					Name:        "search",
					Description: "Search the collection of quotes for a specific string",
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "query",
							Description: "Get a random quote for a specific user",
							Required:    true,
						},
					},
				},
			},
		},
		{
			Name:        "incident",
			Description: "Commands for interacting with the collection of incidents",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:        "add",
					Description: "Add an incident to the collection",
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "name",
							Description: "Name of the incident",
							Required:    true,
						},
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "attendees",
							Description: "Attendees of the incident",
							Required:    true,
						},
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "description",
							Description: "Description of the incident",
							Required:    true,
						},
					},
				},
				{
					Name:        "random",
					Description: "Get a random incident from the collection",
					Type:        discordgo.ApplicationCommandOptionSubCommand,
				},
				{
					Name:        "count",
					Description: "Get the current number of incidents in the collection",
					Type:        discordgo.ApplicationCommandOptionSubCommand,
				},
			},
		},
	}
)
