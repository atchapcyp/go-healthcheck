package main

import (
	"math"
	"syscall"
)

func FileDescriptorSize() int {
	var rLimit syscall.Rlimit
	err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rLimit)
	if err != nil {
		return -1
	}
	if (rLimit.Cur) > uint64(math.MaxInt32) {
		return math.MaxInt32
	}
	return int(rLimit.Cur)
}
