package jstat

import (
	"encoding/json"
	"time"
)

type swap struct {
	interval time.Duration
	icons    []string
}

func (mod *swap) Init() error {
	return nil
}

func (mod *swap) Run() (json.RawMessage, error) {
	var (
		meminfo  map[string]int
		used     int
		usedPerc float64
		err      error
	)

	meminfo, err = meminfoMap([]string{"SwapCached", "SwapTotal", "SwapFree"})
	if err != nil {
		return nil, err
	}

	used = meminfo["SwapTotal"] - meminfo["SwapFree"] + meminfo["SwapCached"]
	usedPerc = float64(used) / float64(meminfo["SwapTotal"]) * 100

	return json.Marshal(struct {
		Total, Free, Used int
		UsedPerc          float64
		Icon              string
	}{
		Total:    meminfo["SwapTotal"],
		Free:     meminfo["SwapFree"],
		Used:     used,
		UsedPerc: usedPerc,
		Icon:     icon(mod.icons, 100, usedPerc),
	})
}

func (mod *swap) Sleep() error {
	time.Sleep(mod.interval)

	return nil
}

func (mod *swap) Cleanup() error {
	return nil
}

func NewSwap(interval time.Duration, icons []string) *swap {
	return &swap{
		interval: interval,
		icons:    icons,
	}
}
