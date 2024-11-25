package jstat

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

type batOpts struct {
	interval time.Duration
	icons    []string
}

type bat struct {
	opts batOpts
}

func (mod *bat) Init() error {
	return nil
}

func (mod *bat) Run() (json.RawMessage, error) {
	type batInfo struct {
		Name, Status, Icon string
		Capacity           int
	}

	var (
		batsPath []string
		batsInfo []batInfo
		path     string
		status   []byte
		capacity int
		err      error
	)

	batsPath, err = filepath.Glob("/sys/class/power_supply/BAT*")
	if err != nil {
		return nil, err
	}

	for _, path = range batsPath {
		status, err = os.ReadFile(filepath.Join(path, "status"))
		if err != nil {
			return nil, err
		}

		capacity, err = fileAtoi(filepath.Join(path, "capacity"))
		if err != nil {
			return nil, err
		}

		batsInfo = append(batsInfo, batInfo{
			Name:     filepath.Base(path),
			Status:   string(status[:len(status)-1]),
			Icon:     icon(mod.opts.icons, 100, float64(capacity)),
			Capacity: capacity,
		})
	}

	return json.Marshal(batsInfo)
}

func (mod *bat) Sleep() error {
	time.Sleep(mod.opts.interval)

	return nil
}

func (mod *bat) Cleanup() error {
	return nil
}

func NewBat(interval time.Duration, icons []string) *bat {
	return &bat{
		opts: batOpts{
			interval: interval,
			icons:    icons,
		},
	}
}
