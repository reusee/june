// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package storepebble

import (
	"fmt"

	"github.com/cockroachdb/pebble"
)

type Logger struct{}

var _ pebble.Logger = new(Logger)

func (l *Logger) Infof(format string, args ...any) {
	print(we(fmt.Errorf(format, args...)).Error())
}

func (l *Logger) Fatalf(format string, args ...any) {
	ce(we(fmt.Errorf(format, args...)))
}
