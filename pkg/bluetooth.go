package jstat

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/godbus/dbus/v5"
)

type bluetoothUpdater func(dbus.ObjectPath, map[string]dbus.Variant) error

type bluetoothOpts struct {
	icons          []string
	scrollInterval time.Duration
	limit          int
}

type bluetoothOutput struct {
	Adapters map[dbus.ObjectPath]*bluetoothAdapter
	Limit    int
}

type bluetoothDevice struct {
	Name, Icon      string
	Battery, Scroll int
	nameChan        chan<- string
	Connected       bool
}

type bluetoothAdapter struct {
	Name                 string
	Devices              map[dbus.ObjectPath]*bluetoothDevice
	Scroll               int
	nameChan             chan<- string
	Powered, Discovering bool
}

type bluetooth struct {
	opts        *bluetoothOpts
	output      *bluetoothOutput
	updaters    map[string]bluetoothUpdater
	sysbus      *dbus.Conn
	updatesChan chan func()
	signalChan  chan *dbus.Signal
}

func (mod *bluetooth) Init() error {
	var (
		objects map[dbus.ObjectPath]map[string]map[string]dbus.Variant
		paths   []dbus.ObjectPath
		path    dbus.ObjectPath
		ok      bool
		err     error
	)

	mod.output = &bluetoothOutput{
		Adapters: make(map[dbus.ObjectPath]*bluetoothAdapter),
		Limit:    mod.opts.limit,
	}

	mod.updaters = map[string]bluetoothUpdater{
		"org.bluez.Adapter1": mod.updateAdapter,
		"org.bluez.Device1":  mod.updateDevice,
		"org.bluez.Battery1": mod.updateBattery,
	}

	mod.sysbus, err = dbus.ConnectSystemBus()
	if err != nil {
		return err
	}

	mod.updatesChan = make(chan func())
	mod.signalChan = make(chan *dbus.Signal, 10)
	mod.sysbus.Signal(mod.signalChan)

	err = mod.sysbus.Object("org.bluez", "/").Call("org.freedesktop.DBus.ObjectManager.GetManagedObjects", 0).Store(&objects)
	if err != nil {
		return err
	}

	err = mod.sysbus.AddMatchSignal(dbus.WithMatchDestination("org.bluez"), dbus.WithMatchObjectPath("/"), dbus.WithMatchInterface("org.freedesktop.DBus.ObjectManager"), dbus.WithMatchMember("InterfacesAdded"))
	if err != nil {
		return err
	}

	for path = range objects {
		_, ok = objects[path]["org.bluez.Adapter1"]
		if ok {
			paths = append([]dbus.ObjectPath{path}, paths...)

			continue
		}

		paths = append(paths, path)
	}

	for _, path = range paths {
		err = mod.updater(objects[path], path)
		if err != nil {
			return err
		}
	}

	return nil
}

func (mod *bluetooth) Run() (json.RawMessage, error) {
	return json.Marshal(mod.output)
}

func (mod *bluetooth) Sleep() error {
	var (
		fn     func()
		signal *dbus.Signal
	)

	select {
	case fn = <-mod.updatesChan:
		fn()
	case signal = <-mod.signalChan:
		mod.signalHandler(signal)
	}

	return nil
}

func (mod *bluetooth) Close() error {
	var (
		adapter *bluetoothAdapter
		device  *bluetoothDevice
	)

	for _, adapter = range mod.output.Adapters {
		close(adapter.nameChan)

		for _, device = range adapter.Devices {
			close(device.nameChan)
		}
	}

	return mod.sysbus.Close()
}

func (mod *bluetooth) updater(object map[string]map[string]dbus.Variant, path dbus.ObjectPath) error {
	var (
		iface   string
		fn      bluetoothUpdater
		members map[string]dbus.Variant
		ok      bool
		err     error
	)

	for iface, fn = range mod.updaters {
		members, ok = object[iface]
		if !ok {
			continue
		}

		err = fn(path, members)
		if err != nil {
			return err
		}
	}

	return nil
}

func (mod *bluetooth) storeVariants(members map[string]dbus.Variant, variants map[string]any) error {
	var (
		variantName string
		variant     dbus.Variant
		value       any
		ok          bool
		err         error
	)

	for variantName, value = range variants {
		variant, ok = members[variantName]
		if !ok {
			continue
		}

		err = variant.Store(value)
		if err != nil {
			return err
		}
	}

	return nil
}

func (mod *bluetooth) deviceAdapter(devicePath dbus.ObjectPath) (dbus.ObjectPath, error) {
	var (
		adapter dbus.ObjectPath
		ok      bool
	)

	for adapter = range mod.output.Adapters {
		_, ok = mod.output.Adapters[adapter].Devices[devicePath]
		if ok {
			return adapter, nil
		}
	}

	return "", fmt.Errorf("%q: device not found in existing adapters", devicePath)
}

