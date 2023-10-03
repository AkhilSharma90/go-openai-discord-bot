package gpt

import "fmt"

type gptCommandOptionType uint8

const (
	gptCommandOptionPrompt      gptCommandOptionType = 1
	gptCommandOptionContext     gptCommandOptionType = 2
	gptCommandOptionContextFile gptCommandOptionType = 3
	gptCommandOptionModel       gptCommandOptionType = 4
	gptCommandOptionTemperature gptCommandOptionType = 5
)

func (t gptCommandOptionType) string() string {
	switch t {
	case gptCommandOptionPrompt:
		return "prompt"
	case gptCommandOptionContext:
		return "context"
	case gptCommandOptionContextFile:
		return "context-file"
	case gptCommandOptionModel:
		return "model"
	case gptCommandOptionTemperature:
		return "temperature"
	}
	return fmt.Sprintf("ApplicationCommandOptionType(%d)", t)
}

func (t gptCommandOptionType) humanReadableString() string {
	switch t {
	case gptCommandOptionPrompt:
		return "Prompt"
	case gptCommandOptionContext:
		return "Context"
	case gptCommandOptionContextFile:
		return "Context file"
	case gptCommandOptionModel:
		return "Model"
	case gptCommandOptionTemperature:
		return "Temperature"
	}
	return fmt.Sprintf("ApplicationCommandOptionType(%d)", t)
}
