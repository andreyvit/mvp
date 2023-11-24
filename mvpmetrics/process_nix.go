package mvpmetrics

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

const userHZ = 100 // USER_HZ

// readFileNoStat is like os.ReadFile, but does not call f.Stat() which is unreliable for procfs
func readFileNoStat(filename string) ([]byte, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return io.ReadAll(io.LimitReader(f, 1024*1024))
}

func collectProcessMetrics(pm *processMetrics) error {
	const mountPoint = "/proc"
	if _, err := os.Stat(mountPoint); err != nil {
		return fmt.Errorf("invalid procfs mount point %q: %w", mountPoint, err)
	}
	pid := os.Getpid()

	err1 := collectProcessStatMetrics(pm, fmt.Sprintf("%s/%d/stat", mountPoint, pid))

	var err2 error
	if isRealProcFS(mountPoint) && collectProcessFdMetricsFast(pm, fmt.Sprintf("%s/%d/fd", mountPoint, pid)) {
		// nop
	} else {
		err2 = collectProcessFdMetricsSlow(pm, fmt.Sprintf("%s/%d/fd", mountPoint, pid))
	}

	return errors.Join(err1, err2)
}

func collectProcessStatMetrics(pm *processMetrics, fn string) error {
	raw, err := readFileNoStat(fn)
	if err != nil {
		return err
	}
	str := string(raw)

	i := strings.LastIndexByte(str, ')')
	if i < 0 {
		return fmt.Errorf("invalid %s: no right paren in %q", fn, str)
	}
	str = strings.TrimSpace(str[i+1:])

	var (
		dummyStr       string
		dummyInt       int64
		dummyUint      uint64
		utime, stime   uint64
		vsize, rss     uint64
		minFlt, majFlt uint64
	)

	// See man 5 proc.
	_, err = fmt.Sscan(str,
		&dummyStr,
		&dummyInt, &dummyInt, &dummyInt, &dummyInt, &dummyInt,
		&dummyUint, // flags
		&minFlt, &dummyUint, &majFlt, &dummyUint,
		&utime, &stime, &dummyUint, &dummyUint,
		&dummyInt, &dummyInt,
		&pm.Threads,
		&dummyInt,
		&dummyUint, // starttime
		&vsize, &rss, &dummyUint,
		&dummyUint, &dummyUint, &dummyUint, &dummyUint,
		&dummyUint, &dummyUint, &dummyUint, &dummyUint,
		&dummyUint, &dummyUint, &dummyUint, &dummyUint,
		&dummyInt,
		&dummyInt, // processor
		&dummyUint,
		&dummyUint,
		&dummyUint, // delayacct_blkio_ticks
	)
	if err != nil {
		return fmt.Errorf("invalid %s: %s in %q", fn, err, str)
	}

	pm.CPUTime = float64(utime+stime) / userHZ
	pm.UserTime = float64(utime) / userHZ
	pm.KernelTime = float64(stime) / userHZ
	pm.VirtualMemory = vsize
	pm.ResidentMemory = rss * uint64(os.Getpagesize())
	return nil
}

func collectProcessFdMetricsFast(pm *processMetrics, fn string) bool {
	stat, err := os.Stat(fn)
	if err != nil {
		return false
	}

	size := stat.Size()
	if size > 0 {
		pm.OpenFiles = size
		return true
	}
	return false
}

func collectProcessFdMetricsSlow(pm *processMetrics, fn string) error {
	d, err := os.Open(fn)
	if err != nil {
		return err
	}
	defer d.Close()

	names, err := d.Readdirnames(-1)
	if err != nil {
		return fmt.Errorf("cannot list dir %s: %w", fn, err)
	}
	pm.OpenFiles = int64(len(names))
	return nil
}
