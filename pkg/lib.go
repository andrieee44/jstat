package jstat

import (
	"os"
	"strconv"
)

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
