// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package virtualfs

/*
#include <stdlib.h>
*/
import "C"

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"sync"
	"unsafe"

	"github.com/reusee/e4"
	"github.com/reusee/june/fsys"
	"github.com/reusee/june/vars"
	"golang.org/x/sys/windows"
)

type NewProjectedFS func(
	rootFS fs.FS,
	destDir string,
) (
	close func(),
	err error,
)

func (_ Def) NewProjectedFS(
	ensureDir fsys.EnsureDir,
	get vars.Get,
	set vars.Set,
) NewProjectedFS {

	return func(
		rootFS fs.FS,
		destDir string,
	) (
		close func(),
		err error,
	) {
		defer he(&err)

		// ensure dest dir
		destDir, err = fsys.RealPath(destDir)
		ce(err)
		ce(ensureDir(destDir))

		// get uuid
		key := "windows-virtual-fs-id/" + destDir
		var idStr string
		err = get(key, &idStr)
		var notFound *vars.NotFound
		var guid windows.GUID
		if as(err, &notFound) {
			guid, err = windows.GenerateGUID()
			ce(err)
			idStr = guid.String()
			ce(set(key, idStr))
		} else if err != nil {
			ce(err)
		} else {
			guid, err = windows.GUIDFromString(idStr)
			ce(err)
		}

		// mark as placeholder
		pathPtr := unsafe.Pointer(
			windows.StringToUTF16Ptr(destDir),
		)
		ce(winErr(PrjMarkDirectoryAsPlaceholder.Call(
			uintptr(pathPtr),
			0,
			0,
			uintptr(unsafe.Pointer(&guid)),
		)))

		pathSeparator := string([]rune{os.PathSeparator})

		type Session struct {
			entries            []fs.DirEntry
			SearchExpr         *uint16
			SearchExprCaptured bool
			NumProvided        int
			NumCalled          int
		}

		var sessions sync.Map

		fsPath := func(path string) string {
			if path == "" {
				return "."
			}
			parts := strings.Split(path, pathSeparator)
			path = strings.Join(parts, "/")
			return path
		}

		getDirEntries := func(
			file fs.File,
		) (
			entries []fs.DirEntry,
			err error,
		) {
			defer he(&err)
			entries, err = file.(fs.ReadDirFile).ReadDir(-1)
			ce(err)
			sort.Slice(entries, func(i, j int) bool {
				a, err := windows.UTF16PtrFromString(entries[i].Name())
				ce(err)
				b, err := windows.UTF16PtrFromString(entries[j].Name())
				ce(err)
				hr, _, _ := PrjFileNameCompare.Call(
					uintptr(unsafe.Pointer(a)),
					uintptr(unsafe.Pointer(b)),
				)
				return int32(hr) < 0
			})
			return
		}

		callbacks := unsafe.Pointer(&PrjCallbacks{

			PRJ_START_DIRECTORY_ENUMERATION_CB: windows.NewCallback(func(
				data *PrjCallbackData,
				guid *windows.GUID,
			) uintptr {
				path := windows.UTF16PtrToString(data.FilePathName)

				file, err := rootFS.Open(fsPath(path))
				if is(err, fs.ErrNotExist) {
					return hresult(0x2)
				}
				ce(err)
				defer file.Close()
				entries, err := getDirEntries(file)
				ce(err)

				session := &Session{
					entries: entries,
				}
				sessions.Store(guid.String(), session)

				return 0
			}),

			PRJ_GET_DIRECTORY_ENUMERATION_CB: windows.NewCallback(func(
				data *PrjCallbackData,
				guid *windows.GUID,
				searchExpr *uint16,
				dirEntryBufferHandle windows.Handle,
			) uintptr {
				path := windows.UTF16PtrToString(data.FilePathName)

				v, ok := sessions.Load(guid.String())
				if !ok {
					return hresult(0x57) // ERROR_INVALID_PARAMETER
				}
				session := v.(*Session)
				defer func() {
					session.NumCalled++
				}()

				if !session.SearchExprCaptured ||
					// PRJ_CB_DATA_FLAG_ENUM_RESTART_SCAN
					data.Flags&1 > 0 {
					if searchExpr != nil {
						expr := windows.UTF16PtrToString(searchExpr)
						session.SearchExpr, err = windows.UTF16PtrFromString(expr)
						ce(err)
					} else {
						session.SearchExpr, err = windows.UTF16PtrFromString("*")
						ce(err)
					}
					session.SearchExprCaptured = true
				}

				if data.Flags&1 > 0 && session.NumCalled > 0 {
					// PRJ_CB_DATA_FLAG_ENUM_RESTART_SCAN

					file, err := rootFS.Open(fsPath(path))
					if is(err, fs.ErrNotExist) {
						return hresult(0x2)
					}
					ce(err)
					defer file.Close()
					entries, err := getDirEntries(file)
					ce(err)

					session.entries = entries
					session.NumProvided = 0
				}

				for len(session.entries) > 0 {
					entry := session.entries[0]

					namePtr, err := windows.UTF16PtrFromString(entry.Name())
					ce(err)
					matched := prjNamePtrMatch(
						namePtr,
						session.SearchExpr,
					)
					if !matched {
						session.entries = session.entries[1:]
						continue
					}

					info, err := entry.Info()
					ce(err)

					fileInfo := (*PrjFileBasicInfo)(C.calloc(1,
						C.ulonglong(unsafe.Sizeof(PrjFileBasicInfo{})),
					))
					if entry.IsDir() {
						fileInfo.IsDirectory = 1
						fileInfo.FileAttributes |= WinFileAttrDirectory
					} else {
						fileInfo.FileSize = info.Size()
						fileInfo.FileAttributes = WinFileAttrNormal
					}

					t := windows.NsecToFiletime(
						info.ModTime().UnixNano(),
					)
					fileInfo.CreationTime = t
					fileInfo.LastAccessTime = t
					fileInfo.LastWriteTime = t
					fileInfo.ChangeTime = t

					var extInfo *PrjExtendedInfo
					if info.Mode()&fs.ModeSymlink > 0 {
						extInfo = (*PrjExtendedInfo)(C.calloc(1,
							C.ulonglong(unsafe.Sizeof(PrjExtendedInfo{})),
						))
						extInfo.InfoType = 1
						f, err := rootFS.Open(fsPath(
							filepath.Join(path, entry.Name()),
						))
						ce(err)
						content, err := io.ReadAll(f)
						ce(err, e4.Close(f))
						ce(f.Close())
						target := string(content)
						extInfo.SymlinkTargetName, err = windows.UTF16PtrFromString(target)
						ce(err)
					}

					hr, _, _ := PrjFillDirEntryBuffer2.Call(
						uintptr(dirEntryBufferHandle),
						uintptr(unsafe.Pointer(namePtr)),
						uintptr(unsafe.Pointer(fileInfo)),
						uintptr(unsafe.Pointer(extInfo)),
					)

					if hr == hresult(0x7a) { // buffer insufficient
						if session.NumProvided == 0 {
							return hresult(0x7a)
						}
						return 0
					}

					session.entries = session.entries[1:]
					session.NumProvided++

				}

				return 0
			}),

			PRJ_END_DIRECTORY_ENUMERATION_CB: windows.NewCallback(func(
				data *PrjCallbackData,
				guid *windows.GUID,
			) uintptr {
				sessions.Delete(guid.String())
				return 0
			}),

			PRJ_GET_PLACEHOLDER_INFO_CB: windows.NewCallback(func(
				data *PrjCallbackData,
			) uintptr {
				path := windows.UTF16PtrToString(data.FilePathName)

				file, err := rootFS.Open(fsPath(path))
				if is(err, fs.ErrNotExist) {
					return hresult(0x2)
				}
				ce(err)
				defer file.Close()
				fileInfo, err := file.Stat()
				ce(err)

				info := (*PrjPlaceholderInfo)(C.calloc(1,
					C.ulonglong(unsafe.Sizeof(PrjPlaceholderInfo{})),
				))
				t := windows.NsecToFiletime(fileInfo.ModTime().UnixNano())
				info.FileBasicInfo = PrjFileBasicInfo{
					CreationTime:   t,
					LastAccessTime: t,
					LastWriteTime:  t,
					ChangeTime:     t,
				}
				if fileInfo.IsDir() {
					info.FileBasicInfo.IsDirectory = 1
				} else {
					info.FileBasicInfo.FileSize = fileInfo.Size()
				}

				path = filepath.Join(
					filepath.Dir(path),
					fileInfo.Name(),
				)
				pathPtr, err := windows.UTF16PtrFromString(path)
				ce(err)

				hr, _, _ := PrjWritePlaceholderInfo.Call(
					uintptr(data.NamespaceVirtualizationContext),
					uintptr(unsafe.Pointer(pathPtr)),
					uintptr(unsafe.Pointer(info)),
					unsafe.Sizeof(*info),
				)

				return hr
			}),

			PRJ_GET_FILE_DATA_CB: windows.NewCallback(func(
				data *PrjCallbackData,
				_offset uint64,
				length uint32,
			) uintptr {
				path := windows.UTF16PtrToString(data.FilePathName)

				file, err := rootFS.Open(fsPath(path))
				if is(err, fs.ErrNotExist) {
					return hresult(0x3)
				}
				ce(err)
				defer file.Close()

				const bufSize = 1 * 1024 * 1024
				ptr, _, _ := PrjAllocateAlignedBuffer.Call(
					uintptr(data.NamespaceVirtualizationContext),
					bufSize,
				)
				var buf []byte
				header := (*reflect.SliceHeader)(unsafe.Pointer(&buf))
				header.Data = ptr
				header.Len = bufSize
				header.Cap = bufSize
				defer func() {
					PrjFreeAlignedBuffer.Call(
						ptr,
					)
				}()

				var offset uintptr
				for {
					n, err := file.Read(buf)
					if n > 0 {
						hr, _, _ := PrjWriteFileData.Call(
							uintptr(data.NamespaceVirtualizationContext),
							uintptr(unsafe.Pointer(&data.DataStreamId)),
							ptr,
							offset,
							uintptr(n),
						)
						if hr != 0 {
							ce(fmt.Errorf("bad write: %d", hr))
						}
						offset += uintptr(n)
					}
					if is(err, io.EOF) {
						break
					}
					ce(err)
				}

				return 0
			}),

			PRJ_NOTIFICATION_CB: windows.NewCallback(func(
				data *PrjCallbackData,
				isDir int32,
				notification uintptr,
				_fileName *uint16,
				params uintptr,
			) uintptr {
				//path := windows.UTF16PtrToString(data.FilePathName)

				switch notification {

				case PRJ_NOTIFY_PRE_RENAME:
					return hresult(0x5) // deny

				case PRJ_NOTIFY_PRE_DELETE:
					return hresultNT(0xC0000121) // STATUS_CANNOT_DELETE from ntstatus.h

				case PRJ_NOTIFY_PRE_SET_HARDLINK,
					PRJ_NOTIFY_FILE_PRE_CONVERT_TO_FULL:
					return hresult(0x5) // deny

				}

				return 0
			}),

			//
		})

		// start
		options := (*PrjStartVirtualizationOptions)(C.calloc(1,
			C.ulonglong(unsafe.Sizeof(PrjStartVirtualizationOptions{})),
		))
		mapping := (*PrjNotificationMapping)(C.calloc(1,
			C.ulonglong(unsafe.Sizeof(PrjNotificationMapping{})),
		))
		mapping.NotificationBitMask = 0 |
			PRJ_NOTIFICATION_PRE_DELETE |
			PRJ_NOTIFICATION_PRE_RENAME |
			PRJ_NOTIFICATION_PRE_SET_HARDLINK |
			PRJ_NOTIFICATION_FILE_PRE_CONVERT_TO_FULL |
			PRJ_NOTIFICATION_NEW_FILE_CREATED |
			PRJ_NOTIFICATION_FILE_OPENED
		rootFilePtr, err := windows.UTF16PtrFromString("")
		ce(err)
		mapping.NotificationRoot = rootFilePtr
		options.NotificationMappings = mapping
		options.NotificationMappingsCount = 1
		var handle windows.Handle
		ce(winErr(PrjStartVirtualizing.Call(
			uintptr(pathPtr),
			uintptr(callbacks),
			0,
			uintptr(unsafe.Pointer(options)),
			uintptr(unsafe.Pointer(&handle)),
		)))

		// close
		close = func() {
			PrjStopVirtualizing.Call(
				uintptr(handle),
			)
		}

		return
	}
}

