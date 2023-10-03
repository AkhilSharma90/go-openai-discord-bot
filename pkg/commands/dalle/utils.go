package dalle

import (
	"fmt"

	"github.com/akhilsharma90/go-openai-bot-discord/pkg/constants"
	discord "github.com/bwmarrin/discordgo"
	"github.com/sashabaranov/go-openai"
)

const (
	imageDefaultSize = openai.CreateImageSize256x256

	imagePriceSize256x256   = 0.016
	imagePriceSize512x512   = 0.018
	imagePriceSize1024x1024 = 0.02
)

func priceForResponse(n int, size string) float64 {
	switch size {
	case openai.CreateImageSize256x256:
		return float64(n) * imagePriceSize256x256
	case openai.CreateImageSize512x512:
		return float64(n) * imagePriceSize512x512
	case openai.CreateImageSize1024x1024:
		return float64(n) * imagePriceSize1024x1024
	}

	return 0
}

func imageCreationUsageEmbedFooter(size string, number int) *discord.MessageEmbedFooter {
	extraInfo := fmt.Sprintf("Size: %s, Images: %d", size, number)
	price := priceForResponse(number, size)
	if price > 0 {
		extraInfo += fmt.Sprintf("\nGeneration Cost: $%g", price)
	}
	return &discord.MessageEmbedFooter{
		Text:    extraInfo,
		IconURL: constants.OpenAIBlackIconURL,
	}
}
