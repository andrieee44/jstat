package jstat

import (
	"encoding/json"
	"errors"
	"regexp"
	"time"
	"unicode/utf8"

	"github.com/fhs/gompd/v2/mpd"
)

type music struct {
	scrollInterval time.Duration
	format         string
	limit          int

	song, state string
	scroll      int
	watcher     *mpd.Watcher
}

func (mod *music) Init() error {
	var err error

	mod.watcher, err = mpd.NewWatcher("tcp", "127.0.0.1:6600", "", "player")
	if err != nil {
		return err
	}

	err = mod.updateSong()
	if err != nil {
		return err
	}

	return nil
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
	var (
		timer    <-chan time.Time
		musicLen int
		ok       bool
		err      error
	)

	musicLen = utf8.RuneCountInString(mod.song)

	if mod.limit != 0 && mod.scrollInterval != 0 && musicLen > mod.limit {
		timer = time.After(mod.scrollInterval)
	}

	select {
	case _, ok = <-mod.watcher.Event:
		if !ok {
			return errors.New("channel closed unexpectedly")
		}

		mod.scroll = 0

		err = mod.updateSong()
		if err != nil {
			return err
		}
	case err, ok = <-mod.watcher.Error:
		if !ok {
			return errors.New("channel closed unexpectedly")
		}

		return err
	case <-timer:
		mod.scroll++
		if mod.scroll > musicLen-mod.limit {
			mod.scroll = 0
		}
	}

	return nil
}

func (mod *music) Cleanup() error {
	return mod.watcher.Close()
}

func (mod *music) updateSong() error {
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

	return nil
}

func NewMusic(scrollInterval time.Duration, format string, limit int) *music {
	return &music{
		scrollInterval: scrollInterval,
		format:         format,
		limit:          limit,
	}
}
