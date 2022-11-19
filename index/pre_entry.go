// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package index

import (
	"fmt"
	"io"
	"reflect"

	"github.com/reusee/sb"
)

type PreEntry struct {
	Key   Key
	Type  any
	Tuple sb.Tuple
}

func NewPreEntry(key Key, indexType any, args ...any) PreEntry {
	var types []reflect.Type
	v, ok := specsByType.Load(reflect.TypeOf(indexType))
	if ok {
		types = v.(Spec).Fields
	}
	entry := PreEntry{
		Key:  key,
		Type: indexType,
	}
	if len(types) > 0 {
		if len(args) <= len(types) {
			entry.Tuple = args
		} else {
			panic(fmt.Errorf("bad argument count: %T", indexType))
		}
	} else {
		entry.Tuple = args
	}
	return entry
}

func (e PreEntry) Clone() PreEntry {
	newEntry := PreEntry{
		Key:  e.Key,
		Type: e.Type,
	}
	if len(e.Tuple) > 0 {
		tuple := make(sb.Tuple, len(e.Tuple))
		copy(tuple, e.Tuple)
		newEntry.Tuple = tuple
	}
	return newEntry
}

var _ sb.SBMarshaler = PreEntry{}

func (e PreEntry) MarshalSB(ctx sb.Ctx, cont sb.Proc) sb.Proc {
	var name string
	v, ok := specsByType.Load(reflect.TypeOf(e.Type))
	if !ok {
		name = nameOfIdxUnknown
	} else {
		name = v.(Spec).Name
	}
	tuple := append(sb.Tuple{e.Key, name}, e.Tuple...)
	return ctx.Marshal(ctx,
		reflect.ValueOf(tuple),
		cont,
	)
}

var _ sb.SBUnmarshaler = new(PreEntry)

func (e *PreEntry) UnmarshalSB(ctx sb.Ctx, cont sb.Sink) sb.Sink {
	var name string
	var types []reflect.Type
	var entry PreEntry

	var unmarshalTyped sb.Sink
	unmarshalTyped = func(token *sb.Token) (sb.Sink, error) {
		if token == nil {
			return nil, we(io.ErrUnexpectedEOF)
		}
		if token.Kind == sb.KindTupleEnd {
			*e = entry
			return cont, nil
		}
		value := reflect.New(types[0])
		types = types[1:]
		return ctx.Unmarshal(
			ctx,
			value,
			func(token *sb.Token) (sb.Sink, error) {
				entry.Tuple = append(entry.Tuple, value.Elem().Interface())
				return unmarshalTyped(token)
			},
		)(token)
	}

	var unmarshalUnknown sb.Sink
	unmarshalUnknown = func(token *sb.Token) (sb.Sink, error) {
		if token == nil {
			return nil, we(io.ErrUnexpectedEOF)
		}
		if token.Kind == sb.KindTupleEnd {
			*e = entry
			return cont, nil
		}
		var value any
		return ctx.Unmarshal(
			ctx,
			reflect.ValueOf(&value),
			func(token *sb.Token) (sb.Sink, error) {
				entry.Tuple = append(entry.Tuple, value)
				return unmarshalUnknown(token)
			},
		)(token)
	}

	unmarshalName := ctx.Unmarshal(
		ctx,
		reflect.ValueOf(&name),
		func(token *sb.Token) (sb.Sink, error) {
			var spec Spec
			v, ok := specsByName.Load(name)
			if !ok {
				entry.Type = idxUnknown(name)
				return unmarshalUnknown(token)
			} else {
				spec = v.(Spec)
			}
			entry.Type = reflect.New(spec.Type).Elem().Interface()
			types = spec.Fields
			return unmarshalTyped(token)
		},
	)

	unmarshalKey := ctx.Unmarshal(
		ctx,
		reflect.ValueOf(&entry.Key),
		unmarshalName,
	)

	return sb.ExpectKind(
		ctx,
		sb.KindTuple,
		unmarshalKey,
	)

}

func (PreEntry) IsSelectOption() {}

func (PreEntry) IsIterOption() {}