var (
	prjDLL                        = windows.NewLazyDLL("ProjectedFSLib.dll")
	PrjMarkDirectoryAsPlaceholder = prjDLL.NewProc("PrjMarkDirectoryAsPlaceholder")
	PrjStartVirtualizing          = prjDLL.NewProc("PrjStartVirtualizing")
	PrjStopVirtualizing           = prjDLL.NewProc("PrjStopVirtualizing")
	PrjWritePlaceholderInfo       = prjDLL.NewProc("PrjWritePlaceholderInfo")
	PrjFileNameMatch              = prjDLL.NewProc("PrjFileNameMatch")
	PrjFileNameCompare            = prjDLL.NewProc("PrjFileNameCompare")
	PrjFillDirEntryBuffer2        = prjDLL.NewProc("PrjFillDirEntryBuffer2")
	PrjAllocateAlignedBuffer      = prjDLL.NewProc("PrjAllocateAlignedBuffer")
	PrjFreeAlignedBuffer          = prjDLL.NewProc("PrjFreeAlignedBuffer")
	PrjWriteFileData              = prjDLL.NewProc("PrjWriteFileData")
)

type PrjCallbacks struct {
	PRJ_START_DIRECTORY_ENUMERATION_CB uintptr
	PRJ_END_DIRECTORY_ENUMERATION_CB   uintptr
	PRJ_GET_DIRECTORY_ENUMERATION_CB   uintptr
	PRJ_GET_PLACEHOLDER_INFO_CB        uintptr
	PRJ_GET_FILE_DATA_CB               uintptr
	PRJ_QUERY_FILE_NAME_CB             uintptr
	PRJ_NOTIFICATION_CB                uintptr
	PRJ_CANCEL_COMMAND_CB              uintptr
}

