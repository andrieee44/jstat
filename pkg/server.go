package jstat

import (
	"encoding/json"
)

type message struct {
	name string
	data json.RawMessage
}

func runModule(msgChan chan<- message, errChan chan<- error, name string, mod Module) {
	var (
		data json.RawMessage
		err  error
	)

	err = mod.Init()
	if err != nil {
		errChan <- err

		return
	}

	defer func() {
		err = mod.Close()
		if err != nil {
			errChan <- err
		}
	}()

	for {
		data, err = mod.Run()
		if err != nil {
			errChan <- err

			return
		}

		msgChan <- message{
			name: name,
			data: data,
		}

		err = mod.Sleep()
		if err != nil {
			errChan <- err

			return
		}
	}
}

func serve(msgChan <-chan message, errChan chan<- error, dataChan chan<- json.RawMessage) {
	var (
		msg     message
		dataMap map[string]json.RawMessage
		data    json.RawMessage
		err     error
	)

	dataMap = make(map[string]json.RawMessage)

	for msg = range msgChan {
		dataMap[msg.name] = msg.data

		data, err = json.Marshal(dataMap)
		if err != nil {
			errChan <- err

			return
		}

		dataChan <- data
	}
}

func NewServer(modules map[string]Module) (<-chan json.RawMessage, <-chan error) {
	var (
		name     string
		mod      Module
		msgChan  chan message
		dataChan chan json.RawMessage
		errChan  chan error
	)

	msgChan = make(chan message)
	dataChan = make(chan json.RawMessage)
	errChan = make(chan error)

	for name, mod = range modules {
		go runModule(msgChan, errChan, name, mod)
	}

	go serve(msgChan, errChan, dataChan)

	return dataChan, errChan
}
