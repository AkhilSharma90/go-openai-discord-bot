package gpt

import (
	lru "github.com/hashicorp/golang-lru/v2"
	"github.com/sashabaranov/go-openai"
)

type IgnoredChannelsCache map[string]struct{}

type MessagesCache struct {
	*lru.Cache[string, *MessagesCacheData]
}

type MessagesCacheData struct {
	Messages      []openai.ChatCompletionMessage
	SystemMessage *openai.ChatCompletionMessage
	Model         string
	Temperature   *float32
	TokenCount    int
}

func NewMessagesCache(size int) (*MessagesCache, error) {
	lruCache, err := lru.New[string, *MessagesCacheData](size)
	if err != nil {
		return nil, err
	}

	return &MessagesCache{
		Cache: lruCache,
	}, nil
}
