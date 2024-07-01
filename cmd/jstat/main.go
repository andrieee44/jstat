package main

import (
	"encoding/json"
	"fmt"

	"github.com/andrieee44/jstat/pkg"
)

func panicIf(err error) {
	if err != nil {
		panic(fmt.Errorf("jstat: %s", err))
	}
}

func main() {
	var (
		msgChan chan jstat.Message
		msgJson map[string]json.RawMessage
		msg     jstat.Message
		data    []byte
		err     error
	)

	msgChan = make(chan jstat.Message)
	msgJson = make(map[string]json.RawMessage)
	runConfig(msgChan)

	for msg = range msgChan {
		msgJson[msg.Name] = msg.Data

		data, err = json.Marshal(msgJson)
		panicIf(err)

		_, err = fmt.Println(string(data))
		panicIf(err)
	}
}
