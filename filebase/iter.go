package filebase

import (
	"github.com/reusee/pp"
	"io"
	"os"
	"time"
)

type IterItem struct {
	*File
	*FileInfo
	*FileInfoThunk
	*PackThunk
	*Pack
	*Virtual
	*ZipItem
}

type Src func() (*IterItem, Src, error)

type Sink func(*IterItem) (Sink, error)

var Get = pp.Get[IterItem, Src]

type FileInfoThunk struct {
	FileInfo FileInfo
	Expand   func(bool)
	Path     string // relative path of file
}

type PackThunk struct {
	Expand func(bool)
	Path   string
	Pack   // relative path of dir
}

type FileInfo struct {
	FileLike
	Path string // relative path of file
}

type FileLike interface {
	// basic infos
	GetIsDir(Scope) bool
	GetName(Scope) string
	GetSize(Scope) int64
	GetMode(Scope) os.FileMode
	GetModTime(Scope) time.Time
	GetDevice(Scope) uint64

	// content
	WithReader(
		Scope,
		func(io.Reader) error,
	) error
}

type ZipItem struct {
	A   *IterItem
	B   *IterItem
	Dir string
}
