package commands

import (
	"fmt"
	"log"

	"github.com/bwmarrin/discordgo"
	"github.com/chocolatetilt/TiltbotGo/pkg/utils"
)

var (
	// Slice of all application commands
	Commands = []*discordgo.ApplicationCommand{
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
			},
		},
	}
)

// Register commands to Discord in an overwrite fashion. For global commands, do not pass a guild flag.
func SetCommands() {
	log.Println("Adding commands...")
	_, err := utils.Session.ApplicationCommandBulkOverwrite(utils.Session.State.User.ID, utils.Conf.GuildID, Commands)
	if err != nil {
		log.Panicf("Error in command creation: %v\n", err)
	}
	fmt.Println("All commands successfully registered (overwrite).")
}