type PrjCallbackData struct {
	Size                           uint32
	Flags                          uint32
	NamespaceVirtualizationContext windows.Handle
	CommandId                      int32
	FileId                         windows.GUID
	DataStreamId                   windows.GUID
	FilePathName                   *uint16
	VersionInfo                    uintptr
	TriggeringProcessId            uint32
	TriggeringProcessImageFileName *uint16
	InstanceContext                uintptr
}

type PrjFileBasicInfo struct {
	IsDirectory    int8
	FileSize       int64
	CreationTime   windows.Filetime
	LastAccessTime windows.Filetime
	LastWriteTime  windows.Filetime
	ChangeTime     windows.Filetime
	FileAttributes uint32
}

type PrjPlaceholderVersionInfo struct {
	ProviderID [128]uint8
	ContentID  [128]uint8
}

type PrjExtendedInfo struct {
	InfoType          uintptr
	NextInfoOffset    uintptr
	SymlinkTargetName *uint16
}

type PrjPlaceholderInfo struct {
	FileBasicInfo PrjFileBasicInfo

	EaInformation struct {
		EaBufferSize    uint32
		OffsetToFirstEa uint32
	}

	SecurityInformation struct {
		SecurityBufferSize         uint32
		OffsetToSecurityDescriptor uint32
	}

	StreamsInformation struct {
		StreamsInfoBufferSize   uint32
		OffsetToFirstStreamInfo uint32
	}

	VersionInfo  PrjPlaceholderVersionInfo
	VariableData [1]uint8
}

