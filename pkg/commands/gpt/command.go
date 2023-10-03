package gpt

import (
	"github.com/akhilsharma90/go-openai-bot-discord/pkg/bot"
	discord "github.com/bwmarrin/discordgo"
	"github.com/sashabaranov/go-openai"
)

var gptDefaultModel = openai.GPT3Dot5Turbo

const commandName = "gpt"

func Command(client *openai.Client, completionModels []string, messagesCache *MessagesCache, ignoredChannelsCache *IgnoredChannelsCache) *bot.Command {
	temperatureOptionMinValue := 0.0
	opts := []*discord.ApplicationCommandOption{
		{
			Type:        discord.ApplicationCommandOptionString,
			Name:        gptCommandOptionPrompt.string(),
			Description: "ChatGPT prompt",
			Required:    true,
		},
		{
			Type:        discord.ApplicationCommandOptionString,
			Name:        gptCommandOptionContext.string(),
			Description: "Sets context that guides the AI assistant's behavior during the conversation",
			Required:    false,
		},
		{
			Type:        discord.ApplicationCommandOptionAttachment,
			Name:        gptCommandOptionContextFile.string(),
			Description: "File that sets context that guides the AI assistant's behavior during the conversation",
			Required:    false,
		},
	}
	numberOfModels := len(completionModels)
	if numberOfModels > 0 {
		gptDefaultModel = completionModels[0] // set first model as default one
	}
	if numberOfModels > 1 {
		var modelChoices []*discord.ApplicationCommandOptionChoice
		for i, model := range completionModels {
			name := model
			if i == 0 {
				name += " (Default)"
			}
			modelChoices = append(modelChoices, &discord.ApplicationCommandOptionChoice{
				Name:  name,
				Value: model,
			})
		}
		opts = append(opts, &discord.ApplicationCommandOption{
			Type:        discord.ApplicationCommandOptionString,
			Name:        gptCommandOptionModel.string(),
			Description: "GPT model",
			Required:    false,
			Choices:     modelChoices,
		})
	}
	opts = append(opts, &discord.ApplicationCommandOption{
		Type:        discord.ApplicationCommandOptionNumber,
		Name:        gptCommandOptionTemperature.string(),
		Description: "What sampling temperature to use, between 0.0 and 2.0. Lower - more focused and deterministic",
		MinValue:    &temperatureOptionMinValue,
		MaxValue:    2.0,
		Required:    false,
	})
	return &bot.Command{
		Name:        commandName,
		Description: "Start conversation with ChatGPT",
		Options:     opts,
		Handler: bot.HandlerFunc(func(ctx *bot.Context) {
			chatGPTHandler(ctx, client, messagesCache)
		}),
		MessageHandler: bot.MessageHandlerFunc(func(ctx *bot.MessageContext) {
			chatGPTMessageHandler(ctx, client, messagesCache, ignoredChannelsCache)
		}),
	}
}
