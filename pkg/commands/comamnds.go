package commands

import (
	"fmt"
	"log"

	"github.com/bwmarrin/discordgo"
	"github.com/chocolatetilt/TiltbotGo/pkg/utils"
)

var (
	// Slice of all available commands
	Commands = []*discordgo.ApplicationCommand{
		{
			Name:        "addquote",
			Description: "Add a quote to the collection",
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
			Name:        "randomquote",
			Description: "Get a random quote from the collection",
		},
		{
			Name:        "latestquote",
			Description: "Get the latest quote from the collection",
		},
		{
			Name:        "countquote",
			Description: "Get the current number of quotes in the collection",
		},
	}
)

// Register commands to Discord in an overwrite fashion. For global commands, do not pass a guild flag.
func SetCommands() {
	log.Println("Adding commands...")
	_, err := utils.Session.ApplicationCommandBulkOverwrite(utils.Session.State.User.ID, utils.ReadConfig().GuildID, Commands)
	if err != nil {
		log.Panicf("Error in command creation: %v\n", err)
	} else {
		fmt.Println("All commands successfully registered (overwrite).")
	}
}
