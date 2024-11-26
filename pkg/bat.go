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
	opts       batOpts
	batInfoMap map[string]*batInfo
}

func (mod *bat) Init() error {
	mod.batInfoMap = make(map[string]*batInfo)

	return nil
}

func (mod *bat) Run() (json.RawMessage, error) {
	var (
		batsPath []string
		path     string
		err      error
	)

	batsPath, err = filepath.Glob("/sys/class/power_supply/BAT*")
	if err != nil {
		return nil, err
	}

	for _, path = range batsPath {
		err = mod.setBatInfo(path)
		if err != nil {
			return nil, err
		}
	}

	return json.Marshal(mod.batInfoMap)
}

func (mod *bat) Sleep() error {
	time.Sleep(mod.opts.interval)

	return nil
}

func (mod *bat) Cleanup() error {
	return nil
}

func (mod *bat) setBatInfo(path string) error {
	var (
		status   []byte
		capacity int
		bat      *batInfo
		ok       bool
		err      error
	)

	status, err = os.ReadFile(filepath.Join(path, "status"))
	if err != nil {
		return err
	}

	capacity, err = fileAtoi(filepath.Join(path, "capacity"))
	if err != nil {
		return err
	}

	bat, ok = mod.batInfoMap[filepath.Base(path)]
	if !ok {
		mod.batInfoMap[filepath.Base(path)] = &batInfo{}
		bat = mod.batInfoMap[filepath.Base(path)]
	}

	bat.Status = string(status[:len(status)-1])
	bat.Icon = icon(mod.opts.icons, 100, float64(capacity))
	bat.Capacity = capacity

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
