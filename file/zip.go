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

type ZipItem struct {
	A   any
	B   any
	Dir string
}

type ZipOption interface {
	IsZipOption()
}

func (_ Def) Zip(
	scope Scope,
) Zip {

	var zip func(a Src, b Src, cont Src, options ...ZipOption) Src
	zip = func(a Src, b Src, cont Src, options ...ZipOption) Src {
		return func() (any, Src, error) {

			valueA, err := a.Next()
			if err != nil { // NOCOVER
				return nil, nil, err
			}
			valueB, err := b.Next()
			if err != nil { // NOCOVER
				return nil, nil, err
			}

			if valueA == nil && valueB == nil {
				return nil, cont, nil
			}

			var dirA, dirB string

			if valueA != nil {
				switch valueA := valueA.(type) {
				case FileInfo:
					dirA = valueA.Path
					dirA = filepath.Dir(dirA)
				case FileInfoThunk:
					dirA = valueA.Path
					dirA = filepath.Dir(dirA)
				case PackThunk:
					dirA = valueA.Path
				default:
					panic(fmt.Errorf("unknown type %T", valueA))
				}
			}
			if valueB != nil {
				switch valueB := valueB.(type) {
				case FileInfo:
					dirB = valueB.Path
					dirB = filepath.Dir(dirB)
				case FileInfoThunk:
					dirB = valueB.Path
					dirB = filepath.Dir(dirB)
				case PackThunk:
					dirB = valueB.Path
				default:
					panic(fmt.Errorf("unknown type %T", valueB))
				}
			}

			if valueB == nil {
				if t, ok := valueA.(FileInfoThunk); ok {
					t.Expand(false)
				} else if t, ok := valueA.(PackThunk); ok {
					t.Expand(false)
				}
				return ZipItem{A: valueA, B: nil, Dir: dirA},
					zip(a, nil, cont, options...),
					nil

			} else if valueA == nil {
				if t, ok := valueB.(FileInfoThunk); ok {
					t.Expand(false)
				} else if t, ok := valueB.(PackThunk); ok {
					t.Expand(false)
				}
				return ZipItem{A: nil, B: valueB, Dir: dirB},
					zip(nil, b, cont, options...),
					nil
			}

			if isDeeper(dirA, dirB) {
				return ZipItem{A: valueA, B: nil, Dir: dirA},
					zip(
						a,
						func() (any, Src, error) {
							return valueB, b, nil
						},
						cont,
						options...,
					),
					nil

			} else if isDeeper(dirB, dirA) {
				return ZipItem{A: nil, B: valueB, Dir: dirB},
					zip(
						func() (any, Src, error) {
							return valueA, a, nil
						},
						b,
						cont,
						options...,
					),
					nil
			}

			switch valueA := valueA.(type) {

			case FileInfo:
				switch valueB := valueB.(type) {

				case FileInfo:
					// (FileInfo, FileInfo)
					nameA := valueA.GetName(scope)
					nameB := valueB.GetName(scope)
					if nameA < nameB {
						return ZipItem{A: valueA, B: nil, Dir: dirA},
							zip(
								a,
								func() (any, Src, error) {
									return valueB, b, nil
								},
								cont,
								options...,
							),
							nil

					} else if nameB < nameA {
						return ZipItem{A: nil, B: valueB, Dir: dirB},
							zip(
								func() (any, Src, error) {
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
					return ZipItem{A: valueA, B: valueB, Dir: dirA},
						zip(
							a,
							b,
							cont,
							options...,
						),
						nil

				case FileInfoThunk:
					// (FileInfo, FileInfoThunk)
					nameA := valueA.GetName(scope)
					nameB := valueB.FileInfo.GetName(scope)
					if nameA < nameB {
						return ZipItem{A: valueA, B: nil, Dir: dirA},
							zip(
								a,
								func() (any, Src, error) {
									return valueB, b, nil
								},
								cont,
								options...,
							),
							nil

					} else if nameB < nameA {
						valueB.Expand(false)
						return ZipItem{A: nil, B: valueB, Dir: dirB},
							zip(
								func() (any, Src, error) {
									return valueA, a, nil
								},
								b,
								cont,
								options...,
							),
							nil
					}

					valueB.Expand(true)
					return nil,
						zip(
							func() (any, Src, error) {
								return valueA, a, nil
							},
							func() (any, Src, error) {
								return valueB.FileInfo, b, nil
							},
							cont,
							options...,
						),
						nil

				case PackThunk:
					// (FileInfo, PackThunk)
					nameA := valueA.GetName(scope)

					if nameA < valueB.Min {
						return ZipItem{A: valueA, B: nil, Dir: dirA},
							zip(
								a,
								func() (any, Src, error) {
									return valueB, b, nil
								},
								cont,
								options...,
							),
							nil

					} else if nameA > valueB.Max {
						valueB.Expand(false)
						return ZipItem{A: nil, B: valueB, Dir: dirB},
							zip(
								func() (any, Src, error) {
									return valueA, a, nil
								},
								b,
								cont,
								options...,
							),
							nil

					} else {
						// unpack
						valueB.Expand(true)
						return nil,
							zip(
								func() (any, Src, error) {
									return valueA, a, nil
								},
								b,
								cont,
								options...,
							),
							nil
					}

				}

			case FileInfoThunk:
				switch valueB := valueB.(type) {

				case FileInfoThunk:
					// (FileInfoThunk, FileInfoThunk)
					nameA := valueA.FileInfo.GetName(scope)
					nameB := valueB.FileInfo.GetName(scope)
					if nameA < nameB {
						valueA.Expand(false)
						return ZipItem{A: valueA, B: nil, Dir: dirA},
							zip(
								a,
								func() (any, Src, error) {
									return valueB, b, nil
								},
								cont,
								options...,
							),
							nil

					} else if nameB < nameA {
						valueB.Expand(false)
						return ZipItem{A: nil, B: valueB, Dir: dirB},
							zip(
								func() (any, Src, error) {
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
						res, err := predictExpand(valueA, valueB)
						if err != nil {
							return nil, nil, err
						}
						expand = res
					}

					if expand {
						// iter subs
						valueA.Expand(true)
						valueB.Expand(true)
						return nil,
							zip(
								func() (any, Src, error) {
									return valueA.FileInfo, a, nil
								},
								func() (any, Src, error) {
									return valueB.FileInfo, b, nil
								},
								cont,
								options...,
							),
							nil

					} else {
						valueA.Expand(false)
						valueB.Expand(false)
						return ZipItem{A: valueA, B: valueB, Dir: dirA},
							zip(
								a,
								b,
								cont,
								options...,
							),
							nil
					}

				case FileInfo:
					// (FileInfoThunk, FileInfo)
					nameA := valueA.FileInfo.GetName(scope)
					nameB := valueB.GetName(scope)
					if nameA < nameB {
						valueA.Expand(false)
						return ZipItem{A: valueA, B: nil, Dir: dirA},
							zip(
								a,
								func() (any, Src, error) {
									return valueB, b, nil
								},
								cont,
								options...,
							),
							nil

					} else if nameB < nameA {
						return ZipItem{A: nil, B: valueB, Dir: dirB},
							zip(
								func() (any, Src, error) {
									return valueA, a, nil
								},
								b,
								cont,
								options...,
							),
							nil
					}

					valueA.Expand(true)
					return nil,
						zip(
							func() (any, Src, error) {
								return valueA.FileInfo, a, nil
							},
							func() (any, Src, error) {
								return valueB, b, nil
							},
							cont,
							options...,
						),
						nil

				case PackThunk:
					// (FileInfoThunk, PackThunk)
					nameA := valueA.FileInfo.GetName(scope)

					if nameA < valueB.Min {
						valueA.Expand(false)
						return ZipItem{A: valueA, B: nil, Dir: dirA},
							zip(
								a,
								func() (any, Src, error) {
									return valueB, b, nil
								},
								cont,
								options...,
							),
							nil

					} else if nameA > valueB.Max {
						valueB.Expand(false)
						return ZipItem{A: nil, B: valueB, Dir: dirB},
							zip(
								func() (any, Src, error) {
									return valueA, a, nil
								},
								b,
								cont,
								options...,
							),
							nil

					} else {
						// unpack
						valueB.Expand(true)
						return nil,
							zip(
								func() (any, Src, error) {
									return valueA, a, nil
								},
								b,
								cont,
								options...,
							),
							nil
					}

				}

			case PackThunk:
				switch valueB := valueB.(type) {

				case PackThunk:
					// (PackThunk, PackThunk)
					if valueA.Pack == valueB.Pack {
						// same, do not unpack
						valueA.Expand(false)
						valueB.Expand(false)
						if dirA != dirB {
							panic(fmt.Errorf("bad iter, expecting same path: %v %v", dirA, dirB))
						}
						return ZipItem{A: valueA, B: valueB, Dir: dirA},
							zip(
								a,
								b,
								cont,
								options...,
							),
							nil
					}

					if valueA.Max < valueB.Min {
						valueA.Expand(false)
						return ZipItem{A: valueA, B: nil, Dir: dirA},
							zip(
								a,
								func() (any, Src, error) {
									return valueB, b, nil
								},
								cont,
								options...,
							),
							nil

					} else if valueB.Max < valueA.Min {
						valueB.Expand(false)
						return ZipItem{A: nil, B: valueB, Dir: dirB},
							zip(
								func() (any, Src, error) {
									return valueA, a, nil
								},
								b,
								cont,
								options...,
							),
							nil

					} else {
						// unpack
						valueA.Expand(true)
						valueB.Expand(true)
						return nil,
							zip(
								a,
								b,
								cont,
								options...,
							),
							nil
					}

				case FileInfo:
					// (PackThunk, FileInfo)
					nameB := valueB.GetName(scope)

					if nameB < valueA.Min {
						return ZipItem{A: nil, B: valueB, Dir: dirB},
							zip(
								func() (any, Src, error) {
									return valueA, a, nil
								},
								b,
								cont,
								options...,
							),
							nil

					} else if nameB > valueA.Max {
						valueA.Expand(false)
						return ZipItem{A: valueA, B: nil, Dir: dirA},
							zip(
								a,
								func() (any, Src, error) {
									return valueB, b, nil
								},
								cont,
								options...,
							),
							nil

					} else {
						// unpack
						valueA.Expand(true)
						return nil,
							zip(
								a,
								func() (any, Src, error) {
									return valueB, b, nil
								},
								cont,
								options...,
							),
							nil
					}

				case FileInfoThunk:
					// (PackThunk, FileInfoThunk)
					nameB := valueB.FileInfo.GetName(scope)

					if nameB < valueA.Min {
						valueB.Expand(false)
						return ZipItem{A: nil, B: valueB, Dir: dirB},
							zip(
								func() (any, Src, error) {
									return valueA, a, nil
								},
								b,
								cont,
								options...,
							),
							nil

					} else if nameB > valueA.Max {
						valueA.Expand(false)
						return ZipItem{A: valueA, B: nil, Dir: dirA},
							zip(
								a,
								func() (any, Src, error) {
									return valueB, b, nil
								},
								cont,
								options...,
							),
							nil

					} else {
						// unpack
						valueA.Expand(true)
						return nil,
							zip(
								a,
								func() (any, Src, error) {
									return valueB, b, nil
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
	sink = func(v any) (Sink, error) {
		if v == nil {
			return nil, nil
		}
		fn(v.(ZipItem))
		return sink, nil
	}
	return sink
}
