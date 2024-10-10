/**
 * This software and associated documentation files (the “Software”),
 * including GFI AppManager, is the property of GFI USA, LLC and its affiliates.
 * No part of the Software may be copied, modified, distributed, sold, or otherwise
 * used except as expressly permitted by the terms of the software license agreement.
 */

package utils

var GetRegStringValue func(registryKey string, valueName string) (string, error)
var SetRegStringValue func(registryKey string, valueName string, value string) error
