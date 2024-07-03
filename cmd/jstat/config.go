package main

import (
	"time"

	"github.com/andrieee44/jstat/pkg"
)

func newConfig() map[string]jstat.Module {
	var batIcons, blockIcons, clockIcons, briIcons []string

	batIcons = []string{"󰂎", "󰁺", "󰁻", "󰁼", "󰁽", "󰁾", "󰁿", "󰂀", "󰂁", "󰂂", "󰁹"}
	blockIcons = []string{" ", "▁", "▂", "▃", "▄", "▅", "▆", "▇", "█"}
	clockIcons = []string{"󱑊", "󱐿", "󱑀", "󱑁", "󱑂", "󱑃", "󱑄", "󱑅", "󱑆", "󱑇", "󱑈", "󱑉"}
	briIcons = []string{"󰃞", "󰃟", "󰃝", "󰃠"}

	return map[string]jstat.Module{
		"User":   newModule(jstat.NewUser()),
		"Date":   newModule(jstat.NewDate(time.Second, "Jan _2 2006 (Mon) 3:04 PM", clockIcons)),
		"Uptime": newModule(jstat.NewUptime(time.Second)),
		"Bat":    newModule(jstat.NewBat(time.Second, batIcons)),
		"Cpu":    newModule(jstat.NewCpu(time.Second, blockIcons)),
		"Bri":    newModule(jstat.NewBri(briIcons)),
	}
}
