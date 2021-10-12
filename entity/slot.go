package entity

import "reflect"

type HasSlotKeys interface {
	SlotKeys() (any, error)
}

var hasSlotKeysType = reflect.TypeOf((*HasSlotKeys)(nil)).Elem()
