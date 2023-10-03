package gpt

import (
	"log"
	"time"

	"github.com/akhilsharma90/go-openai-bot-discord/pkg/bot"
	"github.com/akhilsharma90/go-openai-bot-discord/pkg/utils"
	discord "github.com/bwmarrin/discordgo"
	"github.com/sashabaranov/go-openai"
)

const (
	gptDiscordChannelMessagesRequestMaxRetries = 4
	gptDiscordTypingIndicatorCooldownSeconds   = 10

	gptEmojiAck = "⌛"
	gptEmojiErr = "❌"
)

func chatGPTMessageHandler(ctx *bot.MessageContext, client *openai.Client, messagesCache *MessagesCache, ignoredChannelsCache *IgnoredChannelsCache) {
	if !shouldHandleMessageType(ctx.Message.Type) {
		// ignore message types that should not be handled by this command
		return
	}

	if ctx.Session.State.User.ID == ctx.Message.Author.ID {
		// ignore self messages
		return
	}

	if _, exists := (*ignoredChannelsCache)[ctx.Message.ChannelID]; exists {
		// skip over ignored channels list
		return
	}

	if ctx.Message.Content == "" {
		// ignore messages with empty content
		return
	}

	ch, err := ctx.Session.State.Channel(ctx.Message.ChannelID)
	if err != nil {
		log.Printf("[GID: %s, CHID: %s, MID: %s] Failed to get channel info with the error: %v\n", ctx.Message.GuildID, ctx.Message.ChannelID, ctx.Message.ID, err)
		return
	}

	if !ch.IsThread() {
		// ignore non threads
		(*ignoredChannelsCache)[ctx.Message.ChannelID] = struct{}{}
		return
	}

	if ch.ThreadMetadata != nil && (ch.ThreadMetadata.Locked || ch.ThreadMetadata.Archived) {
		// We don't want to handle messages in locked or archived threads
		log.Printf("[GID: %s, CHID: %s, MID: %s] Ignoring new message in a potential thread as it is locked or/and archived\n", ctx.Message.GuildID, ctx.Message.ChannelID, ctx.Message.ID)
		return
	}

	log.Printf("[GID: %s, CHID: %s, MID: %s] Handling new message in a potential GPT thread\n", ctx.Message.GuildID, ctx.Message.ChannelID, ctx.Message.ID)

	cacheItem, ok := messagesCache.Get(ctx.Message.ChannelID)
	if !ok {
		isGPTThread := true
		cacheItem = &MessagesCacheData{}

		var lastID string
		retries := 0
		for {
			if retries >= gptDiscordChannelMessagesRequestMaxRetries {
				// max retries reached
				break
			}
			// Get messages in batches of 100 (maximum allowed by Discord API)
			batch, err := ctx.Session.ChannelMessages(ch.ID, 100, lastID, "", "")
			if err != nil {
				// Since we cannot fetch messages, that means we cannot determine whether this a GPT thread,
				// and if it was, we cannot get the full context to provide a better user experience. Do retries
				// and print the error in the log
				log.Printf("[GID: %s, CHID: %s, MID: %s] Failed to get channel messages with the error: %v. Retries left: %d\n", ctx.Message.GuildID, ctx.Message.ChannelID, ctx.Message.ID, err, (gptDiscordChannelMessagesRequestMaxRetries - retries))
				retries++
				continue
			}

			transformed := make([]openai.ChatCompletionMessage, 0, len(batch))
			for _, value := range batch {
				role := openai.ChatMessageRoleUser
				if value.Author.ID == ctx.Session.State.User.ID {
					role = openai.ChatMessageRoleAssistant
				}
				content := value.Content
				// First message is always a referenced message
				// Check if it is, and then modify to get the original prompt
				if value.Type == discord.MessageTypeThreadStarterMessage {
					if value.Author.ID != ctx.Session.State.User.ID || value.ReferencedMessage == nil {
						// this is not gpt thread, ignore
						isGPTThread = false
						break
					}
					role = openai.ChatMessageRoleUser

					prompt, context, model, temperature := parseInteractionReply(value.ReferencedMessage)
					if prompt == "" {
						isGPTThread = false
						break
					}
					content = prompt
					var systemMessage *openai.ChatCompletionMessage
					if context != "" {
						context, _ = getContentOrURLData(ctx.Client, context)
						systemMessage = &openai.ChatCompletionMessage{
							Role:    openai.ChatMessageRoleSystem,
							Content: context,
						}
					}
					if model == "" {
						model = gptDefaultModel
					}
					if temperature != nil {
						cacheItem.Temperature = temperature
					}

					cacheItem.SystemMessage = systemMessage
					cacheItem.Model = model
				} else if !shouldHandleMessageType(value.Type) {
					// ignore message types that are
					// not related to conversation
					continue
				}
				transformed = append(transformed, openai.ChatCompletionMessage{
					Role:    role,
					Content: content,
				})
			}

			reverseMessages(&transformed)

			// Add the messages to the beginning of the main list
			cacheItem.Messages = append(transformed, cacheItem.Messages...)

			// If there are no more messages in the thread, break the loop
			if len(batch) == 0 {
				break
			}

			// Set the lastID to the last message's ID to get the next batch of messages
			lastID = batch[len(batch)-1].ID
		}

		if retries >= gptDiscordChannelMessagesRequestMaxRetries {
			// max retries reached on fetching messages
			log.Printf("[GID: %s, CHID: %s, MID: %s] Failed to get channel messages. Reached max retries\n", ctx.Message.GuildID, ctx.Message.ChannelID, ctx.Message.ID)
			return
		}

		if !isGPTThread {
			// this was not a GPT thread
			log.Printf("[GID: %s, CHID: %s, MID: %s] Not a GPT thread, saving to ignored cache to skip over it later\n", ctx.Message.GuildID, ctx.Message.ChannelID, ctx.Message.ID)
			// save threadID to ignored cache, so we can always ignore it later
			(*ignoredChannelsCache)[ctx.Message.ChannelID] = struct{}{}
			return
		}

		messagesCache.Add(ctx.Message.ChannelID, cacheItem)
	} else {
		cacheItem.Messages = append(cacheItem.Messages, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleUser,
			Content: ctx.Message.Content,
		})
	}

	// check if current message cache is within allowed token limit
	if ok, count := isCacheItemWithinTruncateLimit(cacheItem); !ok {
		log.Printf("[GID: %s, CHID: %s, MID: %s] Current thread cache token count of %d exceeds truncate limit. Performing adjustments.\n", ctx.Message.GuildID, ctx.Message.ChannelID, ctx.Message.ID, count)
		adjustMessageTokens(cacheItem)
		log.Printf("[GID: %s, CHID: %s, MID: %s] Tokens adjustments finished. Current cache tokens: %d\n", ctx.Message.GuildID, ctx.Message.ChannelID, ctx.Message.ID, cacheItem.TokenCount)
	}

	// Lock the thread while we are generating ChatGPT answser
	utils.ToggleDiscordThreadLock(ctx.Session, ctx.Message.ChannelID, true)
	// Unlock the thread at the end
	defer utils.ToggleDiscordThreadLock(ctx.Session, ctx.Message.ChannelID, false)

	ctx.AddReaction(gptEmojiAck)
	defer ctx.RemoveReaction(gptEmojiAck)

	// Create a ticker and a channel for signaling request completion
	// Discord stops showing typing indicator after 10 seconds, so we
	// need to send it again
	ctx.ChannelTyping()
	typingTicker := time.NewTicker(gptDiscordTypingIndicatorCooldownSeconds * time.Second)
	done := make(chan bool)
	go func() {
		for {
			select {
			case <-typingTicker.C:
				ctx.ChannelTyping()
			case <-done:
				typingTicker.Stop()
				return
			}
		}
	}()

	log.Printf("[GID: %s, CHID: %s] ChatGPT Request invoked with [Model: %s]. Current cache size: %v\n", ctx.Message.GuildID, ctx.Message.ChannelID, cacheItem.Model, len(cacheItem.Messages))

	resp, err := sendChatGPTRequest(client, cacheItem)

	// Signal the typing ticker to stop
	done <- true

	if err != nil {
		// ChatGPT failed for whatever reason, tell users about it
		log.Printf("[GID: %s, CHID: %s] ChatGPT request ChatCompletion failed with the error: %v\n", ctx.Message.GuildID, ctx.Message.ChannelID, err)
		ctx.AddReaction(gptEmojiErr)
		ctx.EmbedReply(&discord.MessageEmbed{
			Title:       "❌ OpenAI API failed",
			Description: err.Error(),
			Color:       0xff0000,
		})
		return
	}

	log.Printf("[GID: %s, CHID: %s] ChatGPT Request [Model: %s] responded with a usage: [PromptTokens: %d, CompletionTokens: %d, TotalTokens: %d]\n", ctx.Message.GuildID, ctx.Message.ChannelID, cacheItem.Model, resp.usage.PromptTokens, resp.usage.CompletionTokens, resp.usage.TotalTokens)

	messages := splitMessage(resp.content)
	var replyMessage *discord.Message
	for _, message := range messages {
		replyMessage, err = ctx.Reply(message)
		if err != nil {
			log.Printf("[GID: %s, CHID: %s, MID: %s] Failed to reply in the thread with the error: %v\n", ctx.Message.GuildID, ctx.Message.ChannelID, ctx.Message.ID, err)
			ctx.AddReaction(gptEmojiErr)
			ctx.EmbedReply(&discord.MessageEmbed{
				Title:       "❌ Discord API Error",
				Description: err.Error(),
				Color:       0xff0000,
			})
			return
		}
	}

	attachUsageInfo(ctx.Session, replyMessage, resp.usage, cacheItem.Model)
}
