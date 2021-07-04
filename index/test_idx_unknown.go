// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package index

import (
	"testing"

	"github.com/reusee/e4"
	"github.com/reusee/sb"
)

type idxFoo struct{}

func TestIdxUnknown(
	t *testing.T,
) {
	defer he(nil, e4.TestingFatal(t))

	var tokens sb.Tokens
	ce(sb.Copy(
		sb.Marshal(NewEntry(idxFoo{}, "foo", 42)),
		sb.CollectTokens(&tokens),
	))

	tokens[1].Value = "not exists"
	var entry Entry
	ce(sb.Copy(
		tokens.Iter(),
		sb.Unmarshal(&entry),
	))
	if len(entry.Tuple) != 2 {
		t.Fatalf("got %+v\n", entry.Tuple)
	}
	if u, ok := entry.Type.(idxUnknown); !ok {
		t.Fatalf("got %#v", entry.Type)
	} else {
		if u != "not exists" {
			t.Fatal()
		}
	}

	tokens = tokens[:0]
	ce(sb.Copy(
		sb.Marshal(entry),
		sb.CollectTokens(&tokens),
	))
	var entry2 Entry
	ce(sb.Copy(
		tokens.Iter(),
		sb.Unmarshal(&entry2),
	))
	if len(entry.Tuple) != 2 {
		t.Fatal()
	}
	if u, ok := entry.Type.(idxUnknown); !ok {
		t.Fatal()
	} else {
		if u != "not exists" {
			t.Fatal()
		}
	}

}
