package jstat

import (
	"encoding/json"
	"os"
	"strconv"
	"strings"
	"time"
)

type uptimeOpts struct {
	interval time.Duration
}

type uptime struct {
	opts *uptimeOpts
}

func (mod *uptime) Init() error {
	return nil
}

func (mod *uptime) Run() (json.RawMessage, error) {
	var (
		buf    []byte
		uptime int
		err    error
	)

	buf, err = os.ReadFile("/proc/uptime")
	if err != nil {
		return nil, err
	}

	uptime, err = strconv.Atoi(strings.Split(string(buf), ".")[0])
	if err != nil {
		return nil, err
	}

	return json.Marshal(struct {
		Hours, Minutes, Seconds int
	}{
		Hours:   uptime / 3600,
		Minutes: (uptime % 3600) / 60,
		Seconds: uptime % 60,
	})
}

func (mod *uptime) Sleep() error {
	time.Sleep(mod.opts.interval)

	return nil
}

func (mod *uptime) Close() error {
	return nil
}

func NewUptime(interval time.Duration) Module {
	return &uptime{
		opts: &uptimeOpts{interval: interval},
	}
}