type PrjNotificationMapping struct {
	NotificationBitMask uint16
	NotificationRoot    *uint16
}

type PrjStartVirtualizationOptions struct {
	Flags                     uint32
	PoolThreadCount           uint32
	ConcurrentThreadCount     uint32
	NotificationMappings      *PrjNotificationMapping
	NotificationMappingsCount uint32
}

func prjNameMatch(a, b string) bool {
	aPtr, err := windows.UTF16PtrFromString(a)
	ce(err)
	bPtr, err := windows.UTF16PtrFromString(b)
	ce(err)
	return prjNamePtrMatch(aPtr, bPtr)
}

func prjNamePtrMatch(a, b *uint16) bool {
	res, _, _ := PrjFileNameMatch.Call(
		uintptr(unsafe.Pointer(a)),
		uintptr(unsafe.Pointer(b)),
	)
	return int32(res) > 0
}

const (
	WinFileAttrDirectory = 0x10
	WinFileAttrNormal    = 0x80
)

const (
	PRJ_NOTIFY_NONE                               = 0x00000000
	PRJ_NOTIFY_SUPPRESS_NOTIFICATIONS             = 0x00000001
	PRJ_NOTIFY_FILE_OPENED                        = 0x00000002
	PRJ_NOTIFY_NEW_FILE_CREATED                   = 0x00000004
	PRJ_NOTIFY_FILE_OVERWRITTEN                   = 0x00000008
	PRJ_NOTIFY_PRE_DELETE                         = 0x00000010
	PRJ_NOTIFY_PRE_RENAME                         = 0x00000020
	PRJ_NOTIFY_PRE_SET_HARDLINK                   = 0x00000040
	PRJ_NOTIFY_FILE_RENAMED                       = 0x00000080
	PRJ_NOTIFY_HARDLINK_CREATED                   = 0x00000100
	PRJ_NOTIFY_FILE_HANDLE_CLOSED_NO_MODIFICATION = 0x00000200
	PRJ_NOTIFY_FILE_HANDLE_CLOSED_FILE_MODIFIED   = 0x00000400
	PRJ_NOTIFY_FILE_HANDLE_CLOSED_FILE_DELETED    = 0x00000800
	PRJ_NOTIFY_FILE_PRE_CONVERT_TO_FULL           = 0x00001000
	PRJ_NOTIFY_USE_EXISTING_MASK                  = 0xFFFFFFFF
)

const (
	PRJ_NOTIFICATION_FILE_OPENED                        = 0x00000002
	PRJ_NOTIFICATION_NEW_FILE_CREATED                   = 0x00000004
	PRJ_NOTIFICATION_FILE_OVERWRITTEN                   = 0x00000008
	PRJ_NOTIFICATION_PRE_DELETE                         = 0x00000010
	PRJ_NOTIFICATION_PRE_RENAME                         = 0x00000020
	PRJ_NOTIFICATION_PRE_SET_HARDLINK                   = 0x00000040
	PRJ_NOTIFICATION_FILE_RENAMED                       = 0x00000080
	PRJ_NOTIFICATION_HARDLINK_CREATED                   = 0x00000100
	PRJ_NOTIFICATION_FILE_HANDLE_CLOSED_NO_MODIFICATION = 0x00000200
	PRJ_NOTIFICATION_FILE_HANDLE_CLOSED_FILE_MODIFIED   = 0x00000400
	PRJ_NOTIFICATION_FILE_HANDLE_CLOSED_FILE_DELETED    = 0x00000800
	PRJ_NOTIFICATION_FILE_PRE_CONVERT_TO_FULL           = 0x00001000
)
