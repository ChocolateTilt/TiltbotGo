package utils

import (
	"encoding/json"
	"log"
	"os"

	"github.com/bwmarrin/discordgo"
)

type Config struct {
	Token      string `json:"token"`
	GuildID    string `json:"guild"`
	MongoURI   string `json:"mongoURI"`
	Collection string `json:"collection"`
}

var Session *discordgo.Session
var Conf = ReadConfig()

func ReadConfig() Config {
	var config Config
	configFile, err := os.Open("conf.json")

	if err != nil {
		log.Fatalf("Error in utils.ReadConfig(): %v", err)
	}

	defer configFile.Close()
	jsonParse := json.NewDecoder(configFile)
	jsonParse.Decode(&config)
	return config
}
