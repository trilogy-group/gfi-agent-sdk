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
	"os/exec"
	"regexp"
	"runtime"
	"strings"

	"github.com/trilogy-group/gfi-agent-sdk/logger"
)

const urlPattern = `^(https?://)([^/:]+)(:\d+)?(.*)$`

var urlPatternRegex = regexp.MustCompile(urlPattern)

func GetAdminUiUrl(endpoint string, prefix string) string {
	uiUrl := ""

	ip, err := GetDefaultIP()
	if err != nil {
		logger.Logger.Errorf("Failed to get default IP: %s", err)
	} else {
		format := fmt.Sprintf("${1}%s${3}/%s", ip, prefix)
		uiUrl = urlPatternRegex.ReplaceAllString(endpoint, format)
	}

	return uiUrl
}

func GetDefaultWindowsIP() (string, error) {
	out, err := exec.Command("route", "print", "0.0.0.0").Output()
	if err != nil {
		return "", fmt.Errorf("failed to get windows routing table: %s", err)
	}

	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) >= 4 && fields[0] == "0.0.0.0" {
			ip := fields[3]
			return ip, nil
		}
	}

	return "", fmt.Errorf("failed to get default windows route")
}

func GetDefaultIP() (string, error) {
	if runtime.GOOS == "windows" {
		ip, err := GetDefaultWindowsIP()
		if err == nil {
			return ip, nil
		} else {
			logger.Logger.Infof("Unable to identity default windows network interface: '%s'. Fallback to the default method.", err)
		}
	}

	iface, err := GetDefaultInterface()
	if err != nil {
		logger.Logger.Infof("Failed to found default interface, trying fallback")
		iface, err = GetDefaultInterfaceFallback()
		if err != nil {
			return "", fmt.Errorf("failed to found default interface: %s", err)
		}
	}

	ip, err := GetDefaultInterfaceIP(iface)
	if err == nil {
		return ip, nil
	}

	return "", fmt.Errorf("failed to find default network interface")
}

func GetDefaultInterfaceFallback() (*net.Interface, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return nil, fmt.Errorf("failed to get interfaces: %s", err)
	}
	for _, iface := range interfaces {
		if (iface.Flags&net.FlagUp) != 0 && (iface.Flags&net.FlagLoopback) == 0 {
			return &iface, nil
		}
	}

	return nil, fmt.Errorf("failed to found up and not loopback interface")
}

func GetDefaultInterfaceIP(iface *net.Interface) (string, error) {
	addrs, err := iface.Addrs()
	if err != nil {
		return "", fmt.Errorf("failed to get network interface addresses: %s", err)
	}

	for _, addr := range addrs {
		ipNet, ok := addr.(*net.IPNet)
		if ok && !ipNet.IP.IsLoopback() && ipNet.IP.To4() != nil {
			ip := ipNet.IP.String()
			if ip != "" && ip != "127.0.0.1" {
				return ip, nil
			}
		}
	}

	return "", fmt.Errorf("failed to found not loopback interface IPv4 address")
}
