package jstat

import (
	"encoding/json"
	"path/filepath"
	"time"
)

type ethernetOpts struct {
	scrollInterval, interval time.Duration
	limit                    int
}

type ethernetOutput struct {
	Ethernets map[string]*ethernetInfo
	Limit     int
}

type ethernetInfo struct {
	Scroll   int
	nameChan chan<- string
	Powered  bool
}

type ethernet struct {
	opts        *ethernetOpts
	output      *ethernetOutput
	updatesChan chan func()
}

func (mod *ethernet) Init() error {
	mod.output = &ethernetOutput{
		Ethernets: make(map[string]*ethernetInfo),
		Limit:     mod.opts.limit,
	}

	mod.updatesChan = make(chan func())

	return mod.updateEth()
}

func (mod *ethernet) Run() (json.RawMessage, error) {
	return json.Marshal(mod.output)
}

func (mod *ethernet) Sleep() error {
	var fn func()

	select {
	case fn = <-mod.updatesChan:
		fn()

		return nil
	case <-time.After(mod.opts.interval):
		return mod.updateEth()
	}
}

func (mod *ethernet) Close() error {
	var eth *ethernetInfo

	for _, eth = range mod.output.Ethernets {
		close(eth.nameChan)
	}

	return nil
}

func (mod *ethernet) updateEth() error {
	var (
		ethPaths          []string
		ethIface, ethPath string
		info              *ethernetInfo
		powered, ok       bool
		err               error
	)

	ethPaths, err = filepath.Glob("/sys/class/net/e*")
	if err != nil {
		return err
	}

ethsLoop:
	for ethIface, info = range mod.output.Ethernets {
		for _, ethPath = range ethPaths {
			if ethIface == filepath.Base(ethPath) {
				continue ethsLoop
			}
		}

		close(info.nameChan)
		delete(mod.output.Ethernets, ethIface)
	}

	for _, ethPath = range ethPaths {
		ethPath = filepath.Base(ethPath)

		info, ok = mod.output.Ethernets[ethPath]
		if !ok {
			mod.output.Ethernets[ethPath] = new(ethernetInfo)
			info = mod.output.Ethernets[ethPath]
			info.nameChan = scrollEvent(mod.updatesChan, &info.Scroll, mod.opts.scrollInterval, mod.opts.limit)
			info.nameChan <- ethPath
		}

		powered, err = isIfacePowered(ethPath)
		if err != nil {
			return err
		}

		if info.Powered != powered {
			info.nameChan <- ethPath
		}

		info.Powered = powered
	}

	return nil
}

func NewEthernet(scrollInterval, interval time.Duration, limit int) *ethernet {
	return &ethernet{
		opts: &ethernetOpts{
			scrollInterval: scrollInterval,
			interval:       interval,
			limit:          limit,
		},
	}
}
