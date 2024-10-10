//go:build linux
// +build linux

/**
 * This software and associated documentation files (the “Software”),
 * including GFI AppManager, is the property of GFI USA, LLC and its affiliates.
 * No part of the Software may be copied, modified, distributed, sold, or otherwise
 * used except as expressly permitted by the terms of the software license agreement.
 */

package constants

const (
	GFIAgentInstallationDir      = "/usr/local/gfiagent"
	GFIAgentDataDir              = "/var/gfiagent"
	GFIAgentLogDir               = "/var/logs/gfiagent"
	KerioConnectServerConfigPath = "mailserver/mailserver.cfg"
	GFILanguardAPIConfigPath     = "/restapi.cfg"
	KerioControlServerConfigPath = "/winroute.cfg"
)