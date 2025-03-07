package main

import (
	"encoding/json"
	"fmt"

	"github.com/andrieee44/jstat/pkg"
)

type message struct {
	name string
	data json.RawMessage
}

func newMessage(name string, data json.RawMessage) message {
	return message{
		name: name,
		data: data,
	}
}

func panicIf(err error) {
	if err != nil {
		panic(fmt.Errorf("jstat: %s", err))
	}
}

func runModule(msgChan chan<- message, name string, mod jstat.Module) {
	var (
		data json.RawMessage
		err  error
	)

	panicIf(mod.Init())

	defer func() {
		panicIf(mod.Close())
	}()

	for {
		data, err = mod.Run()
		panicIf(err)

		msgChan <- newMessage(name, data)
		panicIf(mod.Sleep())
	}
}

func runConfig(msgChan chan<- message) {
	var (
		name string
		mod  jstat.Module
	)

	for name, mod = range newConfig() {
		go runModule(msgChan, name, mod)
	}
}
