package bot

import (
	discord "github.com/bwmarrin/discordgo"
)

type OptionsMap = map[string]*discord.ApplicationCommandInteractionDataOption

type Context struct {
	*discord.Session
	Caller      *Command
	Interaction *discord.Interaction
	Options     OptionsMap

	handlers []Handler
}

func makeOptionMap(options []*discord.ApplicationCommandInteractionDataOption) (m OptionsMap) {
	m = make(OptionsMap, len(options))

	for _, option := range options {
		m[option.Name] = option
	}

	return
}

func NewContext(s *discord.Session, caller *Command, i *discord.Interaction, parent *discord.ApplicationCommandInteractionDataOption, handlers []Handler) *Context {
	options := i.ApplicationCommandData().Options
	if parent != nil {
		options = parent.Options
	}
	return &Context{
		Session:     s,
		Caller:      caller,
		Interaction: i,
		Options:     makeOptionMap(options),

		handlers: handlers,
	}
}

func (ctx *Context) Respond(response *discord.InteractionResponse) error {
	return ctx.Session.InteractionRespond(ctx.Interaction, response)
}

func (ctx *Context) Edit(content string) error {
	_, err := ctx.Session.InteractionResponseEdit(ctx.Interaction, &discord.WebhookEdit{
		Content: &content,
	})
	return err
}

func (ctx *Context) Response() (*discord.Message, error) {
	return ctx.Session.InteractionResponse(ctx.Interaction)
}

func (ctx *Context) Next() {
	if ctx.handlers == nil || len(ctx.handlers) == 0 {
		return
	}

	handler := ctx.handlers[0]
	ctx.handlers = ctx.handlers[1:]

	handler.HandleCommand(ctx)
}

type MessageContext struct {
	*discord.Session
	Caller  *Command
	Message *discord.Message

	handlers []MessageHandler
}

func NewMessageContext(s *discord.Session, caller *Command, m *discord.Message, handlers []MessageHandler) *MessageContext {
	return &MessageContext{
		Session: s,
		Caller:  caller,
		Message: m,

		handlers: handlers,
	}
}

func (ctx *MessageContext) Reply(content string) (m *discord.Message, err error) {
	m, err = ctx.Session.ChannelMessageSendReply(
		ctx.Message.ChannelID,
		content,
		ctx.Message.Reference(),
	)
	return
}

func (ctx *MessageContext) EmbedReply(embed *discord.MessageEmbed) (m *discord.Message, err error) {
	m, err = ctx.Session.ChannelMessageSendEmbedReply(
		ctx.Message.ChannelID,
		embed,
		ctx.Message.Reference(),
	)
	return
}

func (ctx *MessageContext) AddReaction(emojiID string) error {
	return ctx.Session.MessageReactionAdd(ctx.Message.ChannelID, ctx.Message.ID, emojiID)
}

func (ctx *MessageContext) RemoveReaction(emojiID string) error {
	return ctx.Session.MessageReactionsRemoveEmoji(ctx.Message.ChannelID, ctx.Message.ID, emojiID)
}

func (ctx *MessageContext) ChannelTyping() error {
	return ctx.Session.ChannelTyping(ctx.Message.ChannelID)
}

func (ctx *MessageContext) Next() {
	if ctx.handlers == nil || len(ctx.handlers) == 0 {
		return
	}

	handler := ctx.handlers[0]
	ctx.handlers = ctx.handlers[1:]

	handler.HandleMessageCommand(ctx)
}
