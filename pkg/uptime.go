package jstat

import (
	"encoding/json"
	"os"
	"strconv"
	"strings"
	"time"
)

type Uptime struct {
	interval time.Duration
}

func (mod *Uptime) Run() (json.RawMessage, error) {
	var (
		buf       []byte
		uptime    float64
		uptimeInt int
		err       error
	)

	buf, err = os.ReadFile("/proc/uptime")
	if err != nil {
		return nil, err
	}

	uptime, err = strconv.ParseFloat(strings.Fields(string(buf))[0], 64)
	if err != nil {
		return nil, err
	}

	uptimeInt = int(uptime)

	return json.Marshal(struct {
		Hours, Minutes, Seconds int
	}{
		Hours:   uptimeInt / 3600,
		Minutes: (uptimeInt % 3600) / 60,
		Seconds: uptimeInt % 60,
	})
}

func (mod *Uptime) Sleep() {
	time.Sleep(mod.interval)
}

func NewUptime(interval time.Duration) *Uptime {
	return &Uptime{
		interval: interval,
	}
}
