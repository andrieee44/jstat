package jstat

import "encoding/json"

type Message struct {
	Name string
	Data json.RawMessage
}

func NewMessage(name string, data json.RawMessage) Message {
	return Message{
		Name: name,
		Data: data,
	}
}
