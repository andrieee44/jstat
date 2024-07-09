package jstat

import (
	"bufio"
	"os"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

func scrollEvent(nameChan <-chan string, scrollChan chan<- int, scrollInterval time.Duration, limit int) {
	var (
		name            string
		nameLen, scroll int
	)

	for {
		select {
		case name = <-nameChan:
			nameLen = utf8.RuneCountInString(name)
			scroll = 0
		case <-time.After(scrollInterval):
			if nameLen <= limit {
				continue
			}

			scroll++
			if scroll > nameLen-limit {
				scroll = 0
			}
		}

		scrollChan <- scroll
	}
}

func scroll(scrollInterval time.Duration, limit int) (chan<- string, <-chan int) {
	var (
		nameChan   chan string
		scrollChan chan int
	)

	nameChan = make(chan string)
	scrollChan = make(chan int)

	if scrollInterval == 0 || limit == 0 {
		close(nameChan)
		close(scrollChan)

		return nameChan, scrollChan
	}

	go scrollEvent(nameChan, scrollChan, scrollInterval, limit)

	return nameChan, scrollChan
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
