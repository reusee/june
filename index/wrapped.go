// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package index

import (
	"fmt"
	"reflect"

	"github.com/reusee/e4"
)

func (_ Def) Index(
	manager IndexManager,
	storeID StoreID,
) Index {

	index, err := manager.IndexFor(storeID)
	ce(err)

	return wrapped{
		Index: index,
	}

}

type wrapped struct {
	Index
}

var keyType = reflect.TypeOf((*Key)(nil)).Elem()

func (w wrapped) Save(entry Entry, options ...SaveOption) (err error) {
	defer he(&err)

	if _, ok := entry.Type.(idxUnknown); ok {
		// skip unknown
		return nil
	}

	var tapEntry []TapEntry
	for _, option := range options {
		switch option := option.(type) {
		case TapEntry:
			tapEntry = append(tapEntry, option)
		default:
			panic(fmt.Errorf("unknown option: %T", option))
		}
	}

	if entry.Type == nil {
		return we.With(
			e4.NewInfo("entry type is nil: %v", entry),
		)(ErrInvalidEntry)
	}

	t := reflect.TypeOf(entry.Type)
	v, ok := specsByType.Load(t)
	if !ok {
		return we.With(
			e4.NewInfo("unknown index type: %T", entry.Type),
		)(ErrInvalidEntry)
	}
	spec := v.(Spec)

	if entry.Key == nil {
		return we.With(
			e4.NewInfo("entry has no Key"),
		)(ErrInvalidEntry)
	}

	if len(entry.Tuple) != len(spec.Fields) {
		return we.With(
			e4.NewInfo(
				"%s is expecting %d elements, but got %d",
				spec.Name,
				len(spec.Fields),
				len(entry.Tuple),
			),
		)(ErrInvalidEntry)
	}
	for i, typ := range spec.Fields {
		if argType := reflect.TypeOf(entry.Tuple[i]); argType != typ {
			if argType.AssignableTo(typ) {
				value := reflect.New(typ).Elem()
				value.Set(reflect.ValueOf(entry.Tuple[i]))
				entry.Tuple[i] = value.Interface()
			} else if argType.ConvertibleTo(typ) {
				entry.Tuple[i] = reflect.ValueOf(entry.Tuple[i]).Convert(typ).Interface()
			} else {
				return we.With(
					e4.NewInfo("param %d of %s should be %v", i, spec.Name, typ.String()),
				)(ErrInvalidEntry)
			}
		}
	}

	ce(w.Index.Save(entry))

	for _, tap := range tapEntry {
		tap(entry)
	}

	return nil
}
