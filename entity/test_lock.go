// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package entity

import (
	"sync"
	"testing"

	"github.com/reusee/e4"
	"github.com/reusee/june/key"
	"github.com/reusee/sb"
)

func TestLock(
	t *testing.T,
	newHashState key.NewHashState,
) {
	defer he(nil, e4.TestingFatal(t))

	locks := newEntityLocks()
	var wg sync.WaitGroup
	ns := make([]int, 512)

	for i := 0; i < 512; i++ {
		i := i
		var hash []byte
		ce(sb.Copy(
			sb.Marshal(i),
			sb.Hash(newHashState, &hash, nil),
		))
		var key Key
		copy(key.Hash[:], hash)
		for j := 0; j < 8; j++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				unlock := locks.Lock(key)
				defer unlock()
				ns[i]++
			}()
		}
	}

	wg.Wait()
	for _, n := range ns {
		if n != 8 {
			t.Fatalf("got %d\n", n)
		}
	}

}
