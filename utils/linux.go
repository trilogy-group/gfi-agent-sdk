//go:build linux

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

	"github.com/vishvananda/netlink"
)

func MakeCertificateTrusted(path string) error {
	return nil
}

func GetDefaultInterface() (*net.Interface, error) {
	// Retrieve the default route information from the kernel
	routes, err := netlink.RouteList(nil, netlink.FAMILY_V4)
	if err != nil {
		return nil, err
	}

	for _, route := range routes {
		if route.Dst == nil && route.Gw != nil {
			// Retrieve the interface associated with the default route
			iface, err := net.InterfaceByIndex(route.LinkIndex)
			if err != nil {
				return nil, err
			}
			return iface, nil
		}
	}

	return nil, fmt.Errorf("default interface not found")
}
