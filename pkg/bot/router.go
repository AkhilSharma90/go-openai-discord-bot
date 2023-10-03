package bot

import (
	"fmt"
	"log"

	"github.com/bwmarrin/discordgo"
	discord "github.com/bwmarrin/discordgo"
)

type Router struct {
	commands           map[string]*Command
	registeredCommands []*discord.ApplicationCommand
}

func NewRouter(initial []*Command) (r *Router) {
	r = &Router{commands: make(map[string]*Command, len(initial))}
	for _, cmd := range initial {
		r.Register(cmd)
	}

	return
}

func (r *Router) Register(cmd *Command) {
	if _, ok := r.commands[cmd.Name]; !ok {
		r.commands[cmd.Name] = cmd
	}
}

func (r *Router) Get(name string) *Command {
	if r == nil {
		return nil
	}
	return r.commands[name]
}

func (r *Router) List() (list []*Command) {
	if r == nil {
		return nil
	}

	for _, c := range r.commands {
		list = append(list, c)
	}
	return
}

func (r *Router) Count() (c int) {
	if r == nil {
		return 0
	}
	return len(r.commands)
}

func (r *Router) getSubcommand(cmd *Command, opt *discord.ApplicationCommandInteractionDataOption, parent []Handler) (*Command, *discord.ApplicationCommandInteractionDataOption, []Handler) {
	if cmd == nil {
		return nil, nil, nil
	}

	subcommand := cmd.SubCommands.Get(opt.Name)
	switch opt.Type {
	case discordgo.ApplicationCommandOptionSubCommand:
		return subcommand, opt, append(parent, append(subcommand.Middlewares, subcommand.Handler)...)
	case discordgo.ApplicationCommandOptionSubCommandGroup:
		return r.getSubcommand(subcommand, opt.Options[0], append(parent, subcommand.Middlewares...))
	}

	return cmd, nil, append(parent, cmd.Handler)
}

func (r *Router) getMessageHandlers(cmd *Command) []MessageHandler {
	var handlers []MessageHandler

	if cmd.MessageHandler != nil {
		handlers = append(handlers, cmd.MessageHandler)
	}

	if cmd.SubCommands != nil {
		for _, cmd := range cmd.SubCommands.List() {
			handlers = append(handlers, r.getMessageHandlers(cmd)...)
		}
	}

	return handlers
}

func (r *Router) HandleInteraction(s *discord.Session, i *discord.InteractionCreate) {
	if i.Type != discord.InteractionApplicationCommand {
		return
	}

	data := i.ApplicationCommandData()
	cmd := r.Get(data.Name)
	if cmd == nil {
		return
	}

	var parent *discord.ApplicationCommandInteractionDataOption
	handlers := append(cmd.Middlewares, cmd.Handler)
	if len(data.Options) != 0 {
		cmd, parent, handlers = r.getSubcommand(cmd, data.Options[0], cmd.Middlewares)
	}

	if cmd != nil {
		ctx := NewContext(s, cmd, i.Interaction, parent, handlers)
		ctx.Next()
	}
}

func (r *Router) HandleMessage(s *discord.Session, m *discord.MessageCreate) {
	for _, cmd := range r.commands {
		handlers := r.getMessageHandlers(cmd)
		if len(handlers) > 0 {
			ctx := NewMessageContext(s, cmd, m.Message, handlers)
			ctx.Next()
		}
	}
}

func (r *Router) Sync(s *discord.Session, guild string) (err error) {
	if s.State.User == nil {
		return fmt.Errorf("cannot determine application id")
	}

	var commands []*discord.ApplicationCommand
	for _, c := range r.commands {
		commands = append(commands, c.ApplicationCommand())
	}

	r.registeredCommands, err = s.ApplicationCommandBulkOverwrite(s.State.User.ID, guild, commands)
	return
}

func (r *Router) ClearCommands(s *discord.Session, guild string) (errors []error) {
	if s.State.User == nil {
		return []error{fmt.Errorf("cannot determine application id")}
	}

	for _, v := range r.registeredCommands {
		err := s.ApplicationCommandDelete(s.State.User.ID, guild, v.ID)
		if err != nil {
			log.Panicf("Cannot delete '%v' command: %v", v.Name, err)
		}
	}

	if len(errors) == 0 {
		return nil
	}

	return errors
}
