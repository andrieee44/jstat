package jstat

import (
	"encoding/json"
	"path/filepath"
	"time"
)

type netSpeedOpts struct {
	interval time.Duration
}

type netSpeed struct {
	opts           *netSpeedOpts
	oldUp, oldDown int
}

func (mod *netSpeed) Init() error {
	return nil
}

func (mod *netSpeed) Run() (json.RawMessage, error) {
	var (
		up, down, deltaUp, deltaDown int
		err                          error
	)

	up, err = sumFiles("/sys/class/net/[ew]*/statistics/tx_bytes")
	if err != nil {
		return nil, err
	}

	down, err = sumFiles("/sys/class/net/[ew]*/statistics/rx_bytes")
	if err != nil {
		return nil, err
	}

	deltaUp = up - mod.oldUp
	deltaDown = down - mod.oldDown
	mod.oldUp = up
	mod.oldDown = down

	return json.Marshal(struct {
		Up, Down int
	}{
		Up:   deltaUp,
		Down: deltaDown,
	})
}

func (mod *netSpeed) Sleep() error {
	time.Sleep(mod.opts.interval)

	return nil
}

func (mod *netSpeed) Close() error {
	return nil
}

func sumFiles(pattern string) (int, error) {
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

func NewNetSpeed(interval time.Duration) *netSpeed {
	return &netSpeed{
		opts: &netSpeedOpts{
			interval: interval,
		},
	}
}
