package jstat

import (
	"encoding/json"
	"path/filepath"
	"time"
)

type netSpdOpts struct {
	interval time.Duration
}

type netSpd struct {
	opts           netSpdOpts
	oldUp, oldDown int
}

func (mod *netSpd) Init() error {
	return nil
}

func (mod *netSpd) Run() (json.RawMessage, error) {
	var (
		up, down, deltaUp, deltaDown int
		err                          error
	)

	up, err = mod.sumFiles("/sys/class/net/[ew]*/statistics/tx_bytes")
	if err != nil {
		return nil, err
	}

	down, err = mod.sumFiles("/sys/class/net/[ew]*/statistics/rx_bytes")
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

func (mod *netSpd) Sleep() error {
	time.Sleep(mod.opts.interval)

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
		opts: netSpdOpts{
			interval: interval,
		},
	}
}
