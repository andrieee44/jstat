package jstat

import (
	"encoding/json"
	"time"
)

type Ram struct {
	interval time.Duration
	icons    []string
}

func (mod *Ram) Init() error {
	return nil
}

func (mod *Ram) Run() (json.RawMessage, error) {
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
		Icon:      icon(mod.icons, 100, usedPerc),
	})
}

func (mod *Ram) Sleep() error {
	time.Sleep(mod.interval)

	return nil
}

func (mod *Ram) Cleanup() error {
	return nil
}

func NewRam(interval time.Duration, icons []string) *Ram {
	return &Ram{
		interval: interval,
		icons:    icons,
	}
}
