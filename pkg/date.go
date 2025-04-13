package jstat

import (
	"encoding/json"
	"time"
)

type dateOpts struct {
	icons    []string
	format   string
	interval time.Duration
}

type date struct {
	opts *dateOpts
}

func (mod *date) Init() error {
	return nil
}

func (mod *date) Run() (json.RawMessage, error) {
	var (
		date time.Time
		hour int
	)

	date = time.Now()
	hour = date.Hour()

	if hour >= 12 {
		hour -= 12
	}

	return json.Marshal(struct {
		Date, Icon string
	}{
		Date: date.Format(mod.opts.format),
		Icon: icon(mod.opts.icons, 12, float64(hour)),
	})
}

func (mod *date) Sleep() error {
	time.Sleep(mod.opts.interval)

	return nil
}

func (mod *date) Close() error {
	return nil
}

func NewDate(interval time.Duration, format string, icons []string) Module {
	return &date{
		opts: &dateOpts{
			interval: interval,
			format:   format,
			icons:    icons,
		},
	}
}
