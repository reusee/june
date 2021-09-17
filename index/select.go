// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package index

import (
	"context"
	"errors"
	"fmt"
	"io"
	"reflect"

	"github.com/reusee/pp"
	"github.com/reusee/sb"
)

type SelectOption interface {
	IsSelectOption()
}

type IterOption interface {
	IsIterOption()
}

type SelectOptions []SelectOption

func (_ SelectOptions) IsSelectOption() {}

type IterOptions []IterOption

func (_ IterOptions) IsSelectOption() {}

func (_ IterOptions) IsIterOption() {}

type Lower Entry

func (_ Lower) IsSelectOption() {}

func (_ Lower) IsIterOption() {}

type Upper Entry

func (_ Upper) IsSelectOption() {}

func (_ Upper) IsIterOption() {}

var MatchEntry = NewEntry

var MatchPreEntry = NewPreEntry

type Exact Entry

func (_ Exact) IsSelectOption() {}

func (_ Exact) IsIterOption() {}

type Limit int

func (_ Limit) IsSelectOption() {}

func (_ Limit) IsIterOption() {}

type Offset int

func (_ Offset) IsSelectOption() {}

func (_ Offset) IsIterOption() {}

type Where func(sb.Proc) bool

func (_ Where) IsSelectOption() {}

func (_ Where) IsIterOption() {}

type AssignCount [1]*int

func (_ AssignCount) IsSelectOption() {}

func (_ AssignCount) IsIterOption() {}

func Count(target *int) AssignCount {
	return AssignCount{target}
}

type Sink func() sb.Sink

func (_ Sink) IsSelectOption() {}

func Unmarshal(fn any) Sink {
	return func() sb.Sink {
		return sb.Unmarshal(fn)
	}
}

func Tap(fn any) Sink {
	argTypes := []reflect.Type{
		// idx string
		reflect.TypeOf((*string)(nil)).Elem(),
	}
	fnValue := reflect.ValueOf(fn)
	fnType := fnValue.Type()
	for i := 0; i < fnType.NumIn(); i++ {
		argTypes = append(argTypes, fnType.In(i))
	}
	return func() sb.Sink {
		return sb.Unmarshal(reflect.MakeFunc(
			reflect.FuncOf(
				argTypes,
				[]reflect.Type{},
				false,
			),
			func(args []reflect.Value) []reflect.Value {
				// only tuple
				fnValue.Call(args[1:])
				return nil
			},
		).Interface())
	}
}

func TapPre(fn any) Sink {
	argTypes := []reflect.Type{
		// Key
		reflect.TypeOf((*Key)(nil)).Elem(),
		// idx string
		reflect.TypeOf((*string)(nil)).Elem(),
	}
	fnValue := reflect.ValueOf(fn)
	fnType := fnValue.Type()
	for i := 0; i < fnType.NumIn(); i++ {
		argTypes = append(argTypes, fnType.In(i))
	}
	return func() sb.Sink {
		return sb.Unmarshal(reflect.MakeFunc(
			reflect.FuncOf(
				argTypes,
				[]reflect.Type{},
				false,
			),
			func(args []reflect.Value) []reflect.Value {
				// only tuple
				fnValue.Call(args[2:])
				return nil
			},
		).Interface())
	}
}

type TapTokens func(sb.Tokens)

func (_ TapTokens) IsSelectOption() {}

type TapEntry func(Entry)

func (_ TapEntry) IsSelectOption() {}

type WithCtx [1]context.Context

func (_ WithCtx) IsSelectOption() {}

var Cancel = errors.New("cancel")

