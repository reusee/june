// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package file

type Unzip func(
	src Src,
	fn func(ZipItem) any,
	cont Src,
) Src

func (_ Def) Unzip() Unzip {
	return func(
		src Src,
		fn func(ZipItem) any,
		cont Src,
	) Src {
		var unzip Src
		unzip = func() (any, Src, error) {
			v, err := src.Next()
			if err != nil {
				return nil, nil, err
			}
			if v == nil {
				return nil, cont, nil
			}
			return fn(v.(ZipItem)), unzip, nil
		}
		return unzip
	}
}

type Reverse func(Src, Src) Src

func (_ Def) Reverse() Reverse {
	return func(src Src, cont Src) Src {
		var rev Src
		rev = func() (any, Src, error) {
			v, err := src.Next()
			if err != nil {
				return nil, nil, err
			}
			if v == nil {
				return nil, cont, nil
			}
			item := v.(ZipItem)
			item.A, item.B = item.B, item.A
			return item, rev, nil
		}
		return rev
	}
}
