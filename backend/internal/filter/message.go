package filter

import (
	"encoding/json"

	"github.com/rs/zerolog/log"
)

type OneBotMessage struct {
	Raw     []byte
	Partial OneBotMessagePartial
	Intact  map[string]json.RawMessage
}

type OneBotMessagePartial struct {
	MessageType      string                 `json:"message_type"`
	MessageFormat    string                 `json:"message_format"`
	UnDecodedMessage json.RawMessage        `json:"message"`
	MessageArray     []OneBotMessageContent `json:"-"`
	MessageString    string                 `json:"-"`
	SelfID           int64                  `json:"self_id"`
	UserID           int64                  `json:"user_id"`
	GroupID          int64                  `json:"group_id"`
	RawMessage       string                 `json:"raw_message"`
}

type OneBotMessageContent struct {
	Type string                 `json:"type"`
	Data map[string]interface{} `json:"data"`
}

func ParseOneBotMessage(raw []byte) *OneBotMessage {
	m := &OneBotMessage{Raw: raw}
	if err := json.Unmarshal(raw, &m.Intact); err != nil {
		return nil
	}
	if err := json.Unmarshal(raw, &m.Partial); err != nil {
		return nil
	}
	switch m.Partial.MessageFormat {
	case MessageFormatArray:
		if err := json.Unmarshal(m.Partial.UnDecodedMessage, &m.Partial.MessageArray); err != nil {
			log.Debug().Bytes("payload", m.Partial.UnDecodedMessage).Msg("filter: parse message array failed")
			return nil
		}
	case MessageFormatString:
		if err := json.Unmarshal(m.Partial.UnDecodedMessage, &m.Partial.MessageString); err != nil {
			log.Debug().Bytes("payload", m.Partial.UnDecodedMessage).Msg("filter: parse message string failed")
			return nil
		}
	default:
		return nil
	}
	return m
}
