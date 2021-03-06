// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package entity

import "reflect"

type HasSlotKeys interface {
	SlotKeys() (any, error)
}

var hasSlotKeysType = reflect.TypeOf((*HasSlotKeys)(nil)).Elem()
