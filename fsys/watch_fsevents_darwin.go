// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package fsys

/*
#include <CoreServices/CoreServices.h>

CFArrayRef makePaths(const char* cPath);

void callbackC(
  ConstFSEventStreamRef streamRef,
  void *clientCallBackInfo,
  size_t numEvents,
  void *eventPaths,
  const FSEventStreamEventFlags eventFlags[],
  const FSEventStreamEventId eventIds[]
);

#cgo LDFLAGS: -framework CoreServices
*/
import "C"
import (
	"context"
	"math/rand"
	"reflect"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"
)

type FSEventsWatcher struct {
	watcher   *Watcher
	tapUpdate []TapUpdatePaths
	onUpdated []OnUpdatedSpec
	delaying  []int64
}

var watchers sync.Map

func sysWatcher(
	ctx context.Context,
	path string,
	watcher *Watcher,
	tapUpdate []TapUpdatePaths,
	onUpdated []OnUpdatedSpec,
) (
	add func(string) error,
	err error,
) {
	defer he(&err)

	osWatcher := &FSEventsWatcher{
		watcher:   watcher,
		tapUpdate: tapUpdate,
		onUpdated: onUpdated,
		delaying:  make([]int64, len(onUpdated)),
	}
	id := rand.Int63()
	watchers.Store(id, osWatcher)

	watchFlags := C.FSEventStreamCreateFlags(
		C.kFSEventStreamCreateFlagFileEvents,
	)

	done := make(chan struct{})
	var runLoop C.CFRunLoopRef
	go func() {
		runtime.LockOSThread()
		cPath := C.CString(path)
		stream := C.FSEventStreamCreate(
			C.kCFAllocatorDefault,
			C.FSEventStreamCallback(unsafe.Pointer(C.callbackC)),
			&C.FSEventStreamContext{version: 0, info: unsafe.Pointer(uintptr(id))},
			C.makePaths(cPath),
			C.FSEventStreamEventId(C.kFSEventStreamEventIdSinceNow),
			0,
			watchFlags,
		)
		C.free(unsafe.Pointer(cPath))
		runLoop = C.CFRunLoopGetCurrent()
		C.FSEventStreamScheduleWithRunLoop(
			stream,
			runLoop,
			C.kCFRunLoopDefaultMode,
		)
		C.FSEventStreamStart(stream)
		close(done)
		C.CFRunLoopRun()
		C.FSEventStreamStop(stream)
		C.FSEventStreamUnscheduleFromRunLoop(
			stream,
			runLoop,
			C.kCFRunLoopDefaultMode,
		)
		C.FSEventStreamInvalidate(stream)
		C.FSEventStreamRelease(stream)
	}()
	<-done

	go func() {
		<-ctx.Done()
		C.CFRunLoopStop(runLoop)
	}()

	add = func(string) error {
		return nil
	}

	return
}

//export callback
func callback(
	id unsafe.Pointer,
	numEvents C.size_t,
	pathsC unsafe.Pointer,
	flagsC unsafe.Pointer,
	eventIdsC unsafe.Pointer,
) {
	v, ok := watchers.Load(int64(uintptr(id)))
	if !ok {
		panic("bad id")
	}
	w := v.(*FSEventsWatcher)

	var paths []*C.char
	h := (*reflect.SliceHeader)(unsafe.Pointer(&paths))
	h.Data = uintptr(pathsC)
	h.Len = int(numEvents)
	h.Cap = int(numEvents)
	var flags []C.FSEventStreamEventFlags
	h = (*reflect.SliceHeader)(unsafe.Pointer(&flags))
	h.Data = uintptr(flagsC)
	h.Len = int(numEvents)
	h.Cap = int(numEvents)
	var ids []C.FSEventStreamEventId
	h = (*reflect.SliceHeader)(unsafe.Pointer(&ids))
	h.Data = uintptr(eventIdsC)
	h.Len = int(numEvents)
	h.Cap = int(numEvents)

	now := time.Now()

	var updatedPaths []string
	for i, p := range paths {

		flag := flags[i]
		path := C.GoString(p)

		if flag&C.kFSEventStreamEventFlagMustScanSubDirs > 0 {
			// when kFSEventStreamEventFlagKernelDropped or kFSEventStreamEventFlagUserDropped is set,
			// kFSEventStreamEventFlagMustScanSubDirs will also be set, so no need to check for the two
			_, err := w.watcher.initPath(path, &now)
			if isIgnoreErr(err) {
				continue
			}
			ce(err)

		} else {
			w.watcher.updatePath(now, path)
			updatedPaths = append(updatedPaths, path)
		}

	}

	for _, fn := range w.tapUpdate {
		fn(updatedPaths)
	}

	for i, spec := range w.onUpdated {
		i := i
		spec := spec
		if atomic.CompareAndSwapInt64(&w.delaying[i], 0, 1) {
			time.AfterFunc(spec.MaxDelay, func() {
				spec.Func()
				atomic.StoreInt64(&w.delaying[i], 0)
			})
		}
	}

}
