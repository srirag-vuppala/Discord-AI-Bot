package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	wolframgo "github.com/Krognol/go-wolfram"
	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
	"github.com/tidwall/gjson"
	witai "github.com/wit-ai/wit-go/v2"
)

func main() {
	// Load environment variables
	err := godotenv.Load()
	if err != nil {
		log.Fatal(fmt.Sprintf("Error loading the .env file %s",err))
	}
	
	// Create a Discord session
	dg, err := discordgo.New("Bot "+ os.Getenv("DISCORD_CLIENT_TOKEN"))
	if err != nil {
		log.Fatal(fmt.Sprintf("Error creating the Discord session %s",err))
	}
	// Register the messageCreate func as a callback for MessageCreate events.
	dg.AddHandler(messageCreate)

	// In this example, we only care about receiving message events.
	dg.Identify.Intents = discordgo.IntentsGuildMessages

	// Open a websocket connection to Discord and begin listening.
	err = dg.Open()
	if err != nil {
		log.Fatal(fmt.Sprintf("Error in opening websocket connection %s",err))
	}

	// Wait here until CTRL-C or other term signal is received.
	fmt.Println("Bot is now running. Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	// Cleanly close down the Discord session.
	dg.Close()
}

// This function will be called (due to AddHandler above) every time a new
// message is created on any channel that the authenticated bot has access to.
func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {

	witClient := witai.NewClient(os.Getenv("WIT_AI_TOKEN"))
	wolframClient := wolframgo.Client{os.Getenv("WOLFRAM_TOKEN")}
	

    // Ignore all messages created by the bot itself
    if m.Author.ID == s.State.User.ID {
        return
    }

	if m.Content == "" {
		return
	}

	splitQuery := strings.Split(m.Content, " ")

	switch splitQuery[0] {
		case "!question": 
			query := strings.Join(splitQuery[1:], " ")

			msg, _ := witClient.Parse(&witai.MessageRequest{
				Query: query,
			})
			data, _ := json.MarshalIndent(msg, "", "    ")
			rough := string(data[:])
			value := gjson.Get(rough, "entities.wit$wolfram_search_query:wolfram_search_query.0.value")
			answer := value.String()
			res, err := wolframClient.GetSpokentAnswerQuery(answer, wolframgo.Metric, 1000)
			if err != nil {
				fmt.Println("there is an error")
			}
			fmt.Println(value)
			s.ChannelMessageSend(m.ChannelID, res)

	}
}

