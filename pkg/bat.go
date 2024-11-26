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

type batInfo struct {
	Status, Icon string
	Capacity     int
}

type bat struct {
	opts batOpts
}

func (mod *bat) Init() error {
	return nil
}

func (mod *bat) Run() (json.RawMessage, error) {
	var (
		batPaths   []string
		batInfoMap map[string]*batInfo
		path       string
		err        error
	)

	batPaths, err = filepath.Glob("/sys/class/power_supply/BAT*")
	if err != nil {
		return nil, err
	}

	batInfoMap = make(map[string]*batInfo)

	for _, path = range batPaths {
		batInfoMap[filepath.Base(path)], err = mod.getBatInfo(path)
		if err != nil {
			return nil, err
		}
	}

	return json.Marshal(batInfoMap)
}

func (mod *bat) Sleep() error {
	time.Sleep(mod.opts.interval)

	return nil
}

func (mod *bat) Cleanup() error {
	return nil
}

func (mod *bat) getBatInfo(path string) (*batInfo, error) {
	var (
		status   []byte
		capacity int
		err      error
	)

	status, err = os.ReadFile(filepath.Join(path, "status"))
	if err != nil {
		return nil, err
	}

	capacity, err = fileAtoi(filepath.Join(path, "capacity"))
	if err != nil {
		return nil, err
	}

	return &batInfo{
		Status:   string(status[:len(status)-1]),
		Icon:     icon(mod.opts.icons, 100, float64(capacity)),
		Capacity: capacity,
	}, nil
}

func NewBat(interval time.Duration, icons []string) *bat {
	return &bat{
		opts: batOpts{
			interval: interval,
			icons:    icons,
		},
	}
}
