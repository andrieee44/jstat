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
	cores map[int]*cpuCore
}

func (mod *cpu) Init() error {
	mod.cores = make(map[int]*cpuCore)

	return nil
}

func (mod *cpu) Run() (json.RawMessage, error) {
	var (
		stat            *os.File
		scanner         *bufio.Scanner
		usage, avgUsage float64
		err             error
	)

	stat, err = os.Open("/proc/stat")
	if err != nil {
		return nil, err
	}

	scanner = bufio.NewScanner(stat)

	for scanner.Scan() {
		usage, err = mod.setCore(strings.Fields(scanner.Text()))
		if err != nil {
			return nil, err
		}

		avgUsage += usage
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
		Cores    map[int]*cpuCore
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

func (mod *cpu) setCore(fields []string) (float64, error) {
	var (
		coreNStr, numStr                  string
		idx, coreN, num, idle, sum, delta int
		core                              *cpuCore
		ok                                bool
		err                               error
	)

	coreNStr, ok = strings.CutPrefix(fields[0], "cpu")
	if !ok || len(coreNStr) == 0 {
		return 0, nil
	}

	coreN, err = strconv.Atoi(coreNStr)
	if err != nil {
		return 0, err
	}

	for idx, numStr = range fields[1:] {
		if idx == 7 {
			break
		}

		num, err = strconv.Atoi(numStr)
		if err != nil {
			return 0, err
		}

		if idx == 3 || idx == 4 {
			idle += num
		}

		sum += num
	}

	core, ok = mod.cores[coreN]
	if !ok {
		mod.cores[coreN] = &cpuCore{}
		core = mod.cores[coreN]
	}

	delta = sum - core.oldSum
	core.Usage = float64(delta-(idle-core.oldIdle)) / float64(delta) * 100
	core.oldSum = sum
	core.oldIdle = idle

	core.Freq, err = fileAtoi(fmt.Sprintf("/sys/devices/system/cpu/cpu%d/cpufreq/scaling_cur_freq", coreN))
	if err != nil {
		return 0, err
	}

	return core.Usage, nil
}

func NewCpu(interval time.Duration, icons []string) *cpu {
	return &cpu{
		opts: cpuOpts{
			interval: interval,
			icons:    icons,
		},
	}
}
