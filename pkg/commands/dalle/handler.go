package dalle

import (
	"context"
	"fmt"
	"log"

	"github.com/akhilsharma90/go-openai-bot-discord/pkg/bot"
	"github.com/akhilsharma90/go-openai-bot-discord/pkg/constants"
	discord "github.com/bwmarrin/discordgo"
	"github.com/sashabaranov/go-openai"
)

func imageHandler(ctx *bot.Context, client *openai.Client) {
	var prompt string
	if option, ok := ctx.Options[imageCommandOptionPrompt.String()]; ok {
		prompt = option.StringValue()
	} else {
		// We can't have empty prompt, unfortunately
		// this should not happen, discord prevents empty required options
		log.Printf("[GID: %s, i.ID: %s] Failed to parse prompt option\n", ctx.Interaction.GuildID, ctx.Interaction.ID)
		ctx.FollowupMessageCreate(ctx.Interaction, true, &discord.WebhookParams{
			Embeds: []*discord.MessageEmbed{
				{
					Title:       "❌ Error",
					Description: " Failed to parse prompt option",
					Color:       0xff0000,
				},
			},
		})
		return
	}

	size := imageDefaultSize
	if option, ok := ctx.Options[imageCommandOptionSize.String()]; ok {
		size = option.StringValue()
		log.Printf("[GID: %s, i.ID: %s] Image size provided: %s\n", ctx.Interaction.GuildID, ctx.Interaction.ID, size)
	}

	number := 1
	if option, ok := ctx.Options[imageCommandOptionNumber.String()]; ok {
		number = int(option.IntValue())
		log.Printf("[GID: %s, i.ID: %s] Image number provided: %d\n", ctx.Interaction.GuildID, ctx.Interaction.ID, number)
	}

	log.Printf("[GID: %s, CHID: %s] Dalle Request [Size: %s, Number: %d] invoked", ctx.Interaction.GuildID, ctx.Interaction.ID, size, number)
	resp, err := client.CreateImage(
		context.Background(),
		openai.ImageRequest{
			Prompt:         prompt,
			N:              number,
			Size:           size,
			ResponseFormat: openai.CreateImageResponseFormatURL,
			User:           ctx.Interaction.Member.User.ID,
		},
	)
	if err != nil {
		log.Printf("[GID: %s, i.ID: %s] OpenAI request CreateImage failed with the error: %v\n", ctx.Interaction.GuildID, ctx.Interaction.ID, err)
		ctx.FollowupMessageCreate(ctx.Interaction, true, &discord.WebhookParams{
			Embeds: []*discord.MessageEmbed{
				{
					Title:       "❌ OpenAI API failed",
					Description: err.Error(),
					Color:       0xff0000,
				},
			},
		})
		return
	}

	log.Printf("[GID: %s, i.ID: %s] Dalle Request [Size: %s, Number: %d] responded with a data array size %d\n", ctx.Interaction.GuildID, ctx.Interaction.ID, size, number, len(resp.Data))

	var embeds = []*discord.MessageEmbed{
		{
			URL: constants.OpenAIBlackIconURL,
			Author: &discord.MessageEmbedAuthor{
				Name:         prompt,
				IconURL:      ctx.Interaction.Member.User.AvatarURL("32"),
				ProxyIconURL: constants.OpenAIBlackIconURL,
			},
			Footer: imageCreationUsageEmbedFooter(size, number),
		},
	}
	var buttonComponents []discord.MessageComponent
	for i, data := range resp.Data {
		embeds = append(embeds, &discord.MessageEmbed{
			URL: constants.OpenAIBlackIconURL,
			Image: &discord.MessageEmbedImage{
				URL:    data.URL,
				Width:  256,
				Height: 256,
			},
		})
		buttonComponents = append(buttonComponents, &discord.Button{
			Label: fmt.Sprintf("Image %d", (i + 1)),
			Style: discord.LinkButton,
			URL:   data.URL,
		})
	}

	_, err = ctx.FollowupMessageCreate(ctx.Interaction, true, &discord.WebhookParams{
		Embeds:     embeds,
		Components: []discord.MessageComponent{discord.ActionsRow{Components: buttonComponents}},
	})
	if err != nil {
		log.Printf("[GID: %s, i.ID: %s] Failed to send a follow up message with images with the error: %v\n", ctx.Interaction.GuildID, ctx.Interaction.ID, err)
		ctx.FollowupMessageCreate(ctx.Interaction, true, &discord.WebhookParams{
			Embeds: []*discord.MessageEmbed{
				{
					Title:       "❌ Discord API Error",
					Description: err.Error(),
					Color:       0xff0000,
				},
			},
		})
		return
	}

	if err != nil {
		log.Printf("[GID: %s, i.ID: %s] Discord API failed with the error: %v\n", ctx.Interaction.GuildID, ctx.Interaction.ID, err)
		ctx.FollowupMessageCreate(ctx.Interaction, true, &discord.WebhookParams{
			Content: fmt.Sprintf("> %s", prompt),
			Embeds: []*discord.MessageEmbed{
				{
					Title:       "❌ Discord API Error",
					Description: err.Error(),
					Color:       0xff0000,
				},
			},
		})
		return
	}
}
