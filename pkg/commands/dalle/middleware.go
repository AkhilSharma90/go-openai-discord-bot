package dalle

import (
	"context"
	"log"

	"github.com/akhilsharma90/go-openai-bot-discord/pkg/bot"
	discord "github.com/bwmarrin/discordgo"
	"github.com/sashabaranov/go-openai"
)

func imageInteractionResponseMiddleware(ctx *bot.Context) {
	log.Printf("[GID: %s, i.ID: %s] Image interaction invoked by UserID: %s\n", ctx.Interaction.GuildID, ctx.Interaction.ID, ctx.Interaction.Member.User.ID)

	err := ctx.Respond(&discord.InteractionResponse{
		Type: discord.InteractionResponseDeferredChannelMessageWithSource,
	})
	if err != nil {
		log.Printf("[GID: %s, i.ID: %s] Failed to respond to interactrion with the error: %v\n", ctx.Interaction.GuildID, ctx.Interaction.ID, err)
		return
	}

	ctx.Next()
}

func imageModerationMiddleware(ctx *bot.Context, client *openai.Client) {
	log.Printf("[GID: %s, i.ID: %s] Performing interaction moderation middleware\n", ctx.Interaction.GuildID, ctx.Interaction.ID)

	var prompt string
	if option, ok := ctx.Options[imageCommandOptionPrompt.String()]; ok {
		prompt = option.StringValue()
	} else {
		// We can't have empty prompt, unfortunately
		// this should not happen, discord prevents empty required options
		log.Printf("[GID: %s, i.ID: %s] Failed to parse prompt option\n", ctx.Interaction.GuildID, ctx.Interaction.ID)
		ctx.Respond(&discord.InteractionResponse{
			Type: discord.InteractionResponseChannelMessageWithSource,
			Data: &discord.InteractionResponseData{
				Content: "ERROR: Failed to parse prompt option",
			},
		})
		return
	}

	resp, err := client.Moderations(
		context.Background(),
		openai.ModerationRequest{
			Input: prompt,
		},
	)
	if err != nil {
		// do not block request if moderation api failed
		log.Printf("[GID: %s, i.ID: %s] OpenAI Moderation API request failed with the error: %v\n", ctx.Interaction.GuildID, ctx.Interaction.ID, err)
		ctx.Next()
		return
	}

	if resp.Results[0].Flagged {
		// response was flagged, send error
		log.Printf("[GID: %s, i.ID: %s] Interaction was flagged by Moderation API, prompt: \"%s\"\n", ctx.Interaction.GuildID, ctx.Interaction.ID, prompt)
		ctx.FollowupMessageCreate(ctx.Interaction, true, &discord.WebhookParams{
			Embeds: []*discord.MessageEmbed{
				{
					Title:       "❌ Error",
					Description: "The provided prompt contains text that violates OpenAI's usage policies and is not allowed by their safety system",
					Color:       0xff0000,
				},
			},
		})
		return
	}

	ctx.Next()
}
