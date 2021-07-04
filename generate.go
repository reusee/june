// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package ling

//go:generate go build  -o gen-api.exe generate/api.go
//go:generate ./gen-api.exe

//go:generate go build -o gen-test.exe generate/test.go
//go:generate ./gen-test.exe

//go:generate go build -o gen-file-header.exe generate/header.go
//go:generate ./gen-file-header.exe
