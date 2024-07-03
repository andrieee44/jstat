package jstat

import (
	"encoding/json"
	"time"
)

type Date struct {
	interval time.Duration
	format   string
	icons    []string
}

func (mod *Date) Run() (json.RawMessage, error) {
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
		Icon, Date string
	}{
		Icon: icon(mod.icons, 12, float64(hour)),
		Date: date.Format(mod.format),
	})
}

func (mod *Date) Sleep() error {
	time.Sleep(mod.interval)

	return nil
}

func (mod *Date) Cleanup() error {
	return nil
}

func NewDate(interval time.Duration, format string, icons []string) (*Date, error) {
	return &Date{
		interval: interval,
		format:   format,
		icons:    icons,
	}, nil
}
