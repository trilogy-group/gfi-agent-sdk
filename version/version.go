/**
 * This software and associated documentation files (the “Software”),
 * including GFI AppManager, is the property of GFI USA, LLC and its affiliates.
 * No part of the Software may be copied, modified, distributed, sold, or otherwise
 * used except as expressly permitted by the terms of the software license agreement.
 */

package version

import "fmt"

var GitCommit string
var Major string
var Minor string
var Patch string

func Short() string {
	return fmt.Sprintf("%s.%s", Major, Minor)
}

func Long() string {
	return fmt.Sprintf("%s.%s.%s", Major, Minor, Patch)
}
