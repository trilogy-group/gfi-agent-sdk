/**
 * This software and associated documentation files (the “Software”),
 * including GFI AppManager, is the property of GFI USA, LLC and its affiliates.
 * No part of the Software may be copied, modified, distributed, sold, or otherwise
 * used except as expressly permitted by the terms of the software license agreement.
 */

package utils

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"syscall"
)

var (
	ErrProcessRunning = errors.New("process is running")
	ErrFileStale      = errors.New("pidfile exists but process is not running")
	ErrFileInvalid    = errors.New("pidfile has invalid contents")
)

// RemovePidFile Remove a pidfile
func RemovePidFile(filename string) error {
	return os.RemoveAll(filename)
}

// WritePidFile Write writes a pidfile
func WritePidFile(filename string, pid int) error {
	return os.WriteFile(filename, []byte(fmt.Sprintf("%d\n", pid)), 0644)
}

func GetPidFileContents(filename string) (int, error) {
	contents, err := os.ReadFile(filename)
	if err != nil {
		return 0, err
	}

	pid, err := strconv.Atoi(strings.TrimSpace(string(contents)))
	if err != nil {
		return 0, ErrFileInvalid
	}

	return pid, nil
}

func IsPidRunning(pid int) bool {
	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}

	err = process.Signal(syscall.Signal(0))

	if err != nil && err.Error() == "no such process" {
		return false
	}

	if err != nil && err.Error() == "os: process already finished" {
		return false
	}

	return true
}
