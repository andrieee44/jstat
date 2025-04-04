package main

import (
	"time"

	"github.com/andrieee44/jstat/pkg"
)

func config() map[string]jstat.Module {
	const (
		limit     int = 15
		listLimit int = 5
	)

	var diskPaths, batIcons, blockIcons, clockIcons, briIcons, volIcons, internetIcons []string

	diskPaths = []string{"/"}
	batIcons = []string{"󰂎", "󰁺", "󰁻", "󰁼", "󰁽", "󰁾", "󰁿", "󰂀", "󰂁", "󰂂", "󰁹"}
	blockIcons = []string{" ", "▁", "▂", "▃", "▄", "▅", "▆", "▇", "█"}
	clockIcons = []string{"󱑊", "󱐿", "󱑀", "󱑁", "󱑂", "󱑃", "󱑄", "󱑅", "󱑆", "󱑇", "󱑈", "󱑉"}
	briIcons = []string{"󰃞", "󰃟", "󰃝", "󰃠"}
	volIcons = []string{"󰕿", "󰖀", "󰕾"}
	internetIcons = []string{"󰤯", "󰤟", "󰤢", "󰤥", "󰤨"}

	return map[string]jstat.Module{
		"UserHost":   jstat.NewUserHost(),
		"Date":       jstat.NewDate(time.Second, "Jan _2 2006 (Mon) 3:04 PM", clockIcons),
		"Uptime":     jstat.NewUptime(time.Second),
		"Battery":    jstat.NewBattery(time.Second, batIcons),
		"CPU":        jstat.NewCPU(time.Second, blockIcons),
		"Brightness": jstat.NewBrightness(briIcons),
		"Disk":       jstat.NewDisk(time.Minute, diskPaths, blockIcons),
		"Swap":       jstat.NewSwap(time.Second, blockIcons),
		"Ram":        jstat.NewRam(time.Second, blockIcons),
		"PipeWire":   jstat.NewPipeWire(10*time.Millisecond, volIcons),
		"MPD":        jstat.NewMPD(500*time.Millisecond, "%AlbumArtist% - %Track% - %Album% - %Title%", limit),
		"Internet":   jstat.NewInternet(500*time.Millisecond, time.Second, listLimit, internetIcons),
		"Ethernet":   jstat.NewEthernet(500*time.Millisecond, time.Second, listLimit),
		"Hyprland":   jstat.NewHyprland(500*time.Millisecond, limit),
		"Bluetooth":  jstat.NewBluetooth(500*time.Millisecond, limit-5, batIcons),
		"NetSpeed":   jstat.NewNetSpeed(time.Second),
	}
}
