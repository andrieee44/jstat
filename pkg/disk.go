package jstat

import (
	"encoding/json"
	"time"

	"golang.org/x/sys/unix"
)

type diskOpts struct {
	paths, icons []string
	interval     time.Duration
}

type disk struct {
	opts *diskOpts
}

func (mod *disk) Init() error {
	return nil
}

func (mod *disk) Run() (json.RawMessage, error) {
	type diskInfo struct {
		Free, Total, Used int
		UsedPerc          float64
		Icon              string
	}

	var (
		diskInfoMap       map[string]*diskInfo
		path              string
		statfs            unix.Statfs_t
		total, free, used int
		usedPerc          float64
		err               error
	)

	diskInfoMap = make(map[string]*diskInfo)

	for _, path = range mod.opts.paths {
		err = unix.Statfs(path, &statfs)
		if err != nil {
			return nil, err
		}

		total = int(statfs.Blocks) * int(statfs.Bsize)
		free = int(statfs.Bfree) * int(statfs.Bsize)
		used = total - free
		usedPerc = float64(used) / float64(total) * 100

		diskInfoMap[path] = &diskInfo{
			Total:    total,
			Free:     free,
			Used:     used,
			UsedPerc: usedPerc,
			Icon:     icon(mod.opts.icons, 100, usedPerc),
		}
	}

	return json.Marshal(diskInfoMap)
}

func (mod *disk) Sleep() error {
	time.Sleep(mod.opts.interval)

	return nil
}

func (mod *disk) Close() error {
	return nil
}

func NewDisk(interval time.Duration, paths, icons []string) Module {
	return &disk{
		opts: &diskOpts{
			interval: interval,
			paths:    paths,
			icons:    icons,
		},
	}
}
