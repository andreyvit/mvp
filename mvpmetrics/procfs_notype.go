//go:build !freebsd && !linux

package mvpmetrics

func isRealProcFS(mountPoint string) bool {
	return false
}
