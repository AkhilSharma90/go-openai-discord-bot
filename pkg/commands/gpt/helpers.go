package gpt

import (
	"strings"

	"github.com/sashabaranov/go-openai"
)

const discordMaxMessageLength = 2000

func splitMessage(message string) []string {
	if len(message) <= discordMaxMessageLength {
		// the message is short enough to be sent as is
		return []string{message}
	}

	// split the message by whitespace
	words := strings.Fields(message)
	var messageParts []string
	currentMessage := ""
	for _, word := range words {
		if len(currentMessage)+len(word)+1 > discordMaxMessageLength {
			// start a new message if adding the current word exceeds the maximum length
			messageParts = append(messageParts, currentMessage)
			currentMessage = word + " "
		} else {
			// add the current word to the current message
			currentMessage += word + " "
		}
	}
	// add the last message to the list of message parts
	messageParts = append(messageParts, currentMessage)

	return messageParts
}

func reverseMessages(messages *[]openai.ChatCompletionMessage) {
	length := len(*messages)
	for i := 0; i < length/2; i++ {
		(*messages)[i], (*messages)[length-i-1] = (*messages)[length-i-1], (*messages)[i]
	}
}
