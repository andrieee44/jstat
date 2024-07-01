package jstat

import "encoding/json"

type Module interface {
	Run() (json.RawMessage, error)
	Sleep()
}
