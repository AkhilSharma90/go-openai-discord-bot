package utils

import (
	"log"

	discord "github.com/bwmarrin/discordgo"
)

// ToggleThreadLock locks or unlocks a Discord thread, based on the 'locked' parameter.
func ToggleDiscordThreadLock(s *discord.Session, channelID string, locked bool) {
	_, err := s.ChannelEditComplex(channelID, &discord.ChannelEdit{
		Locked: &locked,
	})
	if err != nil {
		log.Printf("[CHID: %s] Failed to lock/unlock Thread with the error: %v\n", channelID, err)
	}
}

// Sends a message to a specified Discord channel, either as a reply to another message if a message reference is provided or as a standalone message if the message reference is nil
func DiscordChannelMessageSend(s *discord.Session, channelID string, content string, messageReference *discord.MessageReference) (m *discord.Message, err error) {
	if messageReference != nil {
		m, err = s.ChannelMessageSendReply(channelID, content, messageReference)
	} else {
		m, err = s.ChannelMessageSend(channelID, content)
	}
	return
}

func DiscordChannelMessageEdit(s *discord.Session, messageID string, channelID string, content *string, embeds []*discord.MessageEmbed) error {
	_, err := s.ChannelMessageEditComplex(
		&discord.MessageEdit{
			Content: content,
			Embeds:  embeds,
			ID:      messageID,
			Channel: channelID,
		},
	)
	return err
}
