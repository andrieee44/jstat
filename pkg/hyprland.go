package jstat

import (
	"bufio"
	"encoding/json"
	"errors"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type hyprlandWorkspace struct {
	Id   int
	Name string
}

type hyprlandMonitor struct {
	Active     int
	Workspaces []hyprlandWorkspace
}

type hyprland struct {
	scrollInterval time.Duration
	limit          int

	socketPath, window      string
	eventsConn              net.Conn
	nameChan                chan<- string
	eventsChan, updatesChan chan struct{}
	errChan                 chan error
	monitors                map[string]*hyprlandMonitor
	scroll                  int
}

func (mod *hyprland) Init() error {
	var err error

	err = mod.getSocketPath()
	if err != nil {
		return err
	}

	mod.eventsConn, err = net.Dial("unix", filepath.Join(mod.socketPath, ".socket2.sock"))
	if err != nil {
		return err
	}

	mod.eventsChan = make(chan struct{}, 1)
	go mod.eventsLoop()

	mod.updatesChan = make(chan struct{}, 1)
	mod.nameChan = scrollEvent(mod.updatesChan, &mod.scroll, mod.scrollInterval, mod.limit)

	return mod.updateInfo()
}

func (mod *hyprland) Run() (json.RawMessage, error) {
	return json.Marshal(struct {
		Window        string
		Monitors      map[string]*hyprlandMonitor
		Scroll, Limit int
	}{
		Window:   mod.window,
		Monitors: mod.monitors,
		Scroll:   mod.scroll,
		Limit:    mod.limit,
	})
}

func (mod *hyprland) Sleep() error {
	var err error

	select {
	case <-mod.updatesChan:
		return nil
	case <-mod.eventsChan:
		return mod.updateInfo()
	case err = <-mod.errChan:
		return err
	}
}

func (mod *hyprland) Cleanup() error {
	return mod.eventsConn.Close()
}

func (mod *hyprland) updateInfo() error {
	type queryWorkspace struct {
		Id            int
		Monitor, Name string
	}

	type queryMonitor struct {
		Name string

		ActiveWorkspace struct {
			Id int
		}
	}

	var (
		queryConn  net.Conn
		decoder    *json.Decoder
		monitors   []queryMonitor
		monitor    queryMonitor
		workspaces []queryWorkspace
		workspace  queryWorkspace
		err        error

		window struct {
			Title string
		}
	)

	queryConn, err = net.Dial("unix", filepath.Join(mod.socketPath, ".socket.sock"))
	if err != nil {
		return err
	}

	_, err = queryConn.Write([]byte("[[BATCH]]j/activewindow;j/monitors;j/workspaces"))
	if err != nil {
		return err
	}

	decoder = json.NewDecoder(queryConn)

	err = decoder.Decode(&window)
	if err != nil {
		return err
	}

	err = decoder.Decode(&monitors)
	if err != nil {
		return err
	}

	err = decoder.Decode(&workspaces)
	if err != nil {
		return err
	}

	mod.monitors = make(map[string]*hyprlandMonitor)
	mod.window = window.Title
	mod.nameChan <- window.Title

	for _, monitor = range monitors {
		mod.monitors[monitor.Name] = &hyprlandMonitor{
			Active: monitor.ActiveWorkspace.Id,
		}
	}

	for _, workspace = range workspaces {
		mod.monitors[workspace.Monitor].Workspaces = append(mod.monitors[workspace.Monitor].Workspaces, hyprlandWorkspace{
			Id:   workspace.Id,
			Name: workspace.Name,
		})
	}

	return queryConn.Close()
}

func (mod *hyprland) getSocketPath() error {
	var his, runtime string

	his = os.Getenv("HYPRLAND_INSTANCE_SIGNATURE")
	if his == "" {
		return errors.New("HYPRLAND_INSTANCE_SIGNATURE is not set")
	}

	runtime = os.Getenv("XDG_RUNTIME_DIR")
	if runtime == "" {
		return errors.New("XDG_RUNTIME_DIR is not set")
	}

	mod.socketPath = filepath.Join(runtime, "hypr", his)

	return nil
}

func (mod *hyprland) eventsLoop() {
	var (
		scanner *bufio.Scanner
		event   string
	)

	scanner = bufio.NewScanner(mod.eventsConn)

	for {
		if !scanner.Scan() {
			mod.errChan <- scanner.Err()

			return
		}

		event = scanner.Text()

		switch {
		case strings.HasPrefix(event, "activewindow>>"):
		case strings.HasPrefix(event, "workspace>>"):
		case strings.HasPrefix(event, "destroyworkspace>>"):
		default:
			continue
		}

		mod.eventsChan <- struct{}{}
	}
}

func NewHyprland(scrollInterval time.Duration, limit int) *hyprland {
	return &hyprland{
		scrollInterval: scrollInterval,
		limit:          limit,
	}
}
