package jstat

import (
	"encoding/json"
	"regexp"
	"time"

	"github.com/adrg/xdg"
	"github.com/fhs/gompd/v2/mpd"
)

type musicOpts struct {
	format         string
	scrollInterval time.Duration
	limit          int
}

type musicOutput struct {
	Song, State   string
	Scroll, Limit int
}

type music struct {
	opts        *musicOpts
	output      *musicOutput
	watcher     *mpd.Watcher
	nameChan    chan<- string
	updatesChan chan func()
	socketPath  string
}

func (mod *music) Init() error {
	var err error

	mod.socketPath, err = xdg.SearchRuntimeFile("mpd/socket")
	if err != nil {
		return err
	}

	mod.watcher, err = mpd.NewWatcher("unix", mod.socketPath, "", "player")
	if err != nil {
		return err
	}

	mod.output = &musicOutput{Limit: mod.opts.limit}
	mod.updatesChan = make(chan func())
	mod.nameChan = scrollEvent(mod.updatesChan, &mod.output.Scroll, mod.opts.scrollInterval, mod.opts.limit)

	return mod.updateOutput()
}

func (mod *music) Run() (json.RawMessage, error) {
	return json.Marshal(mod.output)
}

func (mod *music) Sleep() error {
	var (
		fn  func()
		err error
	)

	select {
	case fn = <-mod.updatesChan:
		fn()

		return nil
	case <-mod.watcher.Event:
		return mod.updateOutput()
	case err = <-mod.watcher.Error:
		return err
	}
}

func (mod *music) Close() error {
	close(mod.nameChan)

	return mod.watcher.Close()
}

func (mod *music) updateOutput() error {
	var (
		client       *mpd.Client
		song, status mpd.Attrs
		err          error
	)

	client, err = mpd.Dial("unix", mod.socketPath)
	if err != nil {
		return err
	}

	song, err = client.CurrentSong()
	if err != nil {
		return err
	}

	status, err = client.Status()
	if err != nil {
		return err
	}

	err = client.Close()
	if err != nil {
		return err
	}

	mod.output.State = status["state"]
	mod.output.Song = regexp.MustCompilePOSIX("%[A-Za-z]+%").ReplaceAllStringFunc(mod.opts.format, func(key string) string {
		return song[key[1:len(key)-1]]
	})

	mod.nameChan <- mod.output.Song

	return nil
}

func NewMPD(scrollInterval time.Duration, format string, limit int) Module {
	return &music{
		opts: &musicOpts{
			scrollInterval: scrollInterval,
			format:         format,
			limit:          limit,
		},
	}
}
