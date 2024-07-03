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

func runModule(msgChan chan<- jstat.Message, name string, mod jstat.Module) {
	var (
		data json.RawMessage
		err  error
	)

	panicIf(mod.Init())

	defer func() {
		panicIf(mod.Cleanup())
	}()

	for {
		data, err = mod.Run()
		panicIf(err)

		msgChan <- jstat.NewMessage(name, data)
		panicIf(mod.Sleep())
	}
}

func runConfig(ch chan<- jstat.Message) {
	var (
		name string
		mod  jstat.Module
	)

	for name, mod = range newConfig() {
		go runModule(ch, name, mod)
	}
}
