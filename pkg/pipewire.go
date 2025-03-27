package jstat

import (
	"encoding/json"
	"time"

	"github.com/andrieee44/pwmon/pkg"
)

type pipeWireOpts struct {
	icons           []string
	discardInterval time.Duration
}

type pipeWire struct {
	opts     *pipeWireOpts
	infoChan <-chan *pwmon.Info
	errChan  <-chan error
	info     *pwmon.Info
}

func (mod *pipeWire) Init() error {
	var err error

	mod.infoChan, mod.errChan, err = pwmon.Monitor()
	if err != nil {
		return err
	}

	select {
	case mod.info = <-mod.infoChan:
		return nil
	case err = <-mod.errChan:
		return err
	}
}

func (mod *pipeWire) Run() (json.RawMessage, error) {
	return json.Marshal(struct {
		Perc int
		Mute bool
		Icon string
	}{
		Perc: mod.info.Volume,
		Mute: mod.info.Mute,
		Icon: icon(mod.opts.icons, 100, float64(mod.info.Volume)),
	})
}

func (mod *pipeWire) Sleep() error {
	var err error

	select {
	case mod.info = <-mod.infoChan:
		return nil
	case err = <-mod.errChan:
		return err
	}
}

func (mod *pipeWire) Close() error {
	return nil
}

func NewPipeWire(discardInterval time.Duration, icons []string) *pipeWire {
	return &pipeWire{
		opts: &pipeWireOpts{
			discardInterval: discardInterval,
			icons:           icons,
		},
	}
}
