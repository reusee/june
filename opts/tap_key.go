// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package opts

type TapKey func(Key)

func (_ TapKey) IsIndexOption() {}

func (_ TapKey) IsCheckRefOption() {}

func (_ TapKey) IsWriteOption() {}

func (_ TapKey) IsScrubOption() {}

func (_ TapKey) IsCleanIndexOption() {}

func (_ TapKey) IsSelectOption() {}

func (_ TapKey) IsResaveOption() {}

func (_ TapKey) IsSaveSummaryOption() {}
