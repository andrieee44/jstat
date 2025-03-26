package jstat

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type hyprlandOpts struct {
	scrollInterval time.Duration
	limit          int
}

type hyprlandOutput struct {
	Window                                        string
	Monitors                                      map[int]*hyprlandMonitor
	ActiveMonitor, ActiveWorkspace, Scroll, Limit int
}

type hyprlandMonitor struct {
	Name       string
	Workspaces map[int]string
}

type hyprland struct {
	socketPath  string
	opts        *hyprlandOpts
	output      *hyprlandOutput
	eventsChan  <-chan func() error
	errChan     <-chan error
	nameChan    chan<- string
	updatesChan chan func()
}

func (mod *hyprland) Init() error {
	var err error

	mod.socketPath, err = mod.getSocketPath()
	if err != nil {
		return err
	}

	mod.output, err = mod.initOutput()
	if err != nil {
		return err
	}

	mod.eventsChan, mod.errChan, err = mod.events()
	if err != nil {
		return err
	}

	mod.updatesChan = make(chan func())
	mod.nameChan = scrollEvent(mod.updatesChan, &mod.output.Scroll, mod.opts.scrollInterval, mod.opts.limit)
	mod.nameChan <- mod.output.Window

	return nil
}

func (mod *hyprland) Run() (json.RawMessage, error) {
	return json.Marshal(mod.output)
}

func (mod *hyprland) Sleep() error {
	var (
		fn      func()
		eventFn func() error
		err     error
	)

	select {
	case fn = <-mod.updatesChan:
		fn()
	case eventFn = <-mod.eventsChan:
		return eventFn()
	case err = <-mod.errChan:
		return err
	}

	return nil
}

func (mod *hyprland) Close() error {
	close(mod.nameChan)

	return nil
}

func (mod *hyprland) initOutput() (*hyprlandOutput, error) {
	type queryMonitor struct {
		Id      int
		Name    string
		Focused bool

		ActiveWorkspace struct {
			Id int
		}
	}

	type queryWorkspace struct {
		MonitorID, Id int
		Name          string
	}

	type queryWindow struct {
		Title string
	}

	var (
		queryConn  net.Conn
		decoder    *json.Decoder
		value      any
		output     *hyprlandOutput
		monitors   []queryMonitor
		monitor    queryMonitor
		workspaces []queryWorkspace
		workspace  queryWorkspace
		window     queryWindow
		err        error
	)

	queryConn, err = net.Dial("unix", filepath.Join(mod.socketPath, ".socket.sock"))
	if err != nil {
		return nil, err
	}

	_, err = queryConn.Write([]byte("[[BATCH]]j/monitors;j/workspaces;j/activewindow"))
	if err != nil {
		return nil, err
	}

	decoder = json.NewDecoder(queryConn)
	for _, value = range []any{&monitors, &workspaces, &window} {
		err = decoder.Decode(value)
		if err != nil {
			return nil, err
		}
	}

	output = &hyprlandOutput{
		Window:   window.Title,
		Monitors: make(map[int]*hyprlandMonitor),
		Limit:    mod.opts.limit,
	}

	for _, monitor = range monitors {
		if monitor.Focused {
			output.ActiveMonitor = monitor.Id
			output.ActiveWorkspace = monitor.ActiveWorkspace.Id
		}

		output.Monitors[monitor.Id] = &hyprlandMonitor{
			Name:       monitor.Name,
			Workspaces: make(map[int]string),
		}
	}

	for _, workspace = range workspaces {
		output.Monitors[workspace.MonitorID].Workspaces[workspace.Id] = workspace.Name
	}

	return output, queryConn.Close()
}

func (mod *hyprland) getSocketPath() (string, error) {
	var his, runtime string

	his = os.Getenv("HYPRLAND_INSTANCE_SIGNATURE")
	if his == "" {
		return "", errors.New("HYPRLAND_INSTANCE_SIGNATURE is not set")
	}

	runtime = os.Getenv("XDG_RUNTIME_DIR")
	if runtime == "" {
		return "", errors.New("XDG_RUNTIME_DIR is not set")
	}

	return filepath.Join(runtime, "hypr", his), nil
}

