// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package index

import (
	"fmt"
	"reflect"

	"github.com/reusee/sb"
)

type Entry struct {
	Type  any
	Tuple sb.Tuple
	Key   *Key
	Path  *[]Key
}

func NewEntry(indexType any, args ...any) Entry {
	var types []reflect.Type
	v, ok := specsByType.Load(reflect.TypeOf(indexType))
	if ok {
		types = v.(Spec).Fields
	}
	entry := Entry{
		Type: indexType,
	}
	if len(types) > 0 {
		if len(args) <= len(types) {
			entry.Tuple = args
		} else if len(args) == len(types)+1 {
			entry.Tuple = args[:len(args)-1]
			last := args[len(args)-1]
			if last != nil {
				switch elem := last.(type) {
				case Key:
					entry.Key = &elem
				case []Key:
					entry.Path = &elem
				default:
					panic(fmt.Errorf("bad last argument: %T", elem))
				}
			}
		} else {
			panic(fmt.Errorf("bad argument count: %T", indexType))
		}
	} else {
		entry.Tuple = args
	}
	return entry
}

func (e Entry) Clone() Entry {
	newEntry := Entry{
		Type: e.Type,
		Key:  e.Key,
	}
	if len(e.Tuple) > 0 {
		tuple := make(sb.Tuple, len(e.Tuple))
		copy(tuple, e.Tuple)
		newEntry.Tuple = tuple
	}
	if e.Path != nil {
		path := make([]Key, len(*e.Path))
		copy(path, *e.Path)
		newEntry.Path = &path
	}
	return newEntry
}

var _ sb.SBMarshaler = Entry{}

func (e Entry) MarshalSB(ctx sb.Ctx, cont sb.Proc) sb.Proc {
	var name string
	v, ok := specsByType.Load(reflect.TypeOf(e.Type))
	if !ok {
		name = nameOfIdxUnknown
	} else {
		name = v.(Spec).Name
	}
	tuple := append(sb.Tuple{name}, e.Tuple...)
	if e.Key != nil {
		tuple = append(tuple, *e.Key)
	} else if e.Path != nil {
		tuple = append(tuple, *e.Path)
	}
	return ctx.Marshal(ctx,
		reflect.ValueOf(tuple),
		cont,
	)
}

var _ sb.SBUnmarshaler = new(Entry)

func (e *Entry) UnmarshalSB(ctx sb.Ctx, cont sb.Sink) sb.Sink {
	var name string
	var types []reflect.Type
	var entry Entry

	var unmarshalTyped sb.Sink
	unmarshalTyped = func(token *sb.Token) (sb.Sink, error) {
		if token == nil {
			return nil, we()(sb.ExpectingValue)
		}
		if token.Kind == sb.KindTupleEnd {
			*e = entry
			return cont, nil
		}
		if len(types) == 0 {
			// unmarshal last Key
			var key Key
			var path []Key
			return sb.AltSink(
				ctx.Unmarshal(
					ctx,
					reflect.ValueOf(&key),
					func(token *sb.Token) (sb.Sink, error) {
						if key.Valid() {
							entry.Key = &key
						}
						return unmarshalTyped(token)
					},
				),
				ctx.Unmarshal(
					ctx,
					reflect.ValueOf(&path),
					func(token *sb.Token) (sb.Sink, error) {
						if len(path) > 0 {
							entry.Path = &path
						}
						return unmarshalTyped(token)
					},
				),
			)(token)
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
			return nil, we()(sb.ExpectingValue)
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

	return sb.ExpectKind(
		ctx,
		sb.KindTuple,
		// index name
		ctx.Unmarshal(
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
				// params
				return unmarshalTyped(token)
			},
		),
	)

}

func (_ Entry) IsSelectOption() {}

func (_ Entry) IsIterOption() {}
