// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package opts

type TapKeySet func(map[Key]struct{})

func (_ TapKeySet) IsSyncOption() {}
