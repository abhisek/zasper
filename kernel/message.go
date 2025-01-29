package kernel

import (
	"time"

	"github.com/rs/zerolog/log"
)

type (
	MessageHeader struct {
		MsgID           string `json:"msg_id"`
		MsgType         string `json:"msg_type"`
		Username        string `json:"username"`
		Session         string `json:"session"`
		Date            string `json:"date"`
		ProtocolVersion string `json:"version"`
	}

	Message struct {
		Header       MessageHeader `json:"header"`
		ParentHeader MessageHeader `json:"parent_header"`
		MsgId        string        `json:"msg_id"`
		MsgType      string        `json:"msg_type"`
		Content      interface{}   `json:"content"`
		Buffers      []byte        `json:"buffers"`
		Metadata     interface{}   `json:"metadata"`
		Tracker      int           `json:"tracker"`
		Error        error         `json:"error"`
	}
)

func (ks *KernelSession) newMsgHeader(msgType string, userName string, session string) MessageHeader {
	return MessageHeader{
		MsgID:           newID(),
		MsgType:         msgType,
		Username:        userName,
		Session:         ks.Key,
		Date:            time.Now().UTC().Format(time.RFC3339),
		ProtocolVersion: ProtocolVersion,
	}
}

func (ks *KernelSession) createMsg(msgType string,
	content interface{},
	parent MessageHeader,
	header MessageHeader,
	metadata map[string]interface{}) Message {

	msg := Message{}
	// expect header is not provided

	msg.MsgId = msg.Header.MsgID
	// msg.ParentHeader = parent
	msg.Content = "{}"
	msg.Metadata = "{}"
	return msg
}

func (ks *KernelSession) MessageFromString(value string) Message {
	msg := Message{}
	msg.Header = ks.newMsgHeader(value, GetUsername(), ks.Key)
	log.Info().Msgf("message header is %v", msg.Header)
	msg.MsgId = msg.Header.MsgID
	msg.Content = "kernel_info_request"
	msg.Metadata = make(map[string]interface{})

	return msg
}

func (ks *KernelSession) MessageFromDict(value map[string]interface{}) Message {
	return Message{}
}