func Iter(
	options ...IterOption,
) (
	fn func(
		index Index,
	) (
		ProcSrc,
		io.Closer,
		error,
	),
) {

	fn = func(
		index Index,
	) (
		iter ProcSrc,
		closer io.Closer,
		err error,
	) {
		defer he(&err)

		var lower, upper, match, exact *Entry
		var matchPre *PreEntry
		var order = Asc
		var assignCount []AssignCount
		var offset int
		var limit *int
		var where []Where

		var handleOption func(option IterOption)
		handleOption = func(option IterOption) {
			switch option := option.(type) {

			case Lower:
				entry := Entry(option)
				lower = &entry
			case Upper:
				entry := Entry(option)
				upper = &entry
			case Order:
				order = option

			case Entry:
				match = &option

			case PreEntry:
				matchPre = &option

			case AssignCount:
				assignCount = append(assignCount, option)

			case Limit:
				i := int(option)
				limit = &i
			case Offset:
				offset = int(option)

			case Exact:
				entry := Entry(option)
				exact = &entry
				lower = &entry
				order = Asc
				one := 1
				limit = &one

			case IterOptions:
				for _, opt := range option {
					handleOption(opt)
				}

			case Where:
				where = append(where, option)

			default:
				panic(fmt.Errorf("unknown option: %T", option))
			}
		}

		// options
		for _, option := range options {
			handleOption(option)
		}

		// match
		if match != nil {
			lower = &Entry{
				Type:  match.Type,
				Tuple: append(match.Tuple[:0:0], match.Tuple...),
			}
			lower.Tuple = append(lower.Tuple, sb.Min)
			upper = &Entry{
				Type:  match.Type,
				Tuple: append(match.Tuple[:0:0], match.Tuple...),
			}
			upper.Tuple = append(upper.Tuple, sb.Max)
		}

		// exact
		var exactTokens sb.Tokens
		if exact != nil {
			exactTokens, err = sb.TokensFromProc(
				sb.Marshal(*exact),
			)
			ce(err)
		}

		var lowerTokens, upperTokens *sb.Tokens

		if lower != nil {
			var tokens sb.Tokens
			ce(sb.Copy(
				sb.Marshal(*lower),
				sb.CollectTokens(&tokens),
			))
			lowerTokens = &tokens
		}
		if upper != nil {
			var tokens sb.Tokens
			ce(sb.Copy(
				sb.Marshal(*upper),
				sb.CollectTokens(&tokens),
			))
			upperTokens = &tokens
		}

		// match pre
		if matchPre != nil {
			var lTokens, uTokens sb.Tokens
			ce(sb.Copy(
				sb.Marshal(PreEntry{
					Key:   matchPre.Key,
					Type:  matchPre.Type,
					Tuple: append(matchPre.Tuple, sb.Min),
				}),
				sb.CollectTokens(&lTokens),
			))
			lowerTokens = &lTokens
			ce(sb.Copy(
				sb.Marshal(PreEntry{
					Key:   matchPre.Key,
					Type:  matchPre.Type,
					Tuple: append(matchPre.Tuple, sb.Max),
				}),
				sb.CollectTokens(&uTokens),
			))
			upperTokens = &uTokens
		}

		// iter
		iter, closer, err = index.Iter(
			lowerTokens,
			upperTokens,
			order,
		)
		ce(err)

		// exact iter
		if len(exactTokens) > 0 {
			iter = pp.MapSrc(
				iter,
				func(proc sb.Proc) sb.Proc {
					var tokens sb.Tokens
					res := sb.MustCompare(
						sb.Tee(
							proc,
							sb.CollectTokens(&tokens),
						),
						exactTokens.Iter(),
					)
					if res == 0 {
						return tokens.Iter()
					}
					return nil
				},
				nil,
			)
		}

		// where
		for _, fn := range where {
			fn := fn
			iter = pp.MapSrc(
				iter,
				func(proc sb.Proc) sb.Proc {
					var tokens sb.Tokens
					ce(sb.Copy(
						proc,
						sb.CollectTokens(&tokens),
					))
					if fn(tokens.Iter()) {
						return tokens.Iter()
					}
					return nil
				},
				nil,
			)
		}

		// offset
		if offset > 0 {
			iter = pp.SkipSrc(iter, offset, nil)
		}

		// limit
		if limit != nil {
			iter = pp.CapSrc(iter, *limit, nil)
		}

		// assign count
		for _, opt := range assignCount {
			iter = pp.CountSrc(opt[0], iter, nil)
		}

		return
	}
	return
}

func Select(
	index Index,
	options ...SelectOption,
) (
	err error,
) {
	defer he(&err)

	var ctxs []context.Context
	var entryFuncs []func(Entry)
	var tokensFuncs []func(sb.Tokens)
	var iterOptions []IterOption

	var handleOption func(SelectOption)
	handleOption = func(option SelectOption) {
		switch option := option.(type) {

		case TapTokens:
			tokensFuncs = append(tokensFuncs, func(tokens sb.Tokens) {
				option(tokens)
			})

		case Sink:
			tokensFuncs = append(tokensFuncs, func(tokens sb.Tokens) {
				ce(sb.Copy(
					tokens.Iter(),
					option(),
				))
			})

		case TapEntry:
			entryFuncs = append(entryFuncs, func(entry Entry) {
				option(entry)
			})

		case TapKey:
			entryFuncs = append(entryFuncs, func(entry Entry) {
				option(*entry.Key)
			})

		case WithCtx:
			ctxs = append(ctxs, option[0])

		case IterOption:
			iterOptions = append(iterOptions, option)

		case SelectOptions:
			for _, opt := range option {
				handleOption(opt)
			}

		default:
			panic(fmt.Errorf("unknown option: %T", option))
		}
	}

	for _, option := range options {
		handleOption(option)
	}

	iter, closer, err := Iter(iterOptions...)(index)
	ce(err)
	defer closer.Close()

	var sinks []sb.Sink
	ce(pp.Copy(
		iter,
		pp.Tap[sb.Proc, ProcSink](func(proc sb.Proc) (err error) {
			defer he(&err)

			// check context
			for _, ctx := range ctxs {
				select {
				case <-ctx.Done():
					return Cancel
				default:
				}
			}

			// prepare sinks
			sinks = sinks[:0]

			// need Entry
			var entry Entry
			if len(entryFuncs) > 0 {
				sinks = append(sinks, sb.Unmarshal(&entry))
			}

			// need tokens
			var tokens sb.Tokens
			if len(tokensFuncs) > 0 {
				sinks = append(sinks, sb.CollectTokens(&tokens))
			}

			// copy stream
			if len(sinks) > 0 {
				ce(sb.Copy(proc, sinks...))
			}

			// tap funcs
			for _, fn := range entryFuncs {
				fn(entry)
			}
			for _, fn := range tokensFuncs {
				fn(tokens)
			}

			return nil
		}),
	))

	return nil
}
