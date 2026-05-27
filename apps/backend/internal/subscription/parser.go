package subscription

import "fmt"

type Parser interface {
	ParseLink(raw string) (ParsedNode, error)
}

type ParsedNode struct {
	Name      string
	Protocol  string
	Address   string
	Port      int
	UUID      string
	Security  string
	Transport string
	RawLink   string
	Params    map[string]string
}

func ParseLink(raw string) (ParsedNode, error) {
	if raw == "" {
		return ParsedNode{}, fmt.Errorf("link is required")
	}
	return ParseVLESS(raw)
}
