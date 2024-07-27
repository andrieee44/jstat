package jstat

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/godbus/dbus/v5"
)

type bluetoothDevice struct {
	Name, Icon      string
	Battery, Scroll int
	Connected       bool

	nameChan chan<- string
}

type bluetoothAdapter struct {
	Name    string
	Scroll  int
	Powered bool
	Devices map[dbus.ObjectPath]*bluetoothDevice

	nameChan chan<- string
}

type bluetooth struct {
	scrollInterval time.Duration
	limit          int
	icons          []string

	sysbus      *dbus.Conn
	adapters    map[dbus.ObjectPath]*bluetoothAdapter
	updatesChan chan struct{}
	events      chan *dbus.Signal
}

func (mod *bluetooth) Init() error {
	var (
		objects map[dbus.ObjectPath]map[string]map[string]dbus.Variant
		path    dbus.ObjectPath
		object  map[string]map[string]dbus.Variant
		err     error
	)

	mod.sysbus, err = dbus.ConnectSystemBus()
	if err != nil {
		return err
	}

	mod.adapters = make(map[dbus.ObjectPath]*bluetoothAdapter)
	mod.updatesChan = make(chan struct{}, 10)
	mod.events = make(chan *dbus.Signal, 10)
	mod.sysbus.Signal(mod.events)

	err = mod.sysbus.Object("org.bluez", "/").Call("org.freedesktop.DBus.ObjectManager.GetManagedObjects", 0).Store(&objects)
	if err != nil {
		return err
	}

	err = mod.sysbus.AddMatchSignal(dbus.WithMatchObjectPath("/"), dbus.WithMatchInterface("org.freedesktop.DBus.ObjectManager"), dbus.WithMatchMember("InterfacesAdded"))
	if err != nil {
		return err
	}

	for path, object = range objects {
		err = mod.addAdapter(path, object)
		if err != nil {
			return err
		}

		err = mod.addDevice(path, object)
		if err != nil {
			return err
		}
	}

	return nil
}

func (mod *bluetooth) Run() (json.RawMessage, error) {
	return json.Marshal(struct {
		Adapters map[dbus.ObjectPath]*bluetoothAdapter
		Limit    int
	}{
		Adapters: mod.adapters,
		Limit:    mod.limit,
	})
}

