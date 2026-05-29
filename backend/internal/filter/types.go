package filter

import (
	"encoding/json"
	"log"
)

// Filter mode constants. JSON-friendly lowercase strings.
const (
	ModeDefault   = "default"
	ModeOn        = "on"
	ModeOff       = "off"
	ModeWhitelist = "whitelist"
	ModeBlacklist = "blacklist"
)

// OneBot message types.
const (
	MessageTypePrivate = "private"
	MessageTypeGroup   = "group"
)

// OneBot message formats.
const (
	MessageFormatArray  = "array"
	MessageFormatString = "string"
	MessageContentText  = "text"
)

// IDRule is the JSON shape for a user/group ID filter rule.
type IDRule struct {
	Mode string  `json:"mode"`
	IDs  []int64 `json:"ids"`
}

// MessageRule is the JSON shape for a message filter rule.
type MessageRule struct {
	Mode          string   `json:"mode"`
	Filters       []string `json:"filters"`
	Prefix        []string `json:"prefix"`
	PrefixReplace string   `json:"prefix_replace"`
}

func DecodeIDRule(raw string) IDRule {
	r := IDRule{}
	if raw == "" {
		return r
	}
	if err := json.Unmarshal([]byte(raw), &r); err != nil {
		log.Printf("filter: decode id rule failed: %v", err)
	}
	return r
}

func DecodeMessageRule(raw string) MessageRule {
	r := MessageRule{}
	if raw == "" {
		return r
	}
	if err := json.Unmarshal([]byte(raw), &r); err != nil {
		log.Printf("filter: decode message rule failed: %v", err)
	}
	return r
}

func EncodeIDRule(r IDRule) string {
	if r.IDs == nil {
		r.IDs = []int64{}
	}
	b, _ := json.Marshal(r)
	return string(b)
}

func EncodeMessageRule(r MessageRule) string {
	if r.Filters == nil {
		r.Filters = []string{}
	}
	if r.Prefix == nil {
		r.Prefix = []string{}
	}
	b, _ := json.Marshal(r)
	return string(b)
}

type wsMsg struct {
	mt     int
	data   []byte
	selfID string
}
