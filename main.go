package main

import (
	"log"
	"os"

	"github.com/akhilsharma90/go-openai-bot-discord/pkg/bot"
	"github.com/akhilsharma90/go-openai-bot-discord/pkg/commands"
	"github.com/akhilsharma90/go-openai-bot-discord/pkg/commands/gpt"
	"github.com/akhilsharma90/go-openai-bot-discord/pkg/constants"
	"github.com/sashabaranov/go-openai"
	"gopkg.in/yaml.v2"
)

type Config struct {
	Discord struct {
		Token          string `yaml:"token"`
		Guild          string `yaml:"guild"`
		RemoveCommands bool   `yaml:"removeCommands"`
	} `yaml:"discord"`
	OpenAI struct {
		APIKey           string   `yaml:"apiKey"`
		CompletionModels []string `yaml:"completionModels"`
	} `yaml:"openAI"`
}

func (c *Config) ReadFromFile(file string) error {
	data, err := os.ReadFile(file)
	if err != nil {
		return err
	}
	err = yaml.Unmarshal(data, c)
	if err != nil {
		return err
	}
	return nil
}

func init() {
	log.SetOutput(os.Stdout)
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

var (
	discordBot   *bot.Bot
	openaiClient *openai.Client

	gptMessagesCache     *gpt.MessagesCache
	ignoredChannelsCache = make(gpt.IgnoredChannelsCache)
)

func main() {
	// Read config from file
	config := &Config{}
	err := config.ReadFromFile("credentials.yaml")
	if err != nil {
		log.Fatalf("Error reading credentials.yaml: %v", err)
	}

	// Initialize cache
	gptMessagesCache, err = gpt.NewMessagesCache(constants.DiscordThreadsCacheSize)
	if err != nil {
		log.Fatalf("Error initializing GPTMessagesCache: %v", err)
	}

	// Initialize discord bot
	discordBot, err = bot.NewBot(config.Discord.Token)
	if err != nil {
		log.Fatalf("Invalid bot parameters: %v", err)
	}

	// Register commands
	if config.OpenAI.APIKey != "" {
		openaiClient = openai.NewClient(config.OpenAI.APIKey) // initialize OpenAI client first

		discordBot.Router.Register(commands.ChatCommand(&commands.ChatCommandParams{
			OpenAIClient:           openaiClient,
			OpenAICompletionModels: config.OpenAI.CompletionModels,
			GPTMessagesCache:       gptMessagesCache,
			IgnoredChannelsCache:   &ignoredChannelsCache,
		}))

		discordBot.Router.Register(commands.ImageCommand(openaiClient))
	}
	discordBot.Router.Register(commands.InfoCommand())

	// Run the bot
	discordBot.Run(config.Discord.Guild, config.Discord.RemoveCommands)
}
