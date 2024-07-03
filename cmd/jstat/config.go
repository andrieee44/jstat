package main

import (
	"encoding/json"
	"time"

	"github.com/andrieee44/jstat/pkg"
)

func newConfig() map[string]jstat.Module {
	var batIcons, blockIcons, clockIcons []string

	batIcons = []string{"󰂎", "󰁺", "󰁻", "󰁼", "󰁽", "󰁾", "󰁿", "󰂀", "󰂁", "󰂂", "󰁹"}
	blockIcons = []string{" ", "▁", "▂", "▃", "▄", "▅", "▆", "▇", "█"}
	clockIcons = []string{"󱑊", "󱐿", "󱑀", "󱑁", "󱑂", "󱑃", "󱑄", "󱑅", "󱑆", "󱑇", "󱑈", "󱑉"}

	return map[string]jstat.Module{
		"User":   jstat.NewUser(),
		"Date":   jstat.NewDate(time.Second, "Jan _2 2006 (Mon) 3:04 PM", clockIcons),
		"Uptime": jstat.NewUptime(time.Second),
		"Bat":    jstat.NewBat(time.Second, batIcons),
		"Cpu":    jstat.NewCpu(time.Second, blockIcons),
	}
}

func runModule(msgChan chan<- jstat.Message, name string, mod jstat.Module) {
	var (
		data json.RawMessage
		err  error
	)

	for {
		data, err = mod.Run()
		panicIf(err)

		msgChan <- jstat.NewMessage(name, data)
		mod.Sleep()
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
