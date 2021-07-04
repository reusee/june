// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package fsys

/*
#include <CoreServices/CoreServices.h>

void callback(
  void*,
  size_t,
  void*,
  void*,
  void*
);

void callbackC(
  ConstFSEventStreamRef streamRef,
  void *clientCallBackInfo,
  size_t numEvents,
  void *eventPaths,
  const FSEventStreamEventFlags eventFlags[],
  const FSEventStreamEventId eventIds[]
) {
  callback(
    clientCallBackInfo,
    numEvents,
    eventPaths,
    (void*)eventFlags,
    (void*)eventIds
  );
}

CFArrayRef makePaths(const char* cPath) {
  CFStringRef path = CFStringCreateWithCString(
    kCFAllocatorDefault,
    cPath,
    kCFStringEncodingUTF8
  );
  return CFArrayCreate(
    kCFAllocatorDefault,
    (const void**)&path,
    1,
    NULL
  );
}

*/
import "C"
