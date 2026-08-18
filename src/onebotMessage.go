package onebotfilter

import (
	"encoding/json"
	"log"
)

type OneBotMessage struct {
	Raw     []byte
	Partial OneBotMessagePartial
	Intact  map[string]json.RawMessage
}

func ParseOneBotMessage(Raw []byte) *OneBotMessage {
	oneBotMessage := &OneBotMessage{
		Raw: Raw,
	}
	if err := json.Unmarshal(Raw, &oneBotMessage.Intact); err != nil {
		return nil
	}
	if err := json.Unmarshal(Raw, &oneBotMessage.Partial); err != nil {
		return nil
	}
	// 部分客户端不传 message_format，根据 message 的 JSON 类型自动推断
	if oneBotMessage.Partial.MessageFormat == "" {
		msg := oneBotMessage.Partial.UnDecodedMessage
		if len(msg) > 0 && msg[0] == '[' {
			oneBotMessage.Partial.MessageFormat = MESSAGE_FORMAT_ARRAY
		} else if len(msg) > 0 && msg[0] == '"' {
			oneBotMessage.Partial.MessageFormat = MESSAGE_FORMAT_STRING
		}
	}

	switch oneBotMessage.Partial.MessageFormat {
	case MESSAGE_FORMAT_ARRAY:
		if err := json.Unmarshal(oneBotMessage.Partial.UnDecodedMessage, &oneBotMessage.Partial.MessageArray); err != nil {
			log.Printf("将%s解析为array失败\n", oneBotMessage.Partial.UnDecodedMessage)
			return nil
		}
	case MESSAGE_FORMAT_STRING:
		if err := json.Unmarshal(oneBotMessage.Partial.UnDecodedMessage, &oneBotMessage.Partial.MessageString); err != nil {
			log.Printf("将%s解析为string失败\n", oneBotMessage.Partial.UnDecodedMessage)
			return nil
		}
	default: // meta_event 等没有 message 字段的事件，直接放行
		return oneBotMessage
	}
	return oneBotMessage
}

type OneBotMessagePartial struct {
	MessageType      string           `json:"message_type"`
	MessageFormat    string           `json:"message_format"`
	UnDecodedMessage json.RawMessage  `json:"message"`
	MessageArray     []MessageContent `json:"-"`
	MessageString    string           `json:"-"`
	UserId           int64            `json:"user_id"`
	GroupId          int64            `json:"group_id"`
	RawMessage       string           `json:"raw_message"`
}
type MessageContent struct {
	Type string                 `json:"type"`
	Data map[string]interface{} `json:"data"`
}
