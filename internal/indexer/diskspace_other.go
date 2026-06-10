//go:build !unix

package indexer

// availableDiskSpace is not implemented on this platform. Returning 0 with a
// nil error makes checkDiskSpace skip the pre-flight check rather than fail.
func availableDiskSpace(dir string) (uint64, error) {
	return 0, nil
}
