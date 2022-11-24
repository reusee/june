// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package store

import (
	"bytes"
	"context"
	"fmt"

	"github.com/reusee/june/key"
	"github.com/reusee/june/opts"
	"github.com/reusee/june/sys"
	"github.com/reusee/pr2"
	"github.com/reusee/sb"
)

type ScrubOption interface {
	IsScrubOption()
}

type Scrub func(
	ctx context.Context,
	store Store,
	options ...ScrubOption,
) error

func (Def) Scrub(
	newHashState key.NewHashState,
	parallel sys.Parallel,
) Scrub {

	return func(
		ctx context.Context,
		store Store,
		options ...ScrubOption,
	) (err error) {
		defer he(&err)

		var tapKey opts.TapKey
		var tapBad opts.TapBadKey
		for _, option := range options {
			switch option := option.(type) {
			case opts.TapKey:
				tapKey = option
			case opts.TapBadKey:
				tapBad = option
			default:
				panic(fmt.Errorf("bad option: %T", option))
			}
		}

		wg := pr2.NewWaitGroup(ctx)
		defer wg.Cancel()
		put, wait := pr2.Consume(wg, int(parallel), func(_ int, v any) error {
			key := v.(Key)
			if tapKey != nil {
				tapKey(key)
			}
			if err := store.Read(key, func(s sb.Stream) error {
				var sum []byte
				if err := sb.Copy(
					s,
					sb.Hash(newHashState, &sum, nil),
				); err != nil {
					return err
				}
				if !bytes.Equal(key.Hash[:], sum) {
					if tapBad != nil {
						tapBad(key)
					}
				}
				return nil
			}); err != nil {
				return err
			}
			return nil
		})

		if err := store.IterAllKeys(func(key Key) error {
			put(key)
			return nil
		}); err != nil {
			return err
		}
		ce(wait(true))

		return nil
	}
}
