package jstat

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/mdlayher/wifi"
)

type internetOpts struct {
	icons                    []string
	scrollInterval, interval time.Duration
	limit                    int
}

type internetOutput struct {
	Internets map[string]*internetInfo
	Limit     int
}

type internetInfo struct {
	Name, Icon        string
	Scroll            int
	Strength          float64
	nameChan          chan<- string
	Powered, Scanning bool
}

type internet struct {
	opts        *internetOpts
	output      *internetOutput
	updatesChan chan func()
	client      *wifi.Client
}

func (mod *internet) Init() error {
	var err error

	mod.output = &internetOutput{
		Internets: make(map[string]*internetInfo),
		Limit:     mod.opts.limit,
	}

	mod.updatesChan = make(chan func())

	mod.client, err = wifi.New()
	if err != nil {
		return err
	}

	return mod.updateWifi()
}

func (mod *internet) Run() (json.RawMessage, error) {
	return json.Marshal(mod.output)
}

func (mod *internet) Sleep() error {
	var fn func()

	select {
	case fn = <-mod.updatesChan:
		fn()

		return nil
	case <-time.After(mod.opts.interval):
		return mod.updateWifi()
	}
}

func (mod *internet) Close() error {
	var net *internetInfo

	for _, net = range mod.output.Internets {
		close(net.nameChan)
	}

	return mod.client.Close()
}

func (mod *internet) isScanning(iface string) (bool, error) {
	var (
		flags []byte
		err   error
	)

	flags, err = os.ReadFile(filepath.Join("/sys/class/net", iface, "flags"))
	if err != nil {
		return false, err
	}

	return string(flags) == "0x1003\n", nil
}

func (mod *internet) strength(iface string) (float64, error) {
	var (
		wireless *os.File
		err      error
		scanner  *bufio.Scanner
		fields   []string
		strength float64
	)

	wireless, err = os.Open("/proc/net/wireless")
	if err != nil {
		return 0, err
	}

	scanner = bufio.NewScanner(wireless)
	for range 2 {
		if scanner.Scan() {
			continue
		}

		err = scanner.Err()
		if err != nil {
			return 0, nil
		}

		return 0, errors.New("unexpected /proc/net/wireless headers")
	}

	for scanner.Scan() {
		fields = strings.Fields(scanner.Text())

		if fields[0][:len(fields[0])-1] != iface {
			continue
		}

		strength, err = strconv.ParseFloat(fields[2], 64)
		if err != nil {
			return 0, nil
		}

		err = wireless.Close()
		if err != nil {
			return 0, nil
		}

		return strength / 70 * 100, nil
	}

	return 0, fmt.Errorf("%s: not found in /proc/net/wireless", iface)
}

func (mod *internet) updateWifi() error {
	var (
		wifiIfaces []*wifi.Interface
		wifiIface  *wifi.Interface
		info       *internetInfo
		bss        *wifi.BSS
		ok         bool
		err        error
	)

	wifiIfaces, err = mod.client.Interfaces()
	if err != nil {
		return err
	}

	for _, wifiIface = range wifiIfaces {
		if wifiIface.Type != wifi.InterfaceTypeStation {
			continue
		}

		info, ok = mod.output.Internets[wifiIface.Name]
		if !ok {
			mod.output.Internets[wifiIface.Name] = new(internetInfo)
			info = mod.output.Internets[wifiIface.Name]
			info.nameChan = scrollEvent(mod.updatesChan, &info.Scroll, mod.opts.scrollInterval, mod.opts.limit)
		}

		info.Powered, err = isIfacePowered(wifiIface.Name)
		if err != nil {
			return err
		}

		if !info.Powered {
			info.Scanning, err = mod.isScanning(wifiIface.Name)
			if err != nil {
				return err
			}

			continue
		}

		info.Strength, err = mod.strength(wifiIface.Name)
		if err != nil {
			return err
		}

		bss, err = mod.client.BSS(wifiIface)
		if err != nil {
			return err
		}

		if info.Name != bss.SSID {
			info.Name = bss.SSID
			info.nameChan <- bss.SSID
		}

		info.Icon = icon(mod.opts.icons, 100, info.Strength)
	}

	return nil
}

func NewInternet(scrollInterval, interval time.Duration, limit int, icons []string) *internet {
	return &internet{
		opts: &internetOpts{
			scrollInterval: scrollInterval,
			interval:       interval,
			limit:          limit,
			icons:          icons,
		},
	}
}
