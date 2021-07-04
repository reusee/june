// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package opts

import "bytes"

type NewBytesBuffer func() *bytes.Buffer

func (n NewBytesBuffer) IsCodecOption() {}

func (n NewBytesBuffer) IsWriteOption() {}
