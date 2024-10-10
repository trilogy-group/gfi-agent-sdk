//go:build darwin

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
)

func MakeCertificateTrusted(path string) error {
	cli := NewCLI("security", []string{"authorizationdb", "write", "com.apple.trust-settings.admin", "allow"})
	_, _, err := cli.Run()
	if err != nil {
		return fmt.Errorf("failed to update certificate trust settings: %s", err)
	}

	cli = NewCLI("security", []string{"add-trusted-cert", "-d", "-r", "trustRoot", "-k", "/Library/Keychains/System.keychain", path})
	_, _, err = cli.Run()
	if err != nil {
		return fmt.Errorf("failed to put local HTTPS certificate to the trusted store: %s", err)
	}
	return nil
}

func GetDefaultInterface() (*net.Interface, error) {
	return nil, fmt.Errorf("not supported on macos")
}