func (mod *bluetooth) Sleep() error {
	var (
		signal  *dbus.Signal
		iface   string
		path    dbus.ObjectPath
		object  map[string]map[string]dbus.Variant
		members map[string]dbus.Variant
		err     error
	)

	select {
	case <-mod.updatesChan:
	case signal = <-mod.events:
		switch signal.Name {
		case "org.freedesktop.DBus.ObjectManager.InterfacesAdded":
			err = dbus.Store(signal.Body, &path, &object)
			if err != nil {
				return err
			}

			err = mod.addAdapter(path, object)
			if err != nil {
				return err
			}

			err = mod.addDevice(path, object)
			if err != nil {
				return err
			}
		case "org.freedesktop.DBus.Properties.PropertiesChanged":
			err = dbus.Store(signal.Body, &iface, &members, &[]string{})
			if err != nil {
				return err
			}

			err = mod.updateAdapter(signal.Path, iface, members)
			if err != nil {
				return err
			}

			err = mod.updateDevice("", signal.Path, iface, members)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (mod *bluetooth) Cleanup() error {
	return mod.sysbus.Close()
}

func (mod *bluetooth) getAdapter(path dbus.ObjectPath) (dbus.ObjectPath, error) {
	var (
		adapterPath dbus.ObjectPath
		ok          bool
	)

	for adapterPath = range mod.adapters {
		_, ok = mod.adapters[adapterPath].Devices[path]
		if ok {
			return adapterPath, nil
		}
	}

	return "", fmt.Errorf("%q: adapter not found", path)
}

func (mod *bluetooth) updateAdapter(path dbus.ObjectPath, iface string, members map[string]dbus.Variant) error {
	var (
		name        string
		powered, ok bool
		err         error
	)

	if iface != "org.bluez.Adapter1" {
		return nil
	}

	_, ok = members["Name"]
	if ok {
		err = members["Name"].Store(&name)
		if err != nil {
			return err
		}

		mod.adapters[path].Name = name
		mod.adapters[path].nameChan <- name
	}

	_, ok = members["Powered"]
	if ok {
		err = members["Powered"].Store(&powered)
		if err != nil {
			return err
		}

		mod.adapters[path].Powered = powered
	}

	return nil
}

func (mod *bluetooth) updateDevice(adapter, path dbus.ObjectPath, iface string, members map[string]dbus.Variant) error {
	var (
		name          string
		connected, ok bool
		percentage    int
		err           error
	)

	if iface != "org.bluez.Device1" && iface != "org.bluez.Battery1" {
		return nil
	}

	if adapter == "" {
		adapter, err = mod.getAdapter(path)
		if err != nil {
			return err
		}
	}

	_, ok = members["Name"]
	if ok {
		err = members["Name"].Store(&name)
		if err != nil {
			return err
		}

		mod.adapters[adapter].Devices[path].Name = name
		mod.adapters[adapter].Devices[path].nameChan <- name
	}

	_, ok = members["Connected"]
	if ok {
		err = members["Connected"].Store(&connected)
		if err != nil {
			return err
		}

		mod.adapters[adapter].Devices[path].Connected = connected
	}

	_, ok = members["Percentage"]
	if ok {
		err = members["Percentage"].Store(&percentage)
		if err != nil {
			return err
		}

		mod.adapters[adapter].Devices[path].Battery = percentage
		mod.adapters[adapter].Devices[path].Icon = icon(mod.icons, 100, float64(percentage))
	}

	return nil
}

func (mod *bluetooth) addAdapter(path dbus.ObjectPath, object map[string]map[string]dbus.Variant) error {
	var (
		ok  bool
		err error
	)

	_, ok = object["org.bluez.Adapter1"]
	if !ok {
		return nil
	}

	mod.adapters[path] = new(bluetoothAdapter)
	mod.adapters[path].nameChan = scrollEvent(mod.updatesChan, &mod.adapters[path].Scroll, mod.scrollInterval, mod.limit)

	err = mod.updateAdapter(path, "org.bluez.Adapter1", object["org.bluez.Adapter1"])
	if err != nil {
		return err
	}

	err = mod.sysbus.AddMatchSignal(dbus.WithMatchObjectPath(path), dbus.WithMatchInterface("org.freedesktop.DBus.Properties"), dbus.WithMatchMember("PropertiesChanged"))

	return err
}

func (mod *bluetooth) addDevice(path dbus.ObjectPath, object map[string]map[string]dbus.Variant) error {
	var (
		ok         bool
		adapter    dbus.ObjectPath
		percentage int
		err        error
	)

	_, ok = object["org.bluez.Device1"]
	if !ok {
		return nil
	}

	err = object["org.bluez.Device1"]["Adapter"].Store(&adapter)
	if err != nil {
		return err
	}

	_, ok = object["org.bluez.Battery1"]
	if ok {
		err = object["org.bluez.Battery1"]["Percentage"].Store(&percentage)
		if err != nil {
			return err
		}
	}

	if mod.adapters[adapter].Devices == nil {
		mod.adapters[adapter].Devices = make(map[dbus.ObjectPath]*bluetoothDevice)
	}

	mod.adapters[adapter].Devices[path] = new(bluetoothDevice)
	mod.adapters[adapter].Devices[path].nameChan = scrollEvent(mod.updatesChan, &mod.adapters[adapter].Devices[path].Scroll, mod.scrollInterval, mod.limit)
	mod.adapters[adapter].Devices[path].Battery = percentage
	mod.adapters[adapter].Devices[path].Icon = icon(mod.icons, 100, float64(percentage))

	err = mod.updateDevice(adapter, path, "org.bluez.Device1", object["org.bluez.Device1"])
	if err != nil {
		return err
	}

	err = mod.sysbus.AddMatchSignal(dbus.WithMatchObjectPath(path), dbus.WithMatchInterface("org.freedesktop.DBus.Properties"), dbus.WithMatchMember("PropertiesChanged"))

	return err
}

func NewBluetooth(scrollInterval time.Duration, limit int, icons []string) *bluetooth {
	return &bluetooth{
		scrollInterval: scrollInterval,
		limit:          limit,
		icons:          icons,
	}
}
