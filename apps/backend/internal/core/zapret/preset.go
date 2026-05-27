package zapret

import (
	"fmt"
	"strings"
)

type CommandSpec struct {
	Executable string
	Args       []string
}

func ParsePresetCommand(command string) (CommandSpec, error) {
	fields := splitCommand(command)
	if len(fields) == 0 {
		return CommandSpec{}, fmt.Errorf("preset command is empty")
	}
	return CommandSpec{Executable: fields[0], Args: fields[1:]}, nil
}

func splitCommand(command string) []string {
	var result []string
	var current strings.Builder
	inQuote := false
	var quote rune
	for _, r := range command {
		switch {
		case inQuote && r == quote:
			inQuote = false
		case !inQuote && (r == '"' || r == '\''):
			inQuote = true
			quote = r
		case !inQuote && (r == ' ' || r == '\t' || r == '\n'):
			if current.Len() > 0 {
				result = append(result, current.String())
				current.Reset()
			}
		default:
			current.WriteRune(r)
		}
	}
	if current.Len() > 0 {
		result = append(result, current.String())
	}
	return result
}
