//go:build !windows
// +build !windows

/**
 * This software and associated documentation files (the “Software”),
 * including GFI AppManager, is the property of GFI USA, LLC and its affiliates.
 * No part of the Software may be copied, modified, distributed, sold, or otherwise
 * used except as expressly permitted by the terms of the software license agreement.
 */

package utils

import (
	"fmt"
	"os"
	"runtime"
	"syscall"
	"time"

	"github.com/trilogy-group/gfi-agent-sdk/logger"
)

func KillProcess(pid int) error {
	process, err := os.FindProcess(pid)
	if err != nil {
		return fmt.Errorf("failed to kill process with PID %d, error: %s", pid, err.Error())
	}

	// Attempt to kill the process gracefully with SIGKILL
	err = process.Signal(syscall.SIGKILL)
	if err != nil {
		logger.Logger.Errorf("Error killing process with SIGKILL: %s\n", err.Error())

		// Wait for a short duration before attempting SIGTERM
		time.Sleep(1 * time.Second)

		// Attempt to kill the process with SIGTERM
		err = process.Signal(syscall.SIGTERM)
		if err != nil {
			return fmt.Errorf("error killing process with SIGTERM: %s\n", err.Error())
		}
	}
	return nil
}

func GetNonWindowsRegStringValue(registryKey string, valueName string) (string, error) {
	return "", fmt.Errorf("os: %s not supported yet", runtime.GOOS)
}

func GetDetachedStartAttributes() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{
		Setpgid: true,
	}
}

func init() {
	GetRegStringValue = GetNonWindowsRegStringValue
}
