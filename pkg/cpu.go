package jstat

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type cpuCore struct {
	Freq, sum, idle int
	Usage           float64
}

type Cpu struct {
	interval time.Duration
	icons    []string
	cores    []cpuCore
}

func (mod *Cpu) Init() error {
	return nil
}

func (mod *Cpu) Run() (json.RawMessage, error) {
	var (
		stat                                    *os.File
		scanner                                 *bufio.Scanner
		fields                                  []string
		coreNStr, numStr                        string
		idx, coreN, freq, num, idle, sum, delta int
		avgUsage                                float64
		ok                                      bool
		err                                     error
	)

	stat, err = os.Open("/proc/stat")
	if err != nil {
		return nil, err
	}

	scanner = bufio.NewScanner(stat)

	for scanner.Scan() {
		fields = strings.Fields(scanner.Text())
		if fields[0] == "cpu" {
			continue
		}

		coreNStr, ok = strings.CutPrefix(fields[0], "cpu")
		if !ok {
			break
		}

		coreN, err = strconv.Atoi(coreNStr)
		if err != nil {
			return nil, err
		}

		if coreN == len(mod.cores) {
			mod.cores = append(mod.cores, cpuCore{})
		}

		freq, err = fileAtoi(fmt.Sprintf("/sys/devices/system/cpu/cpu%d/cpufreq/scaling_cur_freq", coreN))
		if err != nil {
			return nil, err
		}

		idle = 0
		sum = 0

		for idx, numStr = range fields[1:] {
			if idx == 7 {
				break
			}

			num, err = strconv.Atoi(numStr)
			if err != nil {
				return nil, err
			}

			if idx == 3 || idx == 4 {
				idle += num
			}

			sum += num
		}

		delta = sum - mod.cores[coreN].sum
		mod.cores[coreN].Usage = float64(delta-(idle-mod.cores[coreN].idle)) / float64(delta) * 100
		mod.cores[coreN].Freq = freq
		mod.cores[coreN].idle = idle
		mod.cores[coreN].sum = sum
		avgUsage += mod.cores[coreN].Usage
	}

	if scanner.Err() != nil {
		return nil, err
	}

	err = stat.Close()
	if err != nil {
		return nil, err
	}

	return json.Marshal(struct {
		Cores    []cpuCore
		AvgUsage float64
	}{
		Cores:    mod.cores,
		AvgUsage: avgUsage / float64(len(mod.cores)),
	})
}

func (mod *Cpu) Sleep() error {
	time.Sleep(mod.interval)

	return nil
}

func (mod *Cpu) Cleanup() error {
	return nil
}

func NewCpu(interval time.Duration, icons []string) *Cpu {
	return &Cpu{
		interval: interval,
		icons:    icons,
	}
}