func (mod *hyprland) eventHandler(eventData string) func() error {
	var event, args []string

	event = strings.Split(eventData, ">>")
	args = strings.Split(event[1], ",")

	switch event[0] {
	case "workspacev2":
		return func() error {
			var (
				id  int
				err error
			)

			id, err = strconv.Atoi(args[0])
			if err != nil {
				return err
			}

			mod.output.ActiveWorkspace = id

			return nil
		}
	case "activewindow":
		return func() error {
			mod.output.Window = args[1]
			mod.nameChan <- args[1]

			return nil
		}
	case "monitoraddedv2":
		return func() error {
			var (
				id  int
				err error
			)

			id, err = strconv.Atoi(args[0])
			if err != nil {
				return err
			}

			mod.output.Monitors[id] = &hyprlandMonitor{
				Name:       args[1],
				Workspaces: make(map[int]string),
			}

			return nil
		}
	case "monitorremoved":
		return func() error {
			var k int

			for k = range mod.output.Monitors {
				if mod.output.Monitors[k].Name == args[0] {
					delete(mod.output.Monitors, k)

					return nil
				}
			}

			return fmt.Errorf("monitor %s: not found", args[0])
		}
	case "createworkspacev2":
		return func() error {
			var (
				id  int
				err error
			)

			id, err = strconv.Atoi(args[0])
			if err != nil {
				return err
			}

			mod.output.Monitors[mod.output.ActiveMonitor].Workspaces[id] = args[1]

			return nil
		}
	case "destroyworkspacev2":
		return func() error {
			var (
				id  int
				err error
			)

			id, err = strconv.Atoi(args[0])
			if err != nil {
				return err
			}

			delete(mod.output.Monitors[mod.output.ActiveMonitor].Workspaces, id)

			return nil
		}
	case "moveworkspacev2":
		return func() error {
			var (
				k, id int
				err   error
			)

			id, err = strconv.Atoi(args[0])
			if err != nil {
				return err
			}

			for k = range mod.output.Monitors {
				if mod.output.Monitors[k].Name == args[2] {
					mod.output.Monitors[k].Workspaces[id] = args[1]
					delete(mod.output.Monitors[mod.output.ActiveMonitor].Workspaces, id)

					return nil
				}
			}

			return fmt.Errorf("monitor %s: not found", args[2])
		}
	case "renameworkspace":
		return func() error {
			var (
				id  int
				err error
			)

			id, err = strconv.Atoi(args[0])
			if err != nil {
				return err
			}

			mod.output.Monitors[mod.output.ActiveMonitor].Workspaces[id] = args[1]

			return nil
		}
	default:
		return nil
	}
}

func (mod *hyprland) events() (<-chan func() error, <-chan error, error) {
	var (
		scanner    *bufio.Scanner
		eventsConn net.Conn
		eventsChan chan func() error
		errChan    chan error
		eventFn    func() error
		err        error
	)

	eventsConn, err = net.Dial("unix", filepath.Join(mod.socketPath, ".socket2.sock"))
	if err != nil {
		return nil, nil, err
	}

	eventsChan = make(chan func() error)
	errChan = make(chan error)
	scanner = bufio.NewScanner(eventsConn)

	go func() {
		var err error

		for scanner.Scan() {
			eventFn = mod.eventHandler(scanner.Text())
			if eventFn != nil {
				eventsChan <- eventFn
			}
		}

		err = scanner.Err()
		if err != nil {
			errChan <- err
		}

		close(eventsChan)
		close(errChan)
	}()

	return eventsChan, errChan, nil
}

func NewHyprland(scrollInterval time.Duration, limit int) *hyprland {
	return &hyprland{
		opts: &hyprlandOpts{
			scrollInterval: scrollInterval,
			limit:          limit,
		},
	}
}
