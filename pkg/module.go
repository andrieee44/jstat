package jstat

import "encoding/json"

type Module interface {
	Init() error
	Run() (json.RawMessage, error)
	Sleep() error
	Cleanup() error
}
