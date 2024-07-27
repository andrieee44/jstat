package jstat

import (
	"encoding/json"
	"path/filepath"
	"time"
)

type netSpd struct {
	interval time.Duration

	up, down int
}

func (mod *netSpd) Init() error {
	return nil
}

func (mod *netSpd) Run() (json.RawMessage, error) {
	var (
		sumUp, sumDown, up, down int
		err                      error
	)

	sumUp, err = mod.sumFiles("/sys/class/net/[ew]*/statistics/tx_bytes")
	if err != nil {
		return nil, err
	}

	sumDown, err = mod.sumFiles("/sys/class/net/[ew]*/statistics/rx_bytes")
	if err != nil {
		return nil, err
	}

	up = sumUp - mod.up
	down = sumDown - mod.down
	mod.up = sumUp
	mod.down = sumDown

	return json.Marshal(struct {
		Up, Down int
	}{
		Up:   up,
		Down: down,
	})
}

func (mod *netSpd) Sleep() error {
	time.Sleep(mod.interval)

	return nil
}

func (mod *netSpd) Cleanup() error {
	return nil
}

func (*netSpd) sumFiles(pattern string) (int, error) {
	var (
		paths    []string
		path     string
		num, sum int
		err      error
	)

	paths, err = filepath.Glob(pattern)
	if err != nil {
		return 0, err
	}

	for _, path = range paths {
		num, err = fileAtoi(path)
		if err != nil {
			return 0, err
		}

		sum += num
	}

	return sum, nil
}

func NewNetSpd(interval time.Duration) *netSpd {
	return &netSpd{
		interval: interval,
	}
}
