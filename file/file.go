// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package file

import "github.com/reusee/june/filebase"

type (
	File = filebase.File
	Sub  = filebase.Sub
	Subs = filebase.Subs
	Pack = filebase.Pack

	NewContentReader = filebase.NewContentReader
	ToContents       = filebase.ToContents
	WriteContents    = filebase.WriteContents
	Content          = filebase.Content

	ChunkThreshold = filebase.ChunkThreshold
	MaxChunkSize   = filebase.MaxChunkSize
)

var _ FileLike = new(File)
