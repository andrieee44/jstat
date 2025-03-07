package jstat

import (
	"bufio"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

func scrollLoop(nameChan <-chan string, updates chan<- func(), scrollPtr *int, scrollInterval time.Duration, limit int) {
	var (
		timer           <-chan time.Time
		name            string
		nameLen, scroll int
		ok              bool
	)

	for {
		timer = nil
		if nameLen > limit {
			timer = time.After(scrollInterval)
		}

		select {
		case name, ok = <-nameChan:
			if !ok {
				return
			}

			nameLen = utf8.RuneCountInString(name)
			scroll = 0
		case <-timer:
			scroll++
			if scroll > nameLen-limit {
				scroll = 0
			}
		}

		updates <- func() {
			*scrollPtr = scroll
		}
	}
}

func scrollEvent(updates chan<- func(), scrollPtr *int, scrollInterval time.Duration, limit int) chan<- string {
	var nameChan chan string

	nameChan = make(chan string, 1)
	if scrollInterval > 0 && limit > 0 {
		go scrollLoop(nameChan, updates, scrollPtr, scrollInterval, limit)

		return nameChan
	}

	go func() {
		for {
			<-nameChan
		}
	}()

	return nameChan
}

func icon(icons []string, max, val float64) string {
	var index, iconsLen int

	iconsLen = len(icons)
	if iconsLen == 0 {
		return ""
	}

	index = int(float64(iconsLen) / max * val)
	if index >= iconsLen {
		return icons[iconsLen-1]
	}

	return icons[index]
}

func fileAtoi(file string) (int, error) {
	var (
		buf []byte
		num int
		err error
	)

	buf, err = os.ReadFile(file)
	if err != nil {
		return 0, err
	}

	num, err = strconv.Atoi(string(buf[:len(buf)-1]))
	if err != nil {
		return 0, err
	}

	return num, nil
}

func removeKey(keys []string, key string) ([]string, bool) {
	var (
		i int
		v string
	)

	for i, v = range keys {
		if v == key {
			return slices.Delete(keys, i, i+1), true
		}
	}

	return keys, false
}

func meminfoMap(keys []string) (map[string]int, error) {
	var (
		keyVal  map[string]int
		meminfo *os.File
		scanner *bufio.Scanner
		fields  []string
		key     string
		val     int
		ok      bool
		err     error
	)

	keyVal = make(map[string]int)

	meminfo, err = os.Open("/proc/meminfo")
	if err != nil {
		return nil, err
	}

	scanner = bufio.NewScanner(meminfo)

	for scanner.Scan() {
		fields = strings.Fields(scanner.Text())
		key = fields[0][:len(fields[0])-1]

		keys, ok = removeKey(keys, key)
		if !ok {
			continue
		}

		val, err = strconv.Atoi(fields[1])
		if err != nil {
			return nil, err
		}

		keyVal[key] = val

		if len(keys) == 0 {
			break
		}
	}

	if scanner.Err() != nil {
		return nil, err
	}

	err = meminfo.Close()
	if err != nil {
		return nil, err
	}

	return keyVal, nil
}

func isIfacePowered(iface string) (bool, error) {
	var (
		operstate []byte
		err       error
	)

	operstate, err = os.ReadFile(filepath.Join("/sys/class/net", iface, "operstate"))
	if err != nil {
		return false, err
	}

	return string(operstate) == "up\n", nil
}
