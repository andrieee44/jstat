package jstat

import (
	"encoding/json"
	"time"

	"github.com/mafik/pulseaudio"
)

type volOpts struct {
	discardInterval time.Duration
	icons           []string
}

type vol struct {
	opts    volOpts
	client  *pulseaudio.Client
	updates <-chan struct{}
}

func (mod *vol) Init() error {
	var err error

	mod.client, err = pulseaudio.NewClient()
	if err != nil {
		return err
	}

	mod.updates, err = mod.client.Updates()
	if err != nil {
		return err
	}

	return nil
}

func (mod *vol) Run() (json.RawMessage, error) {
	var (
		volume     float32
		volumePerc float64
		mute       bool
		err        error
	)

	volume, err = mod.client.Volume()
	if err != nil {
		return nil, err
	}

	mute, err = mod.client.Mute()
	if err != nil {
		return nil, err
	}

	volumePerc = float64(volume) * 100

	return json.Marshal(struct {
		Perc float64
		Mute bool
		Icon string
	}{
		Perc: volumePerc,
		Mute: mute,
		Icon: icon(mod.opts.icons, 100, volumePerc),
	})
}

func (mod *vol) Sleep() error {
	var timer <-chan time.Time

	<-mod.updates
	timer = time.After(mod.opts.discardInterval)

	for {
		select {
		case <-mod.updates:
		case <-timer:
			return nil
		}
	}
}

func (mod *vol) Cleanup() error {
	mod.client.Close()

	return nil
}

func NewVol(discardInterval time.Duration, icons []string) *vol {
	return &vol{
		opts: volOpts{
			discardInterval: discardInterval,
			icons:           icons,
		},
	}
}
