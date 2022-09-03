// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package filebase

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"reflect"
	"time"

	"github.com/reusee/e5"
	"github.com/reusee/sb"
)

type ()

type File struct {
	ModTime      time.Time
	Name         string
	Subs         Subs
	Contents     []Key
	ChunkLengths []int64
	ContentBytes []byte
	Size         int64
	Mode         os.FileMode
	IsDir        bool
}

type Sub struct {
	File *File
	Pack *Pack
}

type Subs []Sub

type Pack struct {
	Min    string
	Max    string
	Size   int64
	Height int
	Key    Key // key of Subs
}

var _ sb.SBMarshaler = File{}

func (f File) MarshalSB(ctx sb.Ctx, cont sb.Proc) sb.Proc {
	return sb.MarshalStruct(ctx.SkipEmpty(), reflect.ValueOf(f), cont)
}

var _ sb.SBMarshaler = Sub{}

func (s Sub) MarshalSB(ctx sb.Ctx, cont sb.Proc) sb.Proc {
	// marshal as tuple
	return ctx.Marshal(ctx, reflect.ValueOf(func() (*File, *Pack) {
		return s.File, s.Pack
	}), cont)
	// old way
	//if s.File != nil {
	//	return ctx.Marshal(ctx, reflect.ValueOf(s.File), cont)
	//}
	//if s.Pack != nil {
	//	return ctx.Marshal(ctx, reflect.ValueOf(s.Pack), cont)
	//}
	//panic("bad sub")
}

var _ sb.SBUnmarshaler = new(Sub)

func (s *Sub) UnmarshalSB(ctx sb.Ctx, cont sb.Sink) sb.Sink {
	ctx.DisallowUnknownStructFields = true
	return func(token *sb.Token) (sb.Sink, error) {
		switch token.Kind {
		case sb.KindTuple:
			// tuple
			return ctx.Unmarshal(ctx, reflect.ValueOf(func(file *File, pack *Pack) {
				s.File = file
				s.Pack = pack
			}), cont)(token)
		case sb.KindObject:
			// old way
			return sb.AltSink(
				ctx.Unmarshal(ctx, reflect.ValueOf(&s.File), cont),
				ctx.Unmarshal(ctx, reflect.ValueOf(&s.Pack), cont),
			)(token)
		}
		return nil, fmt.Errorf("bad kind")
	}
}

func (f *File) GetWeight(scope Scope) int {
	weight := 1
	for _, sub := range f.Subs {
		if sub.File != nil {
			weight += sub.File.GetWeight(scope)
		}
		if sub.Pack != nil {
			weight += 1
		}
	}
	return weight
}

func (f *File) GetIsDir(_ Scope) bool {
	return f.IsDir
}

func (f *File) GetName(_ Scope) string {
	return f.Name
}

func (f *File) GetSize(_ Scope) int64 {
	return f.Size
}

func (f *File) GetMode(_ Scope) os.FileMode {
	return f.Mode
}

func (f *File) GetModTime(_ Scope) time.Time {
	return f.ModTime
}

func (f *File) GetDevice(_ Scope) uint64 {
	return 0
}

func (f *File) WithReader(scope Scope, fn func(io.Reader) error) (err error) {
	defer he(&err)
	var r io.Reader
	if len(f.Contents) > 0 {
		var newContentReader NewContentReader
		scope.Assign(&newContentReader)
		r = newContentReader(f.Contents, f.ChunkLengths)
	} else {
		r = bytes.NewReader(f.ContentBytes)
	}
	err = fn(r)
	ce(err, e5.NewInfo("read %+v", f))
	return nil
}

func (f *File) GetTreeSize(scope Scope) int64 {
	var size int64
	size += f.GetSize(scope)
	for _, sub := range f.Subs {
		if sub.File != nil {
			size += sub.File.GetTreeSize(scope)
		}
		if sub.Pack != nil {
			size += sub.Pack.Size
		}
	}
	return size
}
