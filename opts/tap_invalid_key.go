// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package opts

type TapInvalidKey func(Key)

func (_ TapInvalidKey) IsCleanIndexOption() {}
