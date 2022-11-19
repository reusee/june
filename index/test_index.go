// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package index

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"sync"
	"testing"

	"github.com/reusee/e5"
	"github.com/reusee/june/key"
	"github.com/reusee/pp"
	"github.com/reusee/sb"
)

// test Index implementation
type TestIndex func(
	ctx context.Context,
	withIndexManager func(func(IndexManager)),
	t *testing.T,
)

type testingIndex struct {
	Int int
}

var TestingIndex = testingIndex{}

type testingIndex2 struct {
	S1  string
	S2  string
	Int int
}

var TestingIndex2 = testingIndex2{}

func init() {
	Register(TestingIndex)
	Register(TestingIndex2)
}

func (Def) TestIndex(
	scope Scope,
) TestIndex {
	return func(
		ctx context.Context,
		withIndexManager func(func(IndexManager)),
		t *testing.T,
	) {
		defer he(nil, e5.TestingFatal(t))

		n := 30
		wg := new(sync.WaitGroup)
		wg.Add(n)
		for i := 0; i < n; i++ {
			go func() {
				defer wg.Done()

				// basic
				withIndexManager(func(
					indexManager IndexManager,
				) {

					id := StoreID(fmt.Sprintf("%d", rand.Int63()))

					scope.Fork(&id, &indexManager).Call(func(
						index Index,
						selIndex SelectIndex,
					) {

						k, err := key.KeyFromString("foo:beef")
						ce(err)

						// invalid
						err = index.Save(ctx, Entry{
							Type: nil,
						})
						if !is(err, ErrInvalidEntry) {
							t.Fatalf("got %v\n", err)
						}
						err = index.Save(ctx, Entry{
							Type: idxFoo{},
						})
						if !is(err, ErrInvalidEntry) {
							t.Fatalf("got %v\n", err)
						}

						// add
						num := int(rand.Int63())
						entry := NewEntry(TestingIndex, num, k)
						err = index.Save(ctx, entry)
						ce(err)

						// select
						n := 0
						err = selIndex(
							ctx,
							Asc,
							Sink(func() sb.Sink {
								return sb.AltSink(
									// Entry
									sb.Unmarshal(func(name string, i int, _ Key) {
										if name != "index.testingIndex" {
											t.Fatalf("got %s", name)
										}
										n++
										if i != num {
											t.Fatal()
										}
									}),
									// PreEntry
									sb.Unmarshal(func(_ Key, name string, i int) {
										if name != "index.testingIndex" {
											t.Fatalf("got %s", name)
										}
										n++
										if i != num {
											t.Fatal()
										}
									}),
								)
							}),
						)
						ce(err)
						if n != 2 {
							t.Fatalf("got %d\n", n)
						}

						// exact
						n = 0
						ce(selIndex(
							ctx,
							Exact(entry),
							Count(&n),
						))
						if n != 1 {
							t.Fatal()
						}

						// same tuple
						entry = NewEntry(TestingIndex, num, k)
						err = index.Save(ctx, entry)
						ce(err)

						n = 0
						err = selIndex(
							ctx,
							Asc,
							Sink(func() sb.Sink {
								return sb.AltSink(
									// Entry
									sb.Unmarshal(func(name string, i int, _ Key) {
										if name != "index.testingIndex" {
											t.Fatal()
										}
										n++
										if i != num {
											t.Fatal()
										}
									}),
									// PreEntry
									sb.Unmarshal(func(_ Key, name string, i int) {
										if name != "index.testingIndex" {
											t.Fatal()
										}
										n++
										if i != num {
											t.Fatal()
										}
									}),
								)
							}),
						)
						ce(err)
						if n != 2 {
							t.Fatal()
						}

						n = 0
						err = selIndex(
							ctx,
							Asc,
							Sink(func() sb.Sink {
								return sb.AltSink(
									sb.Unmarshal(func(_ string, i int, _ Key) {
										n++
										if i != num {
											t.Fatal()
										}
									}),
									sb.Unmarshal(func(_ Key, _ string, i int) {
										n++
										if i != num {
											t.Fatal()
										}
									}),
								)
							}),
						)
						ce(err)
						if n != 2 {
							t.Fatal()
						}

						n = 0
						err = Select(
							ctx,
							index,
							Asc,
							Sink(func() sb.Sink {
								return sb.AltSink(
									sb.Unmarshal(func(_ string, i int, _ Key) {
										n++
										if i != num {
											t.Fatal()
										}
									}),
									sb.Unmarshal(func(_ Key, _ string, i int) {
										n++
										if i != num {
											t.Fatal()
										}
									}),
								)
							}),
						)
						ce(err)
						if n != 2 {
							t.Fatalf("got %d", n)
						}

						n = 0
						err = Select(
							ctx,
							index,
							Asc,
							Sink(func() sb.Sink {
								return sb.AltSink(
									sb.Unmarshal(func(_ string, i int, _ Key) {
										n++
										if i != num {
											t.Fatal()
										}
									}),
									sb.Unmarshal(func(_ Key, _ string, i int) {
										n++
										if i != num {
											t.Fatal()
										}
									}),
								)
							}),
						)
						ce(err)
						if n != 2 {
							t.Fatal()
						}

						n = 0
						err = Select(
							ctx,
							index,
							Lower(NewEntry(TestingIndex, 1)),
							Upper(NewEntry(TestingIndex, num+1)),
							Asc,
							Unmarshal(func(_ string, i int, _ Key) {
								n++
								if i != num {
									t.Fatal()
								}
							}),
						)
						ce(err)
						if n != 1 {
							t.Fatalf("got %d\n", n)
						}

						n = 0
						err = Select(
							ctx,
							index,
							Lower(NewEntry(TestingIndex, num+1)),
							Upper(NewEntry(TestingIndex, 1)),
							Asc,
							Unmarshal(func(_ int, _ Key) {
								n++
							}),
						)
						ce(err)
						if n != 0 {
							t.Fatal()
						}

						err = Select(
							ctx,
							index,
							Lower(NewEntry(TestingIndex, 1)),
							Upper(NewEntry(TestingIndex, num+1)),
							Count(&n),
						)
						ce(err)
						if n == 0 {
							t.Fatal(err)
						}

						err = Select(
							ctx,
							index,
							Lower(NewEntry(TestingIndex, num+1)),
							Upper(NewEntry(TestingIndex, 1)),
							Count(&n),
						)
						ce(err)
						if n > 0 {
							t.Fatalf("got %d", n)
						}

						err = Select(
							ctx,
							index,
							Lower(NewEntry(TestingIndex, 1)),
							Upper(NewEntry(TestingIndex, num)),
							Count(&n),
						)
						ce(err)
						if n > 0 {
							t.Fatalf("got %d", n)
						}

						entry = NewEntry(TestingIndex, num+1, k)
						err = index.Save(ctx, entry)
						ce(err)

						// iter desc
						n = 0
						err = Select(
							ctx,
							index,
							Desc,
							Sink(func() sb.Sink {
								return sb.AltSink(
									sb.Unmarshal(func(_ string, i int, _ Key) {
										switch n {
										case 0:
											if i != num+1 {
												t.Fatal()
											}
										case 1:
											if i != num {
												t.Fatal()
											}
										}
										n++
									}),
									sb.Unmarshal(func(_ Key, _ string, i int) {
										switch n {
										case 0:
											if i != num+1 {
												t.Fatal()
											}
										case 1:
											if i != num {
												t.Fatal()
											}
										}
										n++
									}),
								)
							}),
						)
						ce(err)
						if n != 4 {
							t.Fatalf("got %d", n)
						}

						n = 0
						err = Select(
							ctx,
							index,
							Asc,
							Sink(func() sb.Sink {
								return sb.AltSink(
									sb.Unmarshal(func(_ string, i int, _ Key) {
										switch n {
										case 0:
											if i != num {
												t.Fatal()
											}
										case 1:
											if i != num+1 {
												t.Fatal()
											}
										}
										n++
									}),
									sb.Unmarshal(func(_ Key, _ string, i int) {
										switch n {
										case 0:
											if i != num {
												t.Fatal()
											}
										case 1:
											if i != num+1 {
												t.Fatal()
											}
										}
										n++
									}),
								)
							}),
						)
						ce(err)
						if n != 4 {
							t.Fatal()
						}

						// prefix
						err = index.Save(ctx, NewEntry(
							TestingIndex2, "foo", "foo", 1, k,
						))
						ce(err)
						err = index.Save(ctx, NewEntry(
							TestingIndex2, "foo", "foo", 2, k,
						))
						ce(err)
						err = index.Save(ctx, NewEntry(
							TestingIndex2, "foo", "bar", 3, k,
						))
						ce(err)

						n = 0
						err = Select(
							ctx,
							index,
							MatchEntry(TestingIndex2, "foo"),
							Unmarshal(func(_ string, p string, _ string, _ int, _ Key) {
								if p != "foo" {
									t.Fatal()
								}
								n++
							}),
						)
						ce(err)
						if n != 3 {
							t.Fatal()
						}

						n = 0
						err = Select(
							ctx,
							index,
							MatchEntry(TestingIndex2, "foo", "foo"),
							Unmarshal(func(_ string, p string, p2 string, _ int, _ Key) {
								if p != "foo" {
									t.Fatal()
								}
								if p2 != "foo" {
									t.Fatal()
								}
								n++
							}),
						)
						ce(err)
						if n != 2 {
							t.Fatal()
						}

						n = 0
						err = Select(
							ctx,
							index,
							MatchEntry(TestingIndex2, "foo", "bar"),
							Unmarshal(func(_ string, p string, p2 string, _ int, _ Key) {
								if p != "foo" {
									t.Fatal()
								}
								if p2 != "bar" {
									t.Fatal()
								}
								n++
							}),
						)
						ce(err)
						if n != 1 {
							t.Fatal()
						}

						n = 0
						ce(Select(
							ctx,
							index,
							MatchEntry(TestingIndex2, "baz"),
							Unmarshal(func(_ string, _ string, _ string, _ int, _ Key) {
								n++
							}),
						))
						if n != 0 {
							t.Fatal()
						}

						// where
						n = 0
						ce(Select(
							ctx,
							index,
							MatchEntry(TestingIndex2),
							Where(func(s sb.Stream) bool {
								var entry Entry
								ce(sb.Copy(s, sb.Unmarshal(&entry)))
								return entry.Tuple[2] == 3
							}),
							Count(&n),
						))
						if n != 1 {
							t.Fatal()
						}

						// count
						n = 0
						ce(Select(
							ctx,
							index,
							MatchEntry(TestingIndex2, "foo"),
							Count(&n),
						))
						if n != 3 {
							t.Fatal()
						}

						n = 0
						ce(Select(
							ctx,
							index,
							Count(&n),
							Unmarshal(func(args ...any) {
							}),
						))
						if n != 10 {
							t.Fatal()
						}

						// limit
						n = 0
						ce(Select(
							ctx,
							index,
							Count(&n),
							Limit(1),
							Unmarshal(func(args ...any) {
							}),
						))
						if n != 1 {
							t.Fatal()
						}

						n = 0
						ce(Select(
							ctx,
							index,
							Count(&n),
							Limit(1),
						))
						if n != 1 {
							t.Fatal()
						}

						// offset
						n = 0
						ce(Select(
							ctx,
							index,
							Offset(0),
							Count(&n),
						))
						if n != 10 {
							t.Fatal()
						}
						n = 0
						ce(Select(
							ctx,
							index,
							Offset(0),
							Limit(3),
							Count(&n),
						))
						if n != 3 {
							t.Fatal()
						}
						n = 0
						ce(Select(
							ctx,
							index,
							Offset(1),
							Count(&n),
						))
						if n != 9 {
							t.Fatalf("got %d\n", n)
						}
						n = 0
						ce(Select(
							ctx,
							index,
							Offset(2),
							Count(&n),
						))
						if n != 8 {
							t.Fatal()
						}
						n = 0
						ce(Select(
							ctx,
							index,
							Offset(2),
							Count(&n),
							MatchEntry(TestingIndex2, "foo"),
						))
						if n != 1 {
							t.Fatal()
						}

						// prefix and variadic call
						n = 0
						ce(Select(
							ctx,
							index,
							MatchEntry(TestingIndex2, "foo"),
							Unmarshal(func(tuple ...any) {
							}),
							Count(&n),
						))
						if n != 3 {
							t.Fatal()
						}

						// context
						ctx, cancel := context.WithCancel(ctx)
						cancel()
						err = Select(
							ctx,
							index,
							WithCtx{ctx},
						)
						if !errors.Is(err, Cancel) {
							t.Fatal()
						}

						// delete
						iter, closer, err := index.Iter(
							ctx,
							nil,
							nil,
							Asc,
						)
						ce(err)
						var toDelete []Entry
						ce(pp.Copy(iter, pp.Tap(func(v any) (err error) {
							s := v.(sb.Stream)
							defer he(&err)
							var entry *Entry
							var preEntry *PreEntry
							ce(sb.Copy(
								s,
								sb.AltSink(
									sb.Unmarshal(&entry),
									sb.Unmarshal(&preEntry),
								),
							))
							if entry != nil {
								toDelete = append(toDelete, *entry)
							}
							return nil
						})))
						ce(closer.Close())
						for _, tuple := range toDelete {
							ce(index.Delete(ctx, tuple))
						}

						iter, closer, err = index.Iter(
							ctx,
							nil,
							nil,
							Asc,
						)
						ce(err)
						ce(pp.Copy(iter, pp.Tap(func(_ any) error {
							// should all be deleted
							t.Fatal()
							return nil
						})))
						ce(closer.Close())

						// extra save, to test id isolation
						entry = NewEntry(TestingIndex, num+1, k)
						ce(index.Save(ctx, entry))

					})

				})
			}()

		}

		wg.Wait()

	}
}
