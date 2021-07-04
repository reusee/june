// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package opts

import "reflect"

func Extract(options any, targets ...any) {
	typeMap := make(map[reflect.Type]int)
	var targetValues []reflect.Value
	for i, target := range targets {
		targetValue := reflect.ValueOf(target)
		targetValues = append(targetValues, targetValue)
		t := targetValue.Type().Elem()
		typeMap[t] = i
	}
	optionsValue := reflect.ValueOf(options)
	for i, l := 0, optionsValue.Len(); i < l; i++ {
		value := optionsValue.Index(i)
		if idx, ok := typeMap[value.Type()]; ok {
			targetValues[idx].Elem().Set(value)
		}
	}
}
