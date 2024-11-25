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

type cpuOpts struct {
	interval time.Duration
	icons    []string
}

type cpuCore struct {
	Freq            int
	Usage           float64
	oldSum, oldIdle int
}

type cpu struct {
	opts  cpuOpts
	cores []cpuCore
}

func (mod *cpu) Init() error {
	return nil
}

func (mod *cpu) Run() (json.RawMessage, error) {
	var (
		stat                              *os.File
		scanner                           *bufio.Scanner
		fields                            []string
		coreNStr, numStr                  string
		idx, coreN, num, idle, sum, delta int
		avgUsage                          float64
		ok                                bool
		err                               error
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

		mod.cores[coreN].Freq, err = fileAtoi(fmt.Sprintf("/sys/devices/system/cpu/cpu%d/cpufreq/scaling_cur_freq", coreN))
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

		delta = sum - mod.cores[coreN].oldSum
		mod.cores[coreN].Usage = float64(delta-(idle-mod.cores[coreN].oldIdle)) / float64(delta) * 100
		mod.cores[coreN].oldIdle = idle
		mod.cores[coreN].oldSum = sum
		avgUsage += mod.cores[coreN].Usage
	}

	if scanner.Err() != nil {
		return nil, err
	}

	err = stat.Close()
	if err != nil {
		return nil, err
	}

	avgUsage /= float64(len(mod.cores))

	return json.Marshal(struct {
		Cores    []cpuCore
		Icon     string
		AvgUsage float64
	}{
		Cores:    mod.cores,
		Icon:     icon(mod.opts.icons, 100, avgUsage),
		AvgUsage: avgUsage,
	})
}

func (mod *cpu) Sleep() error {
	time.Sleep(mod.opts.interval)

	return nil
}

func (mod *cpu) Cleanup() error {
	return nil
}

func NewCpu(interval time.Duration, icons []string) *cpu {
	return &cpu{
		opts: cpuOpts{
			interval: interval,
			icons:    icons,
		},
	}
}
