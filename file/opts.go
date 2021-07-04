// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package file

type TapAddFileInfo func(FileInfo)

func (_ TapAddFileInfo) IsUpdateOption() {}

type TapDeleteFileInfo func(FileInfo)

func (_ TapDeleteFileInfo) IsUpdateOption() {}

type TapModifyFileInfo func(FileInfo, FileInfo)

func (_ TapModifyFileInfo) IsUpdateOption() {}

type TapBuildFile func(FileInfo, *File)

func (_ TapBuildFile) IsBuildOption() {}

type TapReadFile func(FileInfo)

func (_ TapReadFile) IsBuildOption() {}

type UseGitIgnore bool

func (_ UseGitIgnore) IsIterDiskFileOption() {}

type SingleDevice bool

func (_ SingleDevice) IsIterDiskFileOption() {}

type PredictExpandFileInfoThunk func(_, _ FileInfoThunk) (bool, error)

func (_ PredictExpandFileInfoThunk) IsZipOption() {}
