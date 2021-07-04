// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package store

type WriteOption interface {
	IsWriteOption()
}

type TapWriteResult func(WriteResult)

func (_ TapWriteResult) IsSaveOption() {}

func (_ TapWriteResult) IsWriteOption() {}
