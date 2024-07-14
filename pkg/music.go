package jstat

import (
	"encoding/json"
	"regexp"
	"time"

	"github.com/fhs/gompd/v2/mpd"
)

type music struct {
	scrollInterval time.Duration
	format         string
	limit          int

	watcher     *mpd.Watcher
	nameChan    chan<- string
	updatesChan chan struct{}
	scroll      int
	song, state string
}

func (mod *music) Init() error {
	var err error

	mod.watcher, err = mpd.NewWatcher("tcp", "127.0.0.1:6600", "", "player")
	if err != nil {
		return err
	}

	mod.updatesChan = make(chan struct{}, 1)
	mod.nameChan = scrollEvent(mod.updatesChan, &mod.scroll, mod.scrollInterval, mod.limit)

	return mod.updateInfo()
}

func (mod *music) Run() (json.RawMessage, error) {
	return json.Marshal(struct {
		Song, State   string
		Scroll, Limit int
	}{
		Song:   mod.song,
		State:  mod.state,
		Scroll: mod.scroll,
		Limit:  mod.limit,
	})
}

func (mod *music) Sleep() error {
	var err error

	select {
	case <-mod.updatesChan:
	case <-mod.watcher.Event:
		return mod.updateInfo()
	case err = <-mod.watcher.Error:
		return err
	}

	return nil
}

func (mod *music) Cleanup() error {
	return mod.watcher.Close()
}

func (mod *music) updateInfo() error {
	var (
		client       *mpd.Client
		song, status mpd.Attrs
		err          error
	)

	client, err = mpd.Dial("tcp", "127.0.0.1:6600")
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

	mod.state = status["state"]
	mod.song = regexp.MustCompilePOSIX("%[A-Za-z]+%").ReplaceAllStringFunc(mod.format, func(key string) string {
		return song[key[1:len(key)-1]]
	})

	mod.nameChan <- mod.song

	return nil
}

func NewMusic(scrollInterval time.Duration, format string, limit int) *music {
	return &music{
		scrollInterval: scrollInterval,
		format:         format,
		limit:          limit,
	}
}
