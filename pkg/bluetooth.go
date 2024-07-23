package jstat

import (
	"encoding/json"
	"time"

	"github.com/godbus/dbus/v5"
)

type bluetoothDevice struct {
	Name, Icon            string
	Battery, Scroll       int
	HasBattery, Connected bool

	changed chan struct{}
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
	events      chan *dbus.Signal
	updatesChan chan struct{}
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
	mod.events = make(chan *dbus.Signal, 10)
	mod.updatesChan = make(chan struct{}, 1)

	err = mod.sysbus.Object("org.bluez", "/").Call("org.freedesktop.DBus.ObjectManager.GetManagedObjects", 0).Store(&objects)
	if err != nil {
		return err
	}

	err = mod.sysbus.AddMatchSignal(dbus.WithMatchObjectPath("/"), dbus.WithMatchInterface("org.freedesktop.DBus.ObjectManager"), dbus.WithMatchMember("InterfacesAdded"))
	if err != nil {
		return err
	}

	for path, object = range objects {
		mod.updateAdapter(path, object)
		mod.addDevice(path, object)

		err = mod.sysbus.AddMatchSignal(dbus.WithMatchObjectPath(path), dbus.WithMatchInterface("org.freedesktop.DBus.Properties"), dbus.WithMatchMember("PropertiesChanged"))
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
	select {
	case <-mod.updatesChan:
	case _ = <-mod.events:
	}

	return nil
}

func (mod *bluetooth) Cleanup() error {
	return mod.sysbus.Close()
}

func (mod *bluetooth) updateAdapter(path dbus.ObjectPath, object map[string]map[string]dbus.Variant) error {
	var (
		name        string
		powered, ok bool
		err         error
	)

	_, ok = object["org.bluez.Adapter1"]
	if !ok {
		return nil
	}

	err = object["org.bluez.Adapter1"]["Name"].Store(&name)
	if err != nil {
		return err
	}

	err = object["org.bluez.Adapter1"]["Powered"].Store(&powered)
	if err != nil {
		return err
	}

	if mod.adapters[path] == nil {
		mod.adapters[path] = &bluetoothAdapter{
			Devices: make(map[dbus.ObjectPath]*bluetoothDevice),
		}

		mod.adapters[path].nameChan = scrollEvent(mod.updatesChan, &mod.adapters[path].Scroll, mod.scrollInterval, mod.limit)
	}

	if mod.adapters[path].Name != name {
		mod.adapters[path].Name = name
		mod.adapters[path].nameChan <- name
	}

	mod.adapters[path].Powered = powered

	return nil
}

func (mod *bluetooth) addDevice(path dbus.ObjectPath, object map[string]map[string]dbus.Variant) {
}

func NewBluetooth(scrollInterval time.Duration, limit int, icons []string) *bluetooth {
	return &bluetooth{
		scrollInterval: scrollInterval,
		limit:          limit,
		icons:          icons,
	}
}
