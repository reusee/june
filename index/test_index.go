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
	"github.com/reusee/pr2"
	"github.com/reusee/sb"
)

// test Index implementation
type TestIndex func(
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
						wg *pr2.WaitGroup,
					) {

						k, err := key.KeyFromString("foo:beef")
						ce(err)

						// invalid
						err = index.Save(Entry{
							Type: nil,
						})
						if !is(err, ErrInvalidEntry) {
							t.Fatalf("got %v\n", err)
						}
						err = index.Save(Entry{
							Type: idxFoo{},
						})
						if !is(err, ErrInvalidEntry) {
							t.Fatalf("got %v\n", err)
						}

						// add
						num := int(rand.Int63())
						entry := NewEntry(TestingIndex, num, k)
						err = index.Save(entry)
						ce(err)

						// select
						n := 0
						err = selIndex(
							wg,
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
							wg,
							Exact(entry),
							Count(&n),
						))
						if n != 1 {
							t.Fatal()
						}

						// same tuple
						entry = NewEntry(TestingIndex, num, k)
						err = index.Save(entry)
						ce(err)

						n = 0
						err = selIndex(
							wg,
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
							wg,
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
						err = index.Save(entry)
						ce(err)

						// iter desc
						n = 0
						err = Select(
							index,
							Desc,
							Sink(func() sb.Sink {
								return sb.AltSink(
									sb.Unmarshal(func(_ string, i int, _ Key) {
										if n == 0 {
											if i != num+1 {
												t.Fatal()
											}
										} else if n == 1 {
											if i != num {
												t.Fatal()
											}
										}
										n++
									}),
									sb.Unmarshal(func(_ Key, _ string, i int) {
										if n == 0 {
											if i != num+1 {
												t.Fatal()
											}
										} else if n == 1 {
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
							index,
							Asc,
							Sink(func() sb.Sink {
								return sb.AltSink(
									sb.Unmarshal(func(name string, i int, key Key) {
										if n == 1 {
											if i != num+1 {
												t.Fatal()
											}
										} else if n == 0 {
											if i != num {
												t.Fatal()
											}
										}
										n++
									}),
									sb.Unmarshal(func(key Key, name string, i int) {
										if n == 1 {
											if i != num+1 {
												t.Fatal()
											}
										} else if n == 0 {
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
							t.Fatal()
						}

						// prefix
						err = index.Save(NewEntry(
							TestingIndex2, "foo", "foo", 1, k,
						))
						ce(err)
						err = index.Save(NewEntry(
							TestingIndex2, "foo", "foo", 2, k,
						))
						ce(err)
						err = index.Save(NewEntry(
							TestingIndex2, "foo", "bar", 3, k,
						))
						ce(err)

						n = 0
						err = Select(
							index,
							MatchEntry(TestingIndex2, "foo"),
							Unmarshal(func(name string, p string, p2 string, i int, key Key) {
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
							index,
							MatchEntry(TestingIndex2, "foo", "foo"),
							Unmarshal(func(name string, p string, p2 string, i int, key Key) {
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
							index,
							MatchEntry(TestingIndex2, "foo", "bar"),
							Unmarshal(func(name string, p string, p2 string, i int, key Key) {
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
							index,
							MatchEntry(TestingIndex2, "baz"),
							Unmarshal(func(name string, p string, p2 string, i int, key Key) {
								n++
							}),
						))
						if n != 0 {
							t.Fatal()
						}

						// where
						n = 0
						ce(Select(
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
							index,
							MatchEntry(TestingIndex2, "foo"),
							Count(&n),
						))
						if n != 3 {
							t.Fatal()
						}

						n = 0
						ce(Select(
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
							index,
							Offset(0),
							Count(&n),
						))
						if n != 10 {
							t.Fatal()
						}
						n = 0
						ce(Select(
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
							index,
							Offset(1),
							Count(&n),
						))
						if n != 9 {
							t.Fatalf("got %d\n", n)
						}
						n = 0
						ce(Select(
							index,
							Offset(2),
							Count(&n),
						))
						if n != 8 {
							t.Fatal()
						}
						n = 0
						ce(Select(
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
						ctx, cancel := context.WithCancel(wg)
						cancel()
						err = Select(
							index,
							WithCtx{ctx},
						)
						if !errors.Is(err, Cancel) {
							t.Fatal()
						}

						// delete
						iter, closer, err := index.Iter(
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
							ce(index.Delete(tuple))
						}

						iter, closer, err = index.Iter(
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
						ce(index.Save(entry))

					})

				})
			}()

		}

		wg.Wait()

	}
}
