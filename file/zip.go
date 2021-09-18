// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package file

import (
	"fmt"
	"path/filepath"
)

type Zip func(
	a Src,
	b Src,
	cont Src,
	options ...ZipOption,
) Src

type ZipOption interface {
	IsZipOption()
}

func (_ Def) Zip(
	scope Scope,
) Zip {

	var zip func(a Src, b Src, cont Src, options ...ZipOption) Src
	zip = func(a Src, b Src, cont Src, options ...ZipOption) Src {
		return func() (*IterItem, Src, error) {

			valueA, err := Get(&a)
			if err != nil { // NOCOVER
				return nil, nil, err
			}
			valueB, err := Get(&b)
			if err != nil { // NOCOVER
				return nil, nil, err
			}

			if valueA == nil && valueB == nil {
				return nil, cont, nil
			}

			var dirA, dirB string

			if valueA != nil {
				if valueA.FileInfo != nil {
					dirA = valueA.FileInfo.Path
					dirA = filepath.Dir(dirA)
				} else if valueA.FileInfoThunk != nil {
					dirA = valueA.FileInfoThunk.Path
					dirA = filepath.Dir(dirA)
				} else if valueA.PackThunk != nil {
					dirA = valueA.PackThunk.Path
				} else {
					panic(we(fmt.Errorf("unknown type %T", valueA)))
				}
			}
			if valueB != nil {
				if valueB.FileInfo != nil {
					dirB = valueB.FileInfo.Path
					dirB = filepath.Dir(dirB)
				} else if valueB.FileInfoThunk != nil {
					dirB = valueB.FileInfoThunk.Path
					dirB = filepath.Dir(dirB)
				} else if valueB.PackThunk != nil {
					dirB = valueB.PackThunk.Path
				} else {
					panic(we(fmt.Errorf("unknown type %T", valueB)))
				}
			}

			if valueB == nil {
				if valueA.FileInfoThunk != nil {
					valueA.FileInfoThunk.Expand(false)
				} else if valueA.PackThunk != nil {
					valueA.PackThunk.Expand(false)
				}
				return &IterItem{
						ZipItem: &ZipItem{A: valueA, B: nil, Dir: dirA},
					}, zip(a, nil, cont, options...),
					nil

			} else if valueA == nil {
				if valueB.FileInfoThunk != nil {
					valueB.FileInfoThunk.Expand(false)
				} else if valueB.PackThunk != nil {
					valueB.PackThunk.Expand(false)
				}
				return &IterItem{
						ZipItem: &ZipItem{A: nil, B: valueB, Dir: dirB},
					}, zip(nil, b, cont, options...),
					nil
			}

			if isDeeper(dirA, dirB) {
				return &IterItem{
						ZipItem: &ZipItem{A: valueA, B: nil, Dir: dirA},
					}, zip(
						a,
						func() (*IterItem, Src, error) {
							return valueB, b, nil
						},
						cont,
						options...,
					),
					nil

			} else if isDeeper(dirB, dirA) {
				return &IterItem{
						ZipItem: &ZipItem{A: nil, B: valueB, Dir: dirB},
					}, zip(
						func() (*IterItem, Src, error) {
							return valueA, a, nil
						},
						b,
						cont,
						options...,
					),
					nil
			}

			if valueA.FileInfo != nil {

				if valueB.FileInfo != nil {
					// (FileInfo, FileInfo)
					nameA := valueA.FileInfo.GetName(scope)
					nameB := valueB.FileInfo.GetName(scope)
					if nameA < nameB {
						return &IterItem{
								ZipItem: &ZipItem{A: valueA, B: nil, Dir: dirA},
							}, zip(
								a,
								func() (*IterItem, Src, error) {
									return valueB, b, nil
								},
								cont,
								options...,
							),
							nil

					} else if nameB < nameA {
						return &IterItem{
								ZipItem: &ZipItem{A: nil, B: valueB, Dir: dirB},
							}, zip(
								func() (*IterItem, Src, error) {
									return valueA, a, nil
								},
								b,
								cont,
								options...,
							),
							nil
					}

					if dirA != dirB {
						panic(fmt.Errorf("bad iter, expecting same path: %v %v", dirA, dirB))
					}
					return &IterItem{
							ZipItem: &ZipItem{A: valueA, B: valueB, Dir: dirA},
						}, zip(
							a,
							b,
							cont,
							options...,
						),
						nil

				} else if valueB.FileInfoThunk != nil {
					// (FileInfo, FileInfoThunk)
					nameA := valueA.FileInfo.GetName(scope)
					nameB := valueB.FileInfoThunk.FileInfo.GetName(scope)
					if nameA < nameB {
						return &IterItem{
								ZipItem: &ZipItem{A: valueA, B: nil, Dir: dirA},
							}, zip(
								a,
								func() (*IterItem, Src, error) {
									return valueB, b, nil
								},
								cont,
								options...,
							),
							nil

					} else if nameB < nameA {
						valueB.FileInfoThunk.Expand(false)
						return &IterItem{
								ZipItem: &ZipItem{A: nil, B: valueB, Dir: dirB},
							}, zip(
								func() (*IterItem, Src, error) {
									return valueA, a, nil
								},
								b,
								cont,
								options...,
							),
							nil
					}

					valueB.FileInfoThunk.Expand(true)
					return nil,
						zip(
							func() (*IterItem, Src, error) {
								return valueA, a, nil
							},
							func() (*IterItem, Src, error) {
								return &IterItem{
									FileInfo: &valueB.FileInfoThunk.FileInfo,
								}, b, nil
							},
							cont,
							options...,
						),
						nil

				} else if valueB.PackThunk != nil {
					// (FileInfo, PackThunk)
					nameA := valueA.FileInfo.GetName(scope)

					if nameA < valueB.Min {
						return &IterItem{
								ZipItem: &ZipItem{A: valueA, B: nil, Dir: dirA},
							}, zip(
								a,
								func() (*IterItem, Src, error) {
									return valueB, b, nil
								},
								cont,
								options...,
							),
							nil

					} else if nameA > valueB.Max {
						valueB.PackThunk.Expand(false)
						return &IterItem{
								ZipItem: &ZipItem{A: nil, B: valueB, Dir: dirB},
							}, zip(
								func() (*IterItem, Src, error) {
									return valueA, a, nil
								},
								b,
								cont,
								options...,
							),
							nil

					} else {
						// unpack
						valueB.PackThunk.Expand(true)
						return nil,
							zip(
								func() (*IterItem, Src, error) {
									return valueA, a, nil
								},
								b,
								cont,
								options...,
							),
							nil
					}

				}

			} else if valueA.FileInfoThunk != nil {

				if valueB.FileInfoThunk != nil {
					// (FileInfoThunk, FileInfoThunk)
					nameA := valueA.FileInfoThunk.FileInfo.GetName(scope)
					nameB := valueB.FileInfoThunk.FileInfo.GetName(scope)
					if nameA < nameB {
						valueA.FileInfoThunk.Expand(false)
						return &IterItem{
								ZipItem: &ZipItem{A: valueA, B: nil, Dir: dirA},
							}, zip(
								a,
								func() (*IterItem, Src, error) {
									return valueB, b, nil
								},
								cont,
								options...,
							),
							nil

					} else if nameB < nameA {
						valueB.FileInfoThunk.Expand(false)
						return &IterItem{
								ZipItem: &ZipItem{A: nil, B: valueB, Dir: dirB},
							}, zip(
								func() (*IterItem, Src, error) {
									return valueA, a, nil
								},
								b,
								cont,
								options...,
							),
							nil
					}

					if dirA != dirB {
						panic(fmt.Errorf("bad iter, expecting same path: %v %v", dirA, dirB))
					}

					var predictExpand PredictExpandFileInfoThunk
					for _, option := range options {
						switch option := option.(type) {
						case PredictExpandFileInfoThunk:
							predictExpand = option
						default:
							panic(fmt.Errorf("unknown option: %T", option))
						}
					}
					expand := true
					if predictExpand != nil {
						res, err := predictExpand(*valueA.FileInfoThunk, *valueB.FileInfoThunk)
						if err != nil {
							return nil, nil, err
						}
						expand = res
					}

					if expand {
						// iter subs
						valueA.FileInfoThunk.Expand(true)
						valueB.FileInfoThunk.Expand(true)
						return nil,
							zip(
								func() (*IterItem, Src, error) {
									return &IterItem{
										FileInfo: valueA.FileInfo,
									}, a, nil
								},
								func() (*IterItem, Src, error) {
									return &IterItem{
										FileInfo: valueB.FileInfo,
									}, b, nil
								},
								cont,
								options...,
							),
							nil

					} else {
						valueA.FileInfoThunk.Expand(false)
						valueB.FileInfoThunk.Expand(false)
						return &IterItem{
								ZipItem: &ZipItem{A: valueA, B: valueB, Dir: dirA},
							}, zip(
								a,
								b,
								cont,
								options...,
							),
							nil
					}

				} else if valueB.FileInfo != nil {
					// (FileInfoThunk, FileInfo)
					nameA := valueA.FileInfoThunk.FileInfo.GetName(scope)
					nameB := valueB.FileInfo.GetName(scope)
					if nameA < nameB {
						valueA.FileInfoThunk.Expand(false)
						return &IterItem{
								ZipItem: &ZipItem{A: valueA, B: nil, Dir: dirA},
							}, zip(
								a,
								func() (*IterItem, Src, error) {
									return valueB, b, nil
								},
								cont,
								options...,
							),
							nil

					} else if nameB < nameA {
						return &IterItem{
								ZipItem: &ZipItem{A: nil, B: valueB, Dir: dirB},
							}, zip(
								func() (*IterItem, Src, error) {
									return valueA, a, nil
								},
								b,
								cont,
								options...,
							),
							nil
					}

					valueA.FileInfoThunk.Expand(true)
					return nil,
						zip(
							func() (*IterItem, Src, error) {
								return &IterItem{
									FileInfo: valueA.FileInfo,
								}, a, nil
							},
							func() (*IterItem, Src, error) {
								return &IterItem{
									FileInfo: valueB.FileInfo,
								}, b, nil
							},
							cont,
							options...,
						),
						nil

				} else if valueB.PackThunk != nil {
					// (FileInfoThunk, PackThunk)
					nameA := valueA.FileInfoThunk.FileInfo.GetName(scope)

					if nameA < valueB.Min {
						valueA.FileInfoThunk.Expand(false)
						return &IterItem{
								ZipItem: &ZipItem{A: valueA, B: nil, Dir: dirA},
							}, zip(
								a,
								func() (*IterItem, Src, error) {
									return &IterItem{
										PackThunk: valueB.PackThunk,
									}, b, nil
								},
								cont,
								options...,
							),
							nil

					} else if nameA > valueB.Max {
						valueB.PackThunk.Expand(false)
						return &IterItem{
								ZipItem: &ZipItem{A: nil, B: valueB, Dir: dirB},
							}, zip(
								func() (*IterItem, Src, error) {
									return &IterItem{
										FileInfoThunk: valueA.FileInfoThunk,
									}, a, nil
								},
								b,
								cont,
								options...,
							),
							nil

					} else {
						// unpack
						valueB.PackThunk.Expand(true)
						return nil,
							zip(
								func() (*IterItem, Src, error) {
									return &IterItem{
										FileInfoThunk: valueA.FileInfoThunk,
									}, a, nil
								},
								b,
								cont,
								options...,
							),
							nil
					}

				}

			} else if valueA.PackThunk != nil {

				if valueB.PackThunk != nil {
					// (PackThunk, PackThunk)
					if valueA.Pack == valueB.Pack {
						// same, do not unpack
						valueA.PackThunk.Expand(false)
						valueB.PackThunk.Expand(false)
						if dirA != dirB {
							panic(fmt.Errorf("bad iter, expecting same path: %v %v", dirA, dirB))
						}
						return &IterItem{
								ZipItem: &ZipItem{A: valueA, B: valueB, Dir: dirA},
							}, zip(
								a,
								b,
								cont,
								options...,
							),
							nil
					}

					if valueA.Max < valueB.Min {
						valueA.PackThunk.Expand(false)
						return &IterItem{
								ZipItem: &ZipItem{A: valueA, B: nil, Dir: dirA},
							}, zip(
								a,
								func() (*IterItem, Src, error) {
									return &IterItem{
										PackThunk: valueB.PackThunk,
									}, b, nil
								},
								cont,
								options...,
							),
							nil

					} else if valueB.Max < valueA.Min {
						valueB.PackThunk.Expand(false)
						return &IterItem{
								ZipItem: &ZipItem{A: nil, B: valueB, Dir: dirB},
							}, zip(
								func() (*IterItem, Src, error) {
									return &IterItem{
										PackThunk: valueA.PackThunk,
									}, a, nil
								},
								b,
								cont,
								options...,
							),
							nil

					} else {
						// unpack
						valueA.PackThunk.Expand(true)
						valueB.PackThunk.Expand(true)
						return nil,
							zip(
								a,
								b,
								cont,
								options...,
							),
							nil
					}

				} else if valueB.FileInfo != nil {
					// (PackThunk, FileInfo)
					nameB := valueB.FileInfo.GetName(scope)

					if nameB < valueA.Min {
						return &IterItem{
								ZipItem: &ZipItem{A: nil, B: valueB, Dir: dirB},
							}, zip(
								func() (*IterItem, Src, error) {
									return &IterItem{
										PackThunk: valueA.PackThunk,
									}, a, nil
								},
								b,
								cont,
								options...,
							),
							nil

					} else if nameB > valueA.Max {
						valueA.PackThunk.Expand(false)
						return &IterItem{
								ZipItem: &ZipItem{A: valueA, B: nil, Dir: dirA},
							}, zip(
								a,
								func() (*IterItem, Src, error) {
									return &IterItem{
										FileInfo: valueB.FileInfo,
									}, b, nil
								},
								cont,
								options...,
							),
							nil

					} else {
						// unpack
						valueA.PackThunk.Expand(true)
						return nil,
							zip(
								a,
								func() (*IterItem, Src, error) {
									return &IterItem{
										FileInfo: valueB.FileInfo,
									}, b, nil
								},
								cont,
								options...,
							),
							nil
					}

				} else if valueB.FileInfoThunk != nil {
					// (PackThunk, FileInfoThunk)
					nameB := valueB.FileInfoThunk.FileInfo.GetName(scope)

					if nameB < valueA.Min {
						valueB.FileInfoThunk.Expand(false)
						return &IterItem{
								ZipItem: &ZipItem{A: nil, B: valueB, Dir: dirB},
							}, zip(
								func() (*IterItem, Src, error) {
									return &IterItem{
										PackThunk: valueA.PackThunk,
									}, a, nil
								},
								b,
								cont,
								options...,
							),
							nil

					} else if nameB > valueA.Max {
						valueA.PackThunk.Expand(false)
						return &IterItem{
								ZipItem: &ZipItem{A: valueA, B: nil, Dir: dirA},
							}, zip(
								a,
								func() (*IterItem, Src, error) {
									return &IterItem{
										FileInfoThunk: valueB.FileInfoThunk,
									}, b, nil
								},
								cont,
								options...,
							),
							nil

					} else {
						// unpack
						valueA.PackThunk.Expand(true)
						return nil,
							zip(
								a,
								func() (*IterItem, Src, error) {
									return &IterItem{
										FileInfoThunk: valueB.FileInfoThunk,
									}, b, nil
								},
								cont,
								options...,
							),
							nil
					}

				}

			}

			panic(fmt.Errorf("not zippable %T %T", valueA, valueB)) // NOCOVER

		}
	}

	return zip

}

func isDeeper(a, b string) bool {
cmp:
	if a != "." && b == "." {
		return true
	} else if a == "." && b != "." {
		return false
	} else if a == "." && b == "." {
		return false
	}
	a = filepath.Dir(a)
	b = filepath.Dir(b)
	goto cmp
}

func TapZipItem(fn func(ZipItem)) Sink {
	var sink Sink
	sink = func(v *IterItem) (Sink, error) {
		if v == nil {
			return nil, nil
		}
		fn(*v.ZipItem)
		return sink, nil
	}
	return sink
}
