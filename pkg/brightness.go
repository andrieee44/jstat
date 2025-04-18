package jstat

import (
	"encoding/json"

	"github.com/fsnotify/fsnotify"
)

type brightnessOpts struct {
	icons []string
}

type brightness struct {
	opts    *brightnessOpts
	watcher *fsnotify.Watcher
	maxBri  int
}

func (mod *brightness) Init() error {
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

func (mod *brightness) Run() (json.RawMessage, error) {
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
		Icon: icon(mod.opts.icons, 100, perc),
	})
}

func (mod *brightness) Sleep() error {
	var (
		event fsnotify.Event
		err   error
	)

	for {
		select {
		case event = <-mod.watcher.Events:
			if event.Has(fsnotify.Write) {
				return nil
			}
		case err = <-mod.watcher.Errors:
			return err
		}
	}
}

func (mod *brightness) Close() error {
	return mod.watcher.Close()
}

func NewBrightness(icons []string) Module {
	return &brightness{
		opts: &brightnessOpts{icons: icons},
	}
}
