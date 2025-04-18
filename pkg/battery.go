package jstat

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

type batteryOpts struct {
	icons    []string
	interval time.Duration
}

type batteryInfo struct {
	Status, Icon string
	Capacity     int
}

type battery struct {
	opts *batteryOpts
}

func (mod *battery) Init() error {
	return nil
}

func (mod *battery) Run() (json.RawMessage, error) {
	var (
		batPaths   []string
		batInfoMap map[string]*batteryInfo
		path       string
		err        error
	)

	batPaths, err = filepath.Glob("/sys/class/power_supply/BAT*")
	if err != nil {
		return nil, err
	}

	batInfoMap = make(map[string]*batteryInfo)

	for _, path = range batPaths {
		batInfoMap[filepath.Base(path)], err = mod.getBatInfo(path)
		if err != nil {
			return nil, err
		}
	}

	return json.Marshal(batInfoMap)
}

func (mod *battery) Sleep() error {
	time.Sleep(mod.opts.interval)

	return nil
}

func (mod *battery) Close() error {
	return nil
}

func (mod *battery) getBatInfo(path string) (*batteryInfo, error) {
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

	return &batteryInfo{
		Status:   string(status[:len(status)-1]),
		Capacity: capacity,
		Icon:     icon(mod.opts.icons, 100, float64(capacity)),
	}, nil
}

// NewBattery returns a [Module] that reports the battery information of the laptop.
func NewBattery(interval time.Duration, icons []string) Module {
	return &battery{
		opts: &batteryOpts{
			interval: interval,
			icons:    icons,
		},
	}
}
