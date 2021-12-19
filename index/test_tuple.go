// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package index

import (
	"io"
	"reflect"
	"testing"

	"github.com/reusee/june/naming"
	"github.com/reusee/sb"
)

func TestTuple(
	t *testing.T,
) {

	var tuple sb.Tuple
	err := sb.Copy(
		sb.Tokens{
			{Kind: sb.KindTuple},
			{Kind: sb.KindString, Value: "foo"},
		}.Iter(),
		sb.Unmarshal(&tuple),
	)
	if !is(err, io.ErrUnexpectedEOF) {
		t.Fatal()
	}

	err = sb.Copy(
		sb.Tokens{
			{Kind: sb.KindTuple},
			{Kind: sb.KindString, Value: naming.Type(reflect.TypeOf(TestingIndex))},
		}.Iter(),
		sb.Unmarshal(&tuple),
	)
	if !is(err, io.ErrUnexpectedEOF) {
		t.Fatal()
	}

}
