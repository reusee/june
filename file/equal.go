// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package file

import (
	"bytes"
	"fmt"
	"io"
	"reflect"

	"github.com/reusee/june/key"
	"github.com/reusee/sb"
)

type Equal func(
	a, b Src,
	fn func(any, any, string),
) (bool, error)

func (_ Def) Equal(
	scope Scope,
	newHashState key.NewHashState,
) (equal Equal) {
	equal = func(
		a, b Src,
		fn func(any, any, string),
	) (ret bool, err error) {
		defer he(&err)

	l0:

	l1:
		valueA, err := Get(&a)
		if err != nil {
			return false, err
		}
		if t, ok := (*valueA).(PackThunk); ok {
			t.Expand(true)
			goto l1
		} else if t, ok := (*valueA).(FileInfoThunk); ok {
			i := any(t.FileInfo)
			valueA = &i
			t.Expand(true)
		}

	l2:
		valueB, err := Get(&b)
		if err != nil {
			return false, err
		}
		if t, ok := (*valueB).(PackThunk); ok {
			t.Expand(true)
			goto l2
		} else if t, ok := (*valueB).(FileInfoThunk); ok {
			i := any(t.FileInfo)
			valueB = &i
			t.Expand(true)
		}

		if valueA == nil && valueB == nil {
			return true, nil
		}

		cb := func(reason string, args ...any) {
			if fn != nil {
				fn(valueA, valueB, fmt.Sprintf(reason, args...))
			}
		}

		if a, b := reflect.TypeOf(valueA), reflect.TypeOf(valueB); a != b {
			cb("type not match: %v %v", a, b)
			return false, nil
		}

		if aInfo, ok := (*valueA).(FileInfo); ok {
			bInfo := (*valueB).(FileInfo)
			if a, b := aInfo.GetIsDir(scope), bInfo.GetIsDir(scope); a != b {
				cb("dir not match: %v %v", a, b)
				return false, nil
			}
			if a, b := aInfo.GetName(scope), bInfo.GetName(scope); a != b {
				cb("name not match: %v %v", a, b)
				return false, nil
			}
			if a, b := aInfo.GetSize(scope), bInfo.GetSize(scope); a != b {
				cb("size not match: %v %v", a, b)
				return false, nil
			}
			if a, b := aInfo.GetMode(scope), bInfo.GetMode(scope); a != b {
				cb("mode not match: %v %v", a, b)
				return false, nil
			}
			if !aInfo.GetModTime(scope).Equal(bInfo.GetModTime(scope)) {
				pt("%v\n", aInfo.GetModTime(scope))
				pt("%v\n", bInfo.GetModTime(scope))
				cb("mod time not match: %v %v", a, b)
				return false, nil
			}

			if !aInfo.GetIsDir(scope) {
				var hash1 []byte
				ce(aInfo.WithReader(scope, func(r io.Reader) (err error) {
					defer he(&err)
					data, err := io.ReadAll(r)
					ce(err)
					ce(sb.Copy(
						sb.Marshal(data),
						sb.Hash(newHashState, &hash1, nil),
					))
					return
				}))
				var hash2 []byte
				ce(bInfo.WithReader(scope, func(r io.Reader) (err error) {
					defer he(&err)
					data, err := io.ReadAll(r)
					ce(err)
					ce(sb.Copy(
						sb.Marshal(data),
						sb.Hash(newHashState, &hash2, nil),
					))
					return
				}))
				if !bytes.Equal(hash1, hash2) {
					cb("content not match: %v %v", a, b)
					return false, nil
				}
			}

		} else {
			return false, fmt.Errorf("unknown type %#v", valueA)
		}

		goto l0
	}

	return equal
}
