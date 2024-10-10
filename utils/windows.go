//go:build windows
// +build windows

/**
 * This software and associated documentation files (the “Software”),
 * including GFI AppManager, is the property of GFI USA, LLC and its affiliates.
 * No part of the Software may be copied, modified, distributed, sold, or otherwise
 * used except as expressly permitted by the terms of the software license agreement.
 */

package utils

import (
	"fmt"
	"net"
	"strconv"
	"syscall"

	"golang.org/x/sys/windows/registry"
)

func KillProcess(pid int) error {
	cli := NewCLI("taskkill", []string{"/PID", strconv.Itoa(pid), "/F"})
	if stdout, stderr, err := cli.Run(); err != nil {
		return fmt.Errorf("failed to kill process with PID %d. stdout: %s, stderr: %s, error: %s", pid, stdout, stderr, err.Error())
	}
	return nil
}

func GetWindowsRegStringValue(registryKey string, valueName string) (string, error) {
	var access uint32 = registry.QUERY_VALUE
	regKey, err := registry.OpenKey(registry.LOCAL_MACHINE, registryKey, access)
	if err != nil {
		return "", err
	}
	val, _, err := regKey.GetStringValue(valueName)
	if err != nil {
		return "", err
	}
	return val, nil
}

func SetWindowsRegStringValue(registryKey string, valueName string, value string) error {
	var access uint32 = registry.SET_VALUE
	regKey, err := registry.OpenKey(registry.LOCAL_MACHINE, registryKey, access)
	if err != nil {
		return err
	}
	err = regKey.SetStringValue(valueName, value)
	if err != nil {
		return err
	}
	return nil
}

func MakeCertificateTrusted(path string) error {
	return nil
}

func GetDefaultInterface() (*net.Interface, error) {
	return nil, fmt.Errorf("not supported on windows")
}

func GetDetachedStartAttributes() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{
	}
}
func init() {
	GetRegStringValue = GetWindowsRegStringValue
	SetRegStringValue = SetWindowsRegStringValue
}
