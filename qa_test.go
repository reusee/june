// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package ling

import (
	"runtime"
	"testing"

	"github.com/reusee/dscope"
	"github.com/reusee/e4qa"
	"github.com/reusee/ling/v2/lingqa"
	"github.com/reusee/pa"
	"github.com/reusee/qa"
)

func TestQA(t *testing.T) {

	//TODO
	if runtime.GOOS == "darwin" {
		t.Skip()
	}

	// qa
	defs := dscope.Methods(new(qa.Def))
	// lingqa
	defs = append(defs, dscope.Methods(new(lingqa.Def))...)
	// e4qa
	defs = append(defs, dscope.Methods(new(e4qa.Def))...)
	// pa
	defs = append(defs, qa.AnalyzersToDefs(pa.Analyzers)...)

	dscope.New(defs...).Sub(func() qa.Args {
		return []string{"./..."}
	}).Call(func(
		check qa.CheckFunc,
	) {
		errs := check()
		if len(errs) > 0 {
			for _, err := range errs {
				pt("-> %s\n", err.Error())
			}
			t.Fatal()
		}
	})

}
