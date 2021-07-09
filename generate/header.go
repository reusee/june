// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

//go:build ignore
// +build ignore

package main

import (
	"bytes"
	"fmt"
	"go/format"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/reusee/e4"
)

var (
	ce = e4.Check
	he = e4.Handle
	pt = fmt.Printf
)

func main() {
	ce(filepath.WalkDir(".", func(path string, entry fs.DirEntry, err error) (retErr error) {
		defer he(&retErr)
		ce(err)

		ext := filepath.Ext(path)
		if !map[string]bool{
			".go": true,
		}[ext] {
			return nil
		}

		content, err := os.ReadFile(path)
		ce(err)
		content = bytes.ReplaceAll(content, []byte("\r\n"), []byte("\n"))
		lines := bytes.Split(content, []byte("\n"))
		if len(lines) == 0 {
			return nil
		}

		skip := 0
		for i, line := range lines {
			if bytes.Contains(line, []byte("+build")) {
				break
			}
			if bytes.HasPrefix(line, []byte("//")) {
				skip = i + 1
			} else {
				break
			}
		}
		lines = lines[skip:]

		lines = append([][]byte{
			[]byte("// Copyright 2021 The June Authors. All rights reserved."),
			[]byte("// Use of this source code is governed by Apache License"),
			[]byte("// that can be found in the LICENSE file."),
			[]byte(""),
		}, lines...)

		formated, err := format.Source(bytes.Join(lines, []byte("\n")))
		ce(err)

		ce(os.WriteFile(path, formated, 0644))

		return
	}))

}
