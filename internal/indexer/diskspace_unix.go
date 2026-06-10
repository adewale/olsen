//go:build unix

package indexer

import "syscall"

// availableDiskSpace returns the number of bytes available to the current
// user on the filesystem containing dir.
func availableDiskSpace(dir string) (uint64, error) {
	var stat syscall.Statfs_t
	if err := syscall.Statfs(dir, &stat); err != nil {
		return 0, err
	}
	return stat.Bavail * uint64(stat.Bsize), nil
}
