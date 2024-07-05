package main

import (
	"encoding/json"
	"fmt"
)

func main() {
	var (
		msgChan chan message
		msgJson map[string]json.RawMessage
		msg     message
		data    []byte
		err     error
	)

	msgChan = make(chan message)
	msgJson = make(map[string]json.RawMessage)
	runConfig(msgChan)

	for msg = range msgChan {
		msgJson[msg.name] = msg.data

		data, err = json.Marshal(msgJson)
		panicIf(err)

		_, err = fmt.Println(string(data))
		panicIf(err)
	}
}
