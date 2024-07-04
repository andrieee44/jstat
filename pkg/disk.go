package jstat

import (
	"encoding/json"
	"time"

	"golang.org/x/sys/unix"
)

type Disk struct {
	interval     time.Duration
	paths, icons []string
}

func (mod *Disk) Init() error {
	return nil
}

func (mod *Disk) Run() (json.RawMessage, error) {
	type diskStruct struct {
		Free, Total, Used int
		UsedPerc          float64
		Icon              string
	}

	var (
		statfs            unix.Statfs_t
		disks             map[string]diskStruct
		path              string
		free, total, used int
		usedPerc          float64
		err               error
	)

	disks = make(map[string]diskStruct)

	for _, path = range mod.paths {
		err = unix.Statfs(path, &statfs)
		if err != nil {
			return nil, err
		}

		free = int(statfs.Bfree) * int(statfs.Bsize)
		total = int(statfs.Blocks) * int(statfs.Bsize)
		used = total - free
		usedPerc = float64(used) / float64(total) * 100

		disks[path] = diskStruct{
			Free:     free,
			Total:    total,
			Used:     used,
			UsedPerc: usedPerc,
			Icon:     icon(mod.icons, 100, usedPerc),
		}
	}

	return json.Marshal(disks)
}

func (mod *Disk) Sleep() error {
	time.Sleep(mod.interval)

	return nil
}

func (mod *Disk) Cleanup() error {
	return nil
}

func NewDisk(interval time.Duration, paths, icons []string) *Disk {
	return &Disk{
		interval: interval,
		paths:    paths,
		icons:    icons,
	}
}