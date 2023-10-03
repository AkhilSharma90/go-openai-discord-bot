package bot

import (
	discord "github.com/bwmarrin/discordgo"
)

type Handler interface {
	HandleCommand(ctx *Context)
}

type HandlerFunc func(ctx *Context)

func (f HandlerFunc) HandleCommand(ctx *Context) { f(ctx) }

type MessageHandler interface {
	HandleMessageCommand(ctx *MessageContext)
}

type MessageHandlerFunc func(ctx *MessageContext)

func (f MessageHandlerFunc) HandleMessageCommand(ctx *MessageContext) { f(ctx) }

type Command struct {
	Name                     string
	Description              string
	DMPermission             bool
	DefaultMemberPermissions int64
	Options                  []*discord.ApplicationCommandOption
	Type                     discord.ApplicationCommandType

	Handler        Handler
	Middlewares    []Handler
	MessageHandler MessageHandler

	SubCommands *Router
}

func (cmd Command) ApplicationCommand() *discord.ApplicationCommand {
	applicationCommand := &discord.ApplicationCommand{
		Name:                     cmd.Name,
		Description:              cmd.Description,
		DMPermission:             &cmd.DMPermission,
		DefaultMemberPermissions: &cmd.DefaultMemberPermissions,
		Options:                  cmd.Options,
		Type:                     cmd.Type,
	}
	for _, subcommand := range cmd.SubCommands.List() {
		applicationCommand.Options = append(applicationCommand.Options, subcommand.ApplicationCommandOption())
	}
	return applicationCommand
}

func (cmd Command) ApplicationCommandOption() *discord.ApplicationCommandOption {
	applicationCommand := cmd.ApplicationCommand()
	typ := discord.ApplicationCommandOptionSubCommand

	if cmd.SubCommands != nil && cmd.SubCommands.Count() != 0 {
		typ = discord.ApplicationCommandOptionSubCommandGroup
	}

	return &discord.ApplicationCommandOption{
		Name:        applicationCommand.Name,
		Description: applicationCommand.Description,
		Options:     applicationCommand.Options,
		Type:        typ,
	}
}
