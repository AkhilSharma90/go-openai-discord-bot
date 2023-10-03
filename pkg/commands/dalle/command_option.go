package dalle

import "fmt"

type imageCommandOptionType uint8

const (
	imageCommandOptionPrompt imageCommandOptionType = 1
	imageCommandOptionSize   imageCommandOptionType = 2
	imageCommandOptionNumber imageCommandOptionType = 3
)

func (t imageCommandOptionType) String() string {
	switch t {
	case imageCommandOptionPrompt:
		return "prompt"
	case imageCommandOptionSize:
		return "size"
	case imageCommandOptionNumber:
		return "number"
	}
	return fmt.Sprintf("ApplicationCommandOptionType(%d)", t)
}
