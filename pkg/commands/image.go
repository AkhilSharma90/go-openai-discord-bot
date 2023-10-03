package commands

import (
	"github.com/akhilsharma90/go-openai-bot-discord/pkg/bot"
	"github.com/akhilsharma90/go-openai-bot-discord/pkg/commands/dalle"
	discord "github.com/bwmarrin/discordgo"
	"github.com/sashabaranov/go-openai"
)

const imageCommandName = "image"

func ImageCommand(client *openai.Client) *bot.Command {
	return &bot.Command{
		Name:                     imageCommandName,
		Description:              "Generate creative images from textual descriptions",
		DMPermission:             false,
		DefaultMemberPermissions: discord.PermissionViewChannel,
		SubCommands: bot.NewRouter([]*bot.Command{
			dalle.Command(client),
		}),
	}
}
