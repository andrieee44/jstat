package jstat

import (
	"encoding/json"
	"os"
	"strconv"
	"strings"
	"time"
)

type uptime struct {
	interval time.Duration
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
	time.Sleep(mod.interval)

	return nil
}

func (mod *uptime) Cleanup() error {
	return nil
}

func NewUptime(interval time.Duration) *uptime {
	return &uptime{interval: interval}
}
