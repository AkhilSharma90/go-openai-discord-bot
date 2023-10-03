package dalle

import (
	"github.com/akhilsharma90/go-openai-bot-discord/pkg/bot"
	discord "github.com/bwmarrin/discordgo"
	"github.com/sashabaranov/go-openai"
)

const commandName = "dalle"

func Command(client *openai.Client) *bot.Command {
	numberOptionMinValue := 1.0
	return &bot.Command{
		Name:        commandName,
		Description: "Generate creative images from textual descriptions using OpenAI Dalle 2",
		Options: []*discord.ApplicationCommandOption{
			{
				Type:        discord.ApplicationCommandOptionString,
				Name:        imageCommandOptionPrompt.String(),
				Description: "A text description of the desired image",
				Required:    true,
			},
			{
				Type:        discord.ApplicationCommandOptionString,
				Name:        imageCommandOptionSize.String(),
				Description: "The size of the generated images",
				Required:    false,
				Choices: []*discord.ApplicationCommandOptionChoice{
					{
						Name:  openai.CreateImageSize256x256 + " (Default)",
						Value: openai.CreateImageSize256x256,
					},
					{
						Name:  openai.CreateImageSize512x512,
						Value: openai.CreateImageSize512x512,
					},
					{
						Name:  openai.CreateImageSize1024x1024,
						Value: openai.CreateImageSize1024x1024,
					},
				},
			},
			{
				Type:        discord.ApplicationCommandOptionInteger,
				Name:        imageCommandOptionNumber.String(),
				Description: "The number of images to generate (default 1, max 4)",
				MinValue:    &numberOptionMinValue,
				MaxValue:    4,
				Required:    false,
			},
		},
		Handler: bot.HandlerFunc(func(ctx *bot.Context) {
			imageHandler(ctx, client)
		}),
		Middlewares: []bot.Handler{
			bot.HandlerFunc(imageInteractionResponseMiddleware),
			bot.HandlerFunc(func(ctx *bot.Context) {
				imageModerationMiddleware(ctx, client)
			}),
		},
	}
}
