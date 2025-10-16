package main

import (
	"fmt"
	"runtime"

	"github.com/adewale/olsen/internal/indexer"
)

func versionCommand() error {
	fmt.Println("Olsen Photo Indexer")
	fmt.Printf("Go version: %s\n", runtime.Version())
	fmt.Printf("OS/Arch: %s/%s\n", runtime.GOOS, runtime.GOARCH)

	if indexer.IsRawSupported() {
		fmt.Printf("RAW support: enabled (%s)\n", indexer.LibRawImpl)
	} else {
		fmt.Println("RAW support: disabled")
	}

	return nil
}
