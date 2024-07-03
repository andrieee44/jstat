package jstat

import (
	"encoding/json"
	"errors"

	"github.com/fsnotify/fsnotify"
)

type Bri struct {
	icons   []string
	watcher *fsnotify.Watcher
	maxBri  int
}

func (mod *Bri) Run() (json.RawMessage, error) {
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

func (mod *Bri) Sleep() error {
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

func (mod *Bri) Cleanup() error {
	return mod.watcher.Close()
}

func NewBri(icons []string) (*Bri, error) {
	var (
		watcher *fsnotify.Watcher
		maxBri  int
		err     error
	)

	watcher, err = fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	err = watcher.Add("/sys/class/backlight/intel_backlight/brightness")
	if err != nil {
		return nil, err
	}

	maxBri, err = fileAtoi("/sys/class/backlight/intel_backlight/max_brightness")
	if err != nil {
		return nil, err
	}

	return &Bri{
		icons:   icons,
		watcher: watcher,
		maxBri:  maxBri,
	}, nil
}
