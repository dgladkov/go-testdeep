// Copyright (c) 2018, Maxime Soulé
// All rights reserved.
//
// This source code is licensed under the BSD-style license found in the
// LICENSE file in the root directory of this source tree.

package td

import (
	"bytes"
	"errors"
	"fmt"
	"reflect"

	"github.com/maxatome/go-testdeep/internal/color"
	"github.com/maxatome/go-testdeep/internal/ctxerr"
	"github.com/maxatome/go-testdeep/internal/dark"
	"github.com/maxatome/go-testdeep/internal/types"
	"github.com/maxatome/go-testdeep/internal/util"
)

type tdArray struct {
	tdExpectedType
	expectedEntries []reflect.Value
}

var _ TestDeep = &tdArray{}

// ArrayEntries allows to pass array or slice entries to check in
// functions Array and Slice. It is a map whose each key is the item
// index and the corresponding value the expected item value (which
// can be a TestDeep operator as well as a zero value).
type ArrayEntries map[int]interface{}

func newArray(kind reflect.Kind, model interface{}, expectedEntries ArrayEntries) (*tdArray, error) {
	vmodel := reflect.ValueOf(model)

	a := tdArray{
		tdExpectedType: tdExpectedType{
			base: newBase(4),
		},
	}

	switch vmodel.Kind() {
	case reflect.Ptr:
		if vmodel.Type().Elem().Kind() != kind {
			break
		}

		a.isPtr = true

		if vmodel.IsNil() {
			a.expectedType = vmodel.Type().Elem()
			return &a, a.populateExpectedEntries(expectedEntries, reflect.Value{})
		}

		vmodel = vmodel.Elem()
		fallthrough

	case kind:
		a.expectedType = vmodel.Type()
		return &a, a.populateExpectedEntries(expectedEntries, vmodel)
	}

	return nil, nil
}

// summary(Array): compares the contents of an array or a pointer on an array
// input(Array): array,ptr(ptr on array)

// Array operator compares the contents of an array or a pointer on an
// array against the non-zero values of "model" (if any) and the
// values of "expectedEntries".
//
// "model" must be the same type as compared data.
//
// "expectedEntries" can be nil, if no zero entries are expected and
// no TestDeep operator are involved.
//
//   got := [3]int{12, 14, 17}
//   td.Cmp(t, got, td.Array([3]int{0, 14}, td.ArrayEntries{0: 12, 2: 17})) // succeeds
//   td.Cmp(t, got,
//     td.Array([3]int{0, 14}, td.ArrayEntries{0: td.Gt(10), 2: td.Gt(15)})) // succeeds
//
// TypeBehind method returns the reflect.Type of "model".
func Array(model interface{}, expectedEntries ArrayEntries) TestDeep {
	a, err := newArray(reflect.Array, model, expectedEntries)
	if err != nil || a == nil {
		f := dark.GetFatalizer()
		f.Helper()
		if err != nil {
			f.Fatal(err)
		}
		f.Fatal(color.BadUsage("Array(ARRAY|&ARRAY, EXPECTED_ENTRIES)", model, 1, true))
	}
	return a
}

// summary(Slice): compares the contents of a slice or a pointer on a slice
// input(Slice): slice,ptr(ptr on slice)

// Slice operator compares the contents of a slice or a pointer on a
// slice against the non-zero values of "model" (if any) and the
// values of "expectedEntries".
//
// "model" must be the same type as compared data.
//
// "expectedEntries" can be nil, if no zero entries are expected and
// no TestDeep operator are involved.
//
//   got := []int{12, 14, 17}
//   td.Cmp(t, got, td.Slice([]int{0, 14}, td.ArrayEntries{0: 12, 2: 17})) // succeeds
//   td.Cmp(t, got,
//     td.Slice([]int{0, 14}, td.ArrayEntries{0: td.Gt(10), 2: td.Gt(15)})) // succeeds
//
// TypeBehind method returns the reflect.Type of "model".
func Slice(model interface{}, expectedEntries ArrayEntries) TestDeep {
	a, err := newArray(reflect.Slice, model, expectedEntries)
	if err != nil || a == nil {
		f := dark.GetFatalizer()
		f.Helper()
		if err != nil {
			f.Fatal(err)
		}
		f.Fatal(color.BadUsage("Slice(SLICE|&SLICE, EXPECTED_ENTRIES)", model, 1, true))
	}
	return a
}

