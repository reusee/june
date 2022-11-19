// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package index

import (
	"context"
	"fmt"
	"io"
	"reflect"
	"sync"

	"github.com/reusee/e5"
	"github.com/reusee/sb"
)

func (Def) Index(
	manager IndexManager,
	store Store,
) Index {

	return &wrapped{
		manager: manager,
		store:   store,
	}

}

type wrapped struct {
	manager IndexManager
	store   Store
	once    sync.Once
	index   Index
}

var keyType = reflect.TypeOf((*Key)(nil)).Elem()

func (w *wrapped) getIndex(ctx context.Context) (idx Index, err error) {
	w.once.Do(func() {
		var id StoreID
		id, err = w.store.ID(ctx)
		if err != nil {
			return
		}
		w.index, err = w.manager.IndexFor(id)
		if err != nil {
			return
		}
	})
	idx = w.index
	return
}

func (w *wrapped) Save(ctx context.Context, entry Entry, options ...SaveOption) (err error) {
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
			e5.Info("entry type is nil: %v", entry),
		)(ErrInvalidEntry)
	}

	t := reflect.TypeOf(entry.Type)
	v, ok := specsByType.Load(t)
	if !ok {
		return we.With(
			e5.Info("unknown index type: %T", entry.Type),
		)(ErrInvalidEntry)
	}
	spec := v.(Spec)

	if entry.Key == nil {
		return we.With(
			e5.Info("entry has no Key"),
		)(ErrInvalidEntry)
	}

	if len(entry.Tuple) != len(spec.Fields) {
		return we.With(
			e5.Info(
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
					e5.Info("param %d of %s should be %v", i, spec.Name, typ.String()),
				)(ErrInvalidEntry)
			}
		}
	}

	index, err := w.getIndex(ctx)
	ce(err)
	ce(index.Save(ctx, entry))

	for _, tap := range tapEntry {
		tap(entry)
	}

	return nil
}

func (w *wrapped) Delete(ctx context.Context, entry Entry) (err error) {
	defer he(&err)
	index, err := w.getIndex(ctx)
	ce(err)
	return index.Delete(ctx, entry)
}

func (w *wrapped) Iter(
	ctx context.Context,
	lower *sb.Tokens, // inclusive in any order
	upper *sb.Tokens, // exclusive in any order
	order Order,
) (Src, io.Closer, error) {
	index, err := w.getIndex(ctx)
	if err != nil {
		return nil, nil, err
	}
	return index.Iter(ctx, lower, upper, order)
}

func (w *wrapped) Name(ctx context.Context) (string, error) {
	index, err := w.getIndex(ctx)
	if err != nil {
		return "", err
	}
	return index.Name(ctx)
}
