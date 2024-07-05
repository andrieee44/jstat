package jstat

import (
	"encoding/json"
	"errors"

	"github.com/fsnotify/fsnotify"
)

type bri struct {
	icons []string

	watcher *fsnotify.Watcher
	maxBri  int
}

func (mod *bri) Init() error {
	var err error

	mod.watcher, err = fsnotify.NewWatcher()
	if err != nil {
		return err
	}

	err = mod.watcher.Add("/sys/class/backlight/intel_backlight/brightness")
	if err != nil {
		return err
	}

	mod.maxBri, err = fileAtoi("/sys/class/backlight/intel_backlight/max_brightness")
	if err != nil {
		return err
	}

	return nil
}

func (mod *bri) Run() (json.RawMessage, error) {
	var (
		bri  int
		perc float64
		err  error
	)

	bri, err = fileAtoi("/sys/class/backlight/intel_backlight/brightness")
	if err != nil {
		return nil, err
	}

	perc = float64(bri) / float64(mod.maxBri) * 100

	return json.Marshal(struct {
		Perc float64
		Icon string
	}{
		Perc: perc,
		Icon: icon(mod.icons, 100, perc),
	})
}

func (mod *bri) Sleep() error {
	var (
		event fsnotify.Event
		ok    bool
		err   error
	)

	for {
		select {
		case event, ok = <-mod.watcher.Events:
			if !ok {
				return errors.New("channel closed unexpectedly")
			}

			if event.Has(fsnotify.Write) {
				return nil
			}
		case err, ok = <-mod.watcher.Errors:
			if !ok {
				return errors.New("channel closed unexpectedly")
			}

			return err
		}
	}
}

func (mod *bri) Cleanup() error {
	return mod.watcher.Close()
}

func NewBri(icons []string) *bri {
	return &bri{
		icons: icons,
	}
}