func (a *tdArray) populateExpectedEntries(expectedEntries ArrayEntries, expectedModel reflect.Value) error {
	var maxLength, numEntries int

	maxIndex := -1
	for index := range expectedEntries {
		if index > maxIndex {
			maxIndex = index
		}
	}

	if a.expectedType.Kind() == reflect.Array {
		maxLength = a.expectedType.Len()

		if maxLength <= maxIndex {
			return errors.New(color.Bad(
				"array length is %d, so cannot have #%d expected index",
				maxLength,
				maxIndex))
		}
		numEntries = maxLength
	} else {
		maxLength = -1

		numEntries = maxIndex + 1

		// If slice is non-nil
		if expectedModel.IsValid() {
			if numEntries < expectedModel.Len() {
				numEntries = expectedModel.Len()
			}
		}
	}

	a.expectedEntries = make([]reflect.Value, numEntries)

	elemType := a.expectedType.Elem()
	var vexpectedValue reflect.Value
	for index, expectedValue := range expectedEntries {
		if expectedValue == nil {
			switch elemType.Kind() {
			case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map,
				reflect.Ptr, reflect.Slice:
				vexpectedValue = reflect.Zero(elemType) // change to a typed nil
			default:
				return errors.New(color.Bad(
					"expected value of #%d cannot be nil as items type is %s",
					index,
					elemType))
			}
		} else {
			vexpectedValue = reflect.ValueOf(expectedValue)

			if _, ok := expectedValue.(TestDeep); !ok {
				if !vexpectedValue.Type().AssignableTo(elemType) {
					return errors.New(color.Bad(
						"type %s of #%d expected value differs from %s contents (%s)",
						vexpectedValue.Type(),
						index,
						util.TernStr(maxLength < 0, "slice", "array"),
						elemType))
				}
			}
		}

		a.expectedEntries[index] = vexpectedValue
	}

	vzero := reflect.Zero(elemType)
	// Check initialized entries in model
	if expectedModel.IsValid() {
		zero := vzero.Interface()
		for index := expectedModel.Len() - 1; index >= 0; index-- {
			ventry := expectedModel.Index(index)

			// Entry already expected
			if _, ok := expectedEntries[index]; ok {
				// If non-zero entry, consider it as an error (= 2 expected
				// values for the same item)
				if !reflect.DeepEqual(zero, ventry.Interface()) {
					return errors.New(color.Bad(
						"non zero #%d entry in model already exists in expectedEntries",
						index))
				}
				continue
			}

			a.expectedEntries[index] = ventry
		}
	} else if a.expectedType.Kind() == reflect.Slice {
		// nil slice
		return nil
	}

	var index int

	// Array case, all is OK
	if maxLength >= 0 {
		// Non-nil array => a.expectedEntries already fully initialized
		if expectedModel.IsValid() {
			return nil
		}
		// nil array => a.expectedEntries must be initialized from index=0
		// to numEntries - 1 below
	} else {
		// Non-nil slice => a.expectedEntries must be initialized from
		// index=len(slice) to last entry index of expectedEntries
		index = expectedModel.Len()
	}

	// Slice case, initialize missing expected items to zero
	for ; index < numEntries; index++ {
		if _, ok := expectedEntries[index]; !ok {
			a.expectedEntries[index] = vzero
		}
	}
	return nil
}

func (a *tdArray) Match(ctx ctxerr.Context, got reflect.Value) (err *ctxerr.Error) {
	err = a.checkPtr(ctx, &got, true)
	if err != nil {
		return ctx.CollectError(err)
	}

	err = a.checkType(ctx, got)
	if err != nil {
		return ctx.CollectError(err)
	}

	gotLen := got.Len()
	for index, expectedValue := range a.expectedEntries {
		curCtx := ctx.AddArrayIndex(index)

		if index >= gotLen {
			if ctx.BooleanError {
				return ctxerr.BooleanError
			}
			return curCtx.CollectError(&ctxerr.Error{
				Message:  "expected value out of range",
				Got:      types.RawString("<non-existent value>"),
				Expected: expectedValue,
			})
		}

		err = deepValueEqual(curCtx, got.Index(index), expectedValue)
		if err != nil {
			return
		}
	}

	if gotLen > len(a.expectedEntries) {
		if ctx.BooleanError {
			return ctxerr.BooleanError
		}
		return ctx.AddArrayIndex(len(a.expectedEntries)).CollectError(&ctxerr.Error{
			Message:  "got value out of range",
			Got:      got.Index(len(a.expectedEntries)),
			Expected: types.RawString("<non-existent value>"),
		})
	}

	return nil
}

func (a *tdArray) String() string {
	buf := bytes.NewBufferString(util.TernStr(a.expectedType.Kind() == reflect.Array,
		"Array(", "Slice("))

	buf.WriteString(a.expectedTypeStr())

	if len(a.expectedEntries) == 0 {
		buf.WriteString("{})")
	} else {
		buf.WriteString("{\n")

		for index, expectedValue := range a.expectedEntries {
			fmt.Fprintf(buf, "  %d: %s\n", // nolint: errcheck
				index, util.ToString(expectedValue))
		}

		buf.WriteString("})")
	}
	return buf.String()
}
