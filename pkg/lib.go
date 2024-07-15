package jstat

import (
	"bufio"
	"os"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

func scrollLoop(nameChan <-chan string, updates chan<- struct{}, scrollPtr *int, scrollInterval time.Duration, limit int) {
	var (
		name    string
		nameLen int
	)

	for {
		select {
		case name = <-nameChan:
			nameLen = utf8.RuneCountInString(name)
			*scrollPtr = 0
		case <-time.After(scrollInterval):
			if nameLen <= limit {
				name = <-nameChan
				nameLen = utf8.RuneCountInString(name)
				*scrollPtr = 0
				break
			}

			*scrollPtr++
			if *scrollPtr > nameLen-limit {
				*scrollPtr = 0
			}
		}

		updates <- struct{}{}
	}
}

func discardNameChan(nameChan <-chan string) {
	for {
		<-nameChan
	}
}

func scrollEvent(updates chan<- struct{}, scrollPtr *int, scrollInterval time.Duration, limit int) chan<- string {
	var nameChan chan string

	nameChan = make(chan string)

	if scrollInterval == 0 || limit == 0 {
		go discardNameChan(nameChan)

		return nameChan
	}

	go scrollLoop(nameChan, updates, scrollPtr, scrollInterval, limit)

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
			return append(keys[:i], keys[i+1:]...), true
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
