/**
 * This software and associated documentation files (the “Software”),
 * including GFI AppManager, is the property of GFI USA, LLC and its affiliates.
 * No part of the Software may be copied, modified, distributed, sold, or otherwise
 * used except as expressly permitted by the terms of the software license agreement.
 */

package types

var Types = []Type{}

type Type interface {
	Name() string
	Synonyms() []string
}

func Register(type_ Type) {
	Types = append(Types, type_)
}