func (mod *bluetooth) updateAdapter(path dbus.ObjectPath, members map[string]dbus.Variant) error {
	var (
		adapter *bluetoothAdapter
		ok      bool
		err     error
	)

	adapter, ok = mod.output.Adapters[path]
	if !ok {
		mod.output.Adapters[path] = &bluetoothAdapter{
			Devices: make(map[dbus.ObjectPath]*bluetoothDevice),
		}

		adapter = mod.output.Adapters[path]
		adapter.nameChan = scrollEvent(mod.updatesChan, &adapter.Scroll, mod.opts.scrollInterval, mod.opts.limit)

		err = mod.sysbus.AddMatchSignal(dbus.WithMatchDestination("org.bluez"), dbus.WithMatchObjectPath(path), dbus.WithMatchInterface("org.freedesktop.DBus.Properties"), dbus.WithMatchMember("PropertiesChanged"))
		if err != nil {
			return err
		}
	}

	err = mod.storeVariants(members, map[string]any{
		"Name":        &adapter.Name,
		"Powered":     &adapter.Powered,
		"Discovering": &adapter.Discovering,
	})

	if err != nil {
		return err
	}

	adapter.nameChan <- adapter.Name

	return nil
}

func (mod *bluetooth) updateDevice(path dbus.ObjectPath, members map[string]dbus.Variant) error {
	var (
		adapterPath dbus.ObjectPath
		adapter     *bluetoothAdapter
		device      *bluetoothDevice
		ok          bool
		err         error
	)

	err = mod.storeVariants(members, map[string]any{
		"Adapter": &adapterPath,
	})

	if err != nil {
		return err
	}

	adapter, ok = mod.output.Adapters[adapterPath]
	if adapterPath == "" && !ok {
		adapterPath, err = mod.deviceAdapter(path)
		if err != nil {
			return err
		}

		adapter = mod.output.Adapters[adapterPath]
	}

	device, ok = adapter.Devices[path]
	if !ok {
		adapter.Devices[path] = new(bluetoothDevice)
		device = adapter.Devices[path]
		device.nameChan = scrollEvent(mod.updatesChan, &device.Scroll, mod.opts.scrollInterval, mod.opts.limit)

		err = mod.sysbus.AddMatchSignal(dbus.WithMatchDestination("org.bluez"), dbus.WithMatchObjectPath(path), dbus.WithMatchInterface("org.freedesktop.DBus.Properties"), dbus.WithMatchMember("PropertiesChanged"))
		if err != nil {
			return err
		}
	}

	err = mod.storeVariants(members, map[string]any{
		"Name":      &device.Name,
		"Connected": &device.Connected,
	})

	if err != nil {
		return err
	}

	device.nameChan <- device.Name

	return nil
}

func (mod *bluetooth) updateBattery(path dbus.ObjectPath, members map[string]dbus.Variant) error {
	var (
		adapterPath dbus.ObjectPath
		device      *bluetoothDevice
		err         error
	)

	adapterPath, err = mod.deviceAdapter(path)
	if err != nil {
		return err
	}

	device = mod.output.Adapters[adapterPath].Devices[path]

	err = mod.storeVariants(members, map[string]any{
		"Percentage": &device.Battery,
	})

	if err != nil {
		return err
	}

	device.Icon = icon(mod.opts.icons, 100, float64(device.Battery))

	return nil
}

func (mod *bluetooth) signalHandler(signal *dbus.Signal) error {
	var (
		iface   string
		path    dbus.ObjectPath
		object  map[string]map[string]dbus.Variant
		members map[string]dbus.Variant
		updater bluetoothUpdater
		ok      bool
		err     error
	)

	switch signal.Name {
	case "org.freedesktop.DBus.ObjectManager.InterfacesAdded":
		err = dbus.Store(signal.Body, &path, &object)
		if err != nil {
			return err
		}

		return mod.updater(object, path)
	case "org.freedesktop.DBus.Properties.PropertiesChanged":
		err = dbus.Store(signal.Body, &iface, &members, &[]string{})
		if err != nil {
			return err
		}

		updater, ok = mod.updaters[iface]
		if !ok {
			return nil
		}

		err = updater(signal.Path, members)
		if err != nil {
			return err
		}
	}

	return nil
}

func NewBluetooth(scrollInterval time.Duration, limit int, icons []string) *bluetooth {
	return &bluetooth{
		opts: &bluetoothOpts{
			scrollInterval: scrollInterval,
			limit:          limit,
			icons:          icons,
		},
	}
}
