package jstat

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

type Bat struct {
	interval time.Duration
	icons    []string
}

func (mod *Bat) Init() error {
	return nil
}

func (mod *Bat) Run() (json.RawMessage, error) {
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
			Icon:     icon(mod.icons, 100, float64(capacity)),
			Capacity: capacity,
		})
	}

	return json.Marshal(batsInfo)
}

func (mod *Bat) Sleep() error {
	time.Sleep(mod.interval)

	return nil
}

func (mod *Bat) Cleanup() error {
	return nil
}

func NewBat(interval time.Duration, icons []string) *Bat {
	return &Bat{
		interval: interval,
		icons:    icons,
	}
}
