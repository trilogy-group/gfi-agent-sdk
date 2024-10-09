/**
 * This software and associated documentation files (the “Software”),
 * including GFI AppManager, is the property of GFI USA, LLC and its affiliates.
 * No part of the Software may be copied, modified, distributed, sold, or otherwise
 * used except as expressly permitted by the terms of the software license agreement.
 */

package version_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/trilogy-group/gfi-agent-sdk/version"
)

func TestVersion(t *testing.T) {
	assert := assert.New(t)
	version.Major = "1"
	version.Minor = "2"
	version.Patch = "3"
	assert.Equal("1.2.3", version.Long())
	assert.Equal("1.2", version.Short())
}
