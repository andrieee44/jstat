package jstat

import (
	"bufio"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/mdlayher/wifi"
)

type ethInfo struct {
	Powered bool
	Scroll  int
}

type wifiInfo struct {
	Name, Icon        string
	Powered, Scanning bool
	Scroll            int
	Strength          float64

	nameChan chan<- string
}

type net struct {
	interval, scrollInterval time.Duration
	limit                    int
	icons                    []string

	eth     map[string]*ethInfo
	wifi    map[string]*wifiInfo
	updates chan struct{}
	client  *wifi.Client
}

func (mod *net) Init() error {
	var err error

	mod.eth = make(map[string]*ethInfo)
	mod.wifi = make(map[string]*wifiInfo)
	mod.updates = make(chan struct{})

	mod.client, err = wifi.New()
	if err != nil {
		return err
	}

	err = mod.updateEth()
	if err != nil {
		return err
	}

	err = mod.updateWifi()
	if err != nil {
		return err
	}

	return nil
}

func (mod *net) Run() (json.RawMessage, error) {
	return json.Marshal(struct {
		Ethernet map[string]*ethInfo
		Wifi     map[string]*wifiInfo
		Limit    int
	}{
		Ethernet: mod.eth,
		Wifi:     mod.wifi,
		Limit:    mod.limit,
	})
}

func (mod *net) Sleep() error {
	var err error

	select {
	case <-time.After(mod.interval):
		err = mod.updateEth()
		if err != nil {
			return err
		}

		err = mod.updateWifi()
		if err != nil {
			return err
		}
	case <-mod.updates:
	}

	return nil
}

func (mod *net) Cleanup() error {
	return mod.client.Close()
}

func (mod *net) isPowered(iface string) (bool, error) {
	var (
		operstate []byte
		err       error
	)

	operstate, err = os.ReadFile(filepath.Join("/sys/class/net", iface, "operstate"))
	if err != nil {
		return false, err
	}

	return string(operstate[:len(operstate)-1]) == "up", nil
}

func (mod *net) isScanning(iface string) (bool, error) {
	var (
		flags []byte
		err   error
	)

	flags, err = os.ReadFile(filepath.Join("/sys/class/net", iface, "flags"))
	if err != nil {
		return false, err
	}

	return string(flags[:len(flags)-1]) == "0x1003", nil
}

func (mod *net) wifiStrength(iface string) (float64, error) {
	var (
		wireless *os.File
		err      error
		scanner  *bufio.Scanner
		fields   []string
		strength float64
		idx      int
	)

	wireless, err = os.Open("/proc/net/wireless")
	if err != nil {
		return 0, err
	}

	scanner = bufio.NewScanner(wireless)
	for idx = 0; idx < 2; idx++ {
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

	return 0, errors.New("specified interface not found in /proc/net/wireless")
}

func (mod *net) updateEth() error {
	var (
		ethIfaces []string
		ethIface  string
		ok        bool
		err       error
	)

	ethIfaces, err = filepath.Glob("/sys/class/net/e*")
	if err != nil {
		return err
	}

	for _, ethIface = range ethIfaces {
		ethIface = filepath.Base(ethIface)
		_, ok = mod.eth[ethIface]
		if !ok {
			mod.eth[ethIface] = new(ethInfo)
			scrollEvent(mod.updates, &mod.eth[ethIface].Scroll, mod.scrollInterval, mod.limit) <- ethIface
		}

		mod.eth[ethIface].Powered, err = mod.isPowered(ethIface)
		if err != nil {
			return err
		}
	}

	return nil
}

func (mod *net) updateWifi() error {
	var (
		wifiIfaces []*wifi.Interface
		wifiIface  *wifi.Interface
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

		_, ok = mod.wifi[wifiIface.Name]
		if !ok {
			mod.wifi[wifiIface.Name] = new(wifiInfo)
			mod.wifi[wifiIface.Name].nameChan = scrollEvent(mod.updates, &mod.wifi[wifiIface.Name].Scroll, mod.scrollInterval, mod.limit)
		}

		mod.wifi[wifiIface.Name].Powered, err = mod.isPowered(wifiIface.Name)
		if err != nil {
			return err
		}

		if !mod.wifi[wifiIface.Name].Powered {
			mod.wifi[wifiIface.Name].Name = ""
			mod.wifi[wifiIface.Name].nameChan <- ""
			mod.wifi[wifiIface.Name].Scanning, err = mod.isScanning(wifiIface.Name)
			if err != nil {
				return err
			}

			continue
		}

		mod.wifi[wifiIface.Name].Strength, err = mod.wifiStrength(wifiIface.Name)
		if err != nil {
			return err
		}

		mod.wifi[wifiIface.Name].Icon = icon(mod.icons, 100, mod.wifi[wifiIface.Name].Strength)

		bss, err = mod.client.BSS(wifiIface)
		if err != nil {
			return err
		}

		if mod.wifi[wifiIface.Name].Name != bss.SSID {
			mod.wifi[wifiIface.Name].Name = bss.SSID
			mod.wifi[wifiIface.Name].nameChan <- bss.SSID
		}
	}

	return nil
}

func NewNet(interval, scrollInterval time.Duration, limit int, icons []string) *net {
	return &net{
		scrollInterval: scrollInterval,
		interval:       interval,
		limit:          limit,
		icons:          icons,
	}
}
