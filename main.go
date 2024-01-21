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
func setCommands(s *discordgo.Session) error {
	log.Println("Adding commands...")
	_, err := s.ApplicationCommandBulkOverwrite(s.State.User.ID, os.Getenv("DISCORD_GUILD"), commands)
	if err != nil {
		return fmt.Errorf("error in command creation: %w", err)
	}
	log.Println("All commands successfully registered (overwrite)")
	return nil
}

var handlerCtx *HandlerConext

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	db, err := newSQLConn()
	if err != nil {
		log.Fatalf("Cannot connect to the database: %v", err)
	}

	session, err := discordgo.New("Bot " + os.Getenv("DISCORD_TOKEN"))
	if err != nil {
		log.Fatalf("Cannot create a Discord session: %v", err)
	}

	handlerCtx = &HandlerConext{
		Session: session,
		DB:      db,
	}

	session.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		log.Printf("Logged in as: %v#%v", s.State.User.Username, s.State.User.Discriminator)
	})

	if err = session.Open(); err != nil {
		log.Fatalf("Cannot open the session: %v", err)
	}

	if err = setCommands(session); err != nil {
		log.Fatalln(err)
	}

	session.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if h, ok := commandHandlers[i.ApplicationCommandData().Name]; ok {
			h(handlerCtx, i)
		}
	})

	defer session.Close()
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	log.Println("Running version ", os.Getenv("VERSION"))
	log.Println("Stop the container or press Ctrl+C to exit")
	<-stop
	log.Printf("Gracefully disconnected: %v#%v", session.State.User.Username, session.State.User.Discriminator)
}
