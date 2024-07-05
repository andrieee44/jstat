package main

import (
	"time"

	"github.com/andrieee44/jstat/pkg"
)

func newConfig() map[string]jstat.Module {
	var batIcons, blockIcons, clockIcons, briIcons, volIcons []string

	batIcons = []string{"󰂎", "󰁺", "󰁻", "󰁼", "󰁽", "󰁾", "󰁿", "󰂀", "󰂁", "󰂂", "󰁹"}
	blockIcons = []string{" ", "▁", "▂", "▃", "▄", "▅", "▆", "▇", "█"}
	clockIcons = []string{"󱑊", "󱐿", "󱑀", "󱑁", "󱑂", "󱑃", "󱑄", "󱑅", "󱑆", "󱑇", "󱑈", "󱑉"}
	briIcons = []string{"󰃞", "󰃟", "󰃝", "󰃠"}
	volIcons = []string{"󰕿", "󰖀", "󰕾"}

	return map[string]jstat.Module{
		"UserHost": jstat.NewUserHost(),
		"Date":     jstat.NewDate(time.Second, "Jan _2 2006 (Mon) 3:04 PM", clockIcons),
		"Uptime":   jstat.NewUptime(time.Second),
		"Bat":      jstat.NewBat(time.Second, batIcons),
		"Cpu":      jstat.NewCpu(time.Second, blockIcons),
		"Bri":      jstat.NewBri(briIcons),
		"Disk":     jstat.NewDisk(time.Minute, []string{"/"}, blockIcons),
		"Swap":     jstat.NewSwap(time.Second, blockIcons),
		"Ram":      jstat.NewRam(time.Second, blockIcons),
		"Vol":      jstat.NewVol(10*time.Millisecond, volIcons),
	}
}
