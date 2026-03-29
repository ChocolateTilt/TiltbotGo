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

// validateEnv checks that all required environment variables are set and fatals if any are missing.
func validateEnv() {
	required := []string{"DISCORD_TOKEN", "DISCORD_GUILD", "DISC_BOT_OWNER_ID", "SQLITE_DB", "SQLITE_TABLE_NAME"}
	for _, key := range required {
		if os.Getenv(key) == "" {
			log.Fatalf("Required environment variable %s is not set", key)
		}
	}
}

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

var handlerCtx *HandlerContext

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	validateEnv()

	db, err := newSQLConn()
	if err != nil {
		log.Fatalf("Cannot connect to the database: %v", err)
	}
	defer db.Conn.Close()

	session, err := discordgo.New("Bot " + os.Getenv("DISCORD_TOKEN"))
	if err != nil {
		log.Fatalf("Cannot create a Discord session: %v", err)
	}

	handlerCtx = &HandlerContext{
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
