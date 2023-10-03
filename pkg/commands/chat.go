package commands

import (
	"github.com/akhilsharma90/go-openai-bot-discord/pkg/bot"
	"github.com/akhilsharma90/go-openai-bot-discord/pkg/commands/gpt"
	discord "github.com/bwmarrin/discordgo"
	"github.com/sashabaranov/go-openai"
)

const chatCommandName = "chat"

type ChatCommandParams struct {
	OpenAIClient           *openai.Client
	OpenAICompletionModels []string
	GPTMessagesCache       *gpt.MessagesCache
	IgnoredChannelsCache   *gpt.IgnoredChannelsCache
}

func ChatCommand(params *ChatCommandParams) *bot.Command {
	return &bot.Command{
		Name:                     chatCommandName,
		Description:              "Start conversation with LLM",
		DMPermission:             false,
		DefaultMemberPermissions: discord.PermissionViewChannel,
		Type:                     discord.ChatApplicationCommand,
		SubCommands: bot.NewRouter([]*bot.Command{
			gpt.Command(params.OpenAIClient, params.OpenAICompletionModels, params.GPTMessagesCache, params.IgnoredChannelsCache),
		}),
	}
}
