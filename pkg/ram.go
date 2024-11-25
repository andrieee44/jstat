package jstat

import (
	"encoding/json"
	"time"
)

type ramOpts struct {
	interval time.Duration
	icons    []string
}

type ram struct {
	opts ramOpts
}

func (mod *ram) Init() error {
	return nil
}

func (mod *ram) Run() (json.RawMessage, error) {
	var (
		meminfo  map[string]int
		used     int
		usedPerc float64
		err      error
	)

	meminfo, err = meminfoMap([]string{"MemTotal", "MemFree", "MemAvailable", "Buffers", "Cached"})
	if err != nil {
		return nil, err
	}

	used = meminfo["MemTotal"] - meminfo["MemFree"] - meminfo["Buffers"] - meminfo["Cached"]
	usedPerc = float64(used) / float64(meminfo["MemTotal"]) * 100

	return json.Marshal(struct {
		Total, Free, Available, Used int
		UsedPerc                     float64
		Icon                         string
	}{
		Total:     meminfo["MemTotal"],
		Free:      meminfo["MemFree"],
		Available: meminfo["MemAvailable"],
		Used:      used,
		UsedPerc:  usedPerc,
		Icon:      icon(mod.opts.icons, 100, usedPerc),
	})
}

func (mod *ram) Sleep() error {
	time.Sleep(mod.opts.interval)

	return nil
}

func (mod *ram) Cleanup() error {
	return nil
}

func NewRam(interval time.Duration, icons []string) *ram {
	return &ram{
		opts: ramOpts{
			interval: interval,
			icons:    icons,
		},
	}
}
