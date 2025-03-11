package main

import (
	"encoding/json"
	"fmt"

	"github.com/andrieee44/jstat/pkg"
)

func main() {
	var (
		dataChan <-chan json.RawMessage
		errChan  <-chan error
		data     json.RawMessage
		err      error
	)

	dataChan, errChan = jstat.NewServer(config())

	for {
		select {
		case data = <-dataChan:
			_, err = fmt.Println(string(data))
			if err != nil {
				panic(err)
			}
		case err = <-errChan:
			panic(err)
		}
	}

}
