package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/chocolatetilt/TiltbotGo/pkg/commands"
	"github.com/chocolatetilt/TiltbotGo/pkg/utils"
)

func init() {
	var err error
	utils.Session, err = discordgo.New("Bot " + utils.ReadConfig().Token)
	if err != nil {
		log.Fatalf("Invalid bot params: %v", err)
	}
	utils.ConnectMongo()
}

func main() {
	// add a handler for the "ready" event before opening the connection
	utils.Session.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		log.Printf("Logged in as: %v#%v", s.State.User.Username, s.State.User.Discriminator)
	})
	// open the websocket connection to Discord
	err := utils.Session.Open()
	if err != nil {
		log.Fatalf("Cannot open the Session: %v", err)
	}

	commands.SetCommands()

	// Add handler to catch Interactions (app command usage)
	utils.Session.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if h, ok := commands.CommandHandlers[i.ApplicationCommandData().Name]; ok {
			h(s, i)
		}
	})

	// close Session on Ctrl+C
	defer utils.Session.Close()
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	log.Println("Stop the container or press Ctrl+C to exit")
	<-stop
	log.Printf("Gracefully disconnected: %v#%v", utils.Session.State.User.Username, utils.Session.State.User.Discriminator)
}
