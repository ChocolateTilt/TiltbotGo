package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
)

// setCommands registers commands to Discord in an overwrite fashion.
func setCommands(s *discordgo.Session) {
	discGuild := os.Getenv("DISCORD_GUILD")
	log.Println("Adding commands...")
	_, err := s.ApplicationCommandBulkOverwrite(s.State.User.ID, discGuild, commands)
	if err != nil {
		log.Printf("Error in command creation: %v\n", err)
	}
	fmt.Println("All commands successfully registered (overwrite).")
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	discToken := os.Getenv("DISCORD_TOKEN")

	err = connectMongo()
	if err != nil {
		log.Fatalf("Error connecting to MongoDB: %v", err)
	}

	session, err := discordgo.New("Bot " + discToken)
	if err != nil {
		log.Fatalf("Invalid bot params: %v", err)
	}

	session.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		log.Printf("Logged in as: %v#%v", s.State.User.Username, s.State.User.Discriminator)
	})

	err = session.Open()
	if err != nil {
		log.Fatalf("Cannot open the session: %v", err)
	}

	setCommands(session)

	session.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if h, ok := commandHandlers[i.ApplicationCommandData().Name]; ok {
			h(s, i)
		}
	})

	defer session.Close()
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	log.Println("Stop the container or press Ctrl+C to exit")
	<-stop
	log.Printf("Gracefully disconnected: %v#%v", session.State.User.Username, session.State.User.Discriminator)
}
