// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package index

func Has(
	index Index,
	args ...SelectOption) (bool, error) {
	var count int
	args = append(args, Count(&count))
	args = append(args, Limit(1))
	if err := Select(index, args...); err != nil {
		return false, err
	}
	return count > 0, nil
}
