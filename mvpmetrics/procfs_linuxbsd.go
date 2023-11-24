//go:build freebsd || linux

package mvpmetrics

import "syscall"

func isRealProcFS(mountPoint string) bool {
	const PROC_SUPER_MAGIC = 0x9fa0
	var stat syscall.Statfs_t
	err := syscall.Statfs(mountPoint, &stat)
	if err != nil {
		return false
	}
	return stat.Type == PROC_SUPER_MAGIC
}
