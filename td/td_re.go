// Copyright (c) 2018, Maxime Soulé
// All rights reserved.
//
// This source code is licensed under the BSD-style license found in the
// LICENSE file in the root directory of this source tree.

package td

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"

	"github.com/maxatome/go-testdeep/internal/color"
	"github.com/maxatome/go-testdeep/internal/ctxerr"
	"github.com/maxatome/go-testdeep/internal/dark"
	"github.com/maxatome/go-testdeep/internal/types"
)

type tdRe struct {
	base
	re         *regexp.Regexp
	captures   reflect.Value
	numMatches int
}

var _ TestDeep = &tdRe{}

func newRe(regIf interface{}, capture ...interface{}) (*tdRe, error) {
	r := &tdRe{
		base: newBase(4),
	}

	const usage = "(STRING|*regexp.Regexp[, NON_NIL_CAPTURE])"

	switch len(capture) {
	case 0:
	case 1:
		if capture[0] != nil {
			r.captures = reflect.ValueOf(capture[0])
		}
	default:
		return nil, errors.New(color.TooManyParams(r.location.Func + usage))
	}

	switch reg := regIf.(type) {
	case *regexp.Regexp:
		r.re = reg
	case string:
		r.re = regexp.MustCompile(reg)
	default:
		return nil, errors.New(color.BadUsage(r.location.Func+usage, regIf, 1, false))
	}
	return r, nil
}

// summary(Re): allows to apply a regexp on a string (or convertible),
// []byte, error or fmt.Stringer interfaces, and even test the
// captured groups
// input(Re): str,slice([]byte),if(✓ + fmt.Stringer/error)

// Re operator allows to apply a regexp on a string (or convertible),
// []byte, error or fmt.Stringer interface (error interface is tested
// before fmt.Stringer.)
//
// "reg" is the regexp. It can be a string that is automatically
// compiled using regexp.MustCompile, or a *regexp.Regexp.
//
// Optional "capture" parameter can be used to match the contents of
// regexp groups. Groups are presented as a []string or [][]byte
// depending the original matched data. Note that an other operator
// can be used here.
//
//   td.Cmp(t, "foobar zip!", td.Re(`^foobar`)) // succeeds
//   td.Cmp(t, "John Doe",
//     td.Re(`^(\w+) (\w+)`, []string{"John", "Doe"})) // succeeds
//   td.Cmp(t, "John Doe",
//     td.Re(`^(\w+) (\w+)`, td.Bag("Doe", "John"))) // succeeds
func Re(reg interface{}, capture ...interface{}) TestDeep {
	r, err := newRe(reg, capture...)
	if err != nil {
		f := dark.GetFatalizer()
		f.Helper()
		f.Fatal(err)
	}
	r.numMatches = 1
	return r
}

// summary(ReAll): allows to successively apply a regexp on a string
// (or convertible), []byte, error or fmt.Stringer interfaces, and
// even test the captured groups
// input(ReAll): str,slice([]byte),if(✓ + fmt.Stringer/error)

// ReAll operator allows to successively apply a regexp on a string
// (or convertible), []byte, error or fmt.Stringer interface (error
// interface is tested before fmt.Stringer) and to match its groups
// contents.
//
// "reg" is the regexp. It can be a string that is automatically
// compiled using regexp.MustCompile, or a *regexp.Regexp.
//
// "capture" is used to match the contents of regexp groups. Groups
// are presented as a []string or [][]byte depending the original
// matched data. Note that an other operator can be used here.
//
//   td.Cmp(t, "John Doe",
//     td.ReAll(`(\w+)(?: |\z)`, []string{"John", "Doe"})) // succeeds
//   td.Cmp(t, "John Doe",
//     td.ReAll(`(\w+)(?: |\z)`, td.Bag("Doe", "John"))) // succeeds
func ReAll(reg, capture interface{}) TestDeep {
	r, err := newRe(reg, capture)
	if err != nil {
		f := dark.GetFatalizer()
		f.Helper()
		f.Fatal(err)
	}
	r.numMatches = -1
	return r
}

func (r *tdRe) needCaptures() bool {
	return r.captures.IsValid()
}

func (r *tdRe) matchByteCaptures(ctx ctxerr.Context, got []byte, result [][][]byte) *ctxerr.Error {
	if len(result) == 0 {
		return r.doesNotMatch(ctx, got)
	}

	num := 0
	for _, set := range result {
		num += len(set) - 1
	}

	// Not perfect but cast captured groups to string

	// Special case to accepted expected []interface{} type
	if r.captures.Type() == types.SliceInterface {
		captures := make([]interface{}, 0, num)
		for _, set := range result {
			for _, match := range set[1:] {
				captures = append(captures, string(match))
			}
		}
		return r.matchCaptures(ctx, captures)
	}

	captures := make([]string, 0, num)
	for _, set := range result {
		for _, match := range set[1:] {
			captures = append(captures, string(match))
		}
	}
	return r.matchCaptures(ctx, captures)
}

func (r *tdRe) matchStringCaptures(ctx ctxerr.Context, got string, result [][]string) *ctxerr.Error {
	if len(result) == 0 {
		return r.doesNotMatch(ctx, got)
	}

	num := 0
	for _, set := range result {
		num += len(set) - 1
	}

	// Special case to accepted expected []interface{} type
	if r.captures.Type() == types.SliceInterface {
		captures := make([]interface{}, 0, num)
		for _, set := range result {
			for _, match := range set[1:] {
				captures = append(captures, match)
			}
		}
		return r.matchCaptures(ctx, captures)
	}

	captures := make([]string, 0, num)
	for _, set := range result {
		captures = append(captures, set[1:]...)
	}
	return r.matchCaptures(ctx, captures)
}

func (r *tdRe) matchCaptures(ctx ctxerr.Context, captures interface{}) (err *ctxerr.Error) {
	return deepValueEqual(
		ctx.ResetPath("("+ctx.Path.String()+" =~ "+r.String()+")"),
		reflect.ValueOf(captures), r.captures)
}

func (r *tdRe) matchBool(ctx ctxerr.Context, got interface{}, result bool) *ctxerr.Error {
	if result {
		return nil
	}
	return r.doesNotMatch(ctx, got)
}

func (r *tdRe) doesNotMatch(ctx ctxerr.Context, got interface{}) *ctxerr.Error {
	if ctx.BooleanError {
		return ctxerr.BooleanError
	}
	return ctx.CollectError(&ctxerr.Error{
		Message:  "does not match Regexp",
		Got:      got,
		Expected: types.RawString(r.re.String()),
	})
}

func (r *tdRe) Match(ctx ctxerr.Context, got reflect.Value) *ctxerr.Error {
	var str string
	switch got.Kind() {
	case reflect.String:
		str = got.String()

	case reflect.Slice:
		if got.Type().Elem().Kind() == reflect.Uint8 {
			gotBytes := got.Bytes()
			if r.needCaptures() {
				return r.matchByteCaptures(ctx,
					gotBytes, r.re.FindAllSubmatch(gotBytes, r.numMatches))
			}
			return r.matchBool(ctx, gotBytes, r.re.Match(gotBytes))
		}
		if ctx.BooleanError {
			return ctxerr.BooleanError
		}
		return ctx.CollectError(&ctxerr.Error{
			Message:  "bad slice type",
			Got:      types.RawString("[]" + got.Type().Elem().Kind().String()),
			Expected: types.RawString("[]uint8"),
		})

	default:
		var strOK bool
		iface := dark.MustGetInterface(got)

		switch gotVal := iface.(type) {
		case error:
			str = gotVal.Error()
			strOK = true
		case fmt.Stringer:
			str = gotVal.String()
			strOK = true
		default:
		}

		if !strOK {
			if ctx.BooleanError {
				return ctxerr.BooleanError
			}
			return ctx.CollectError(&ctxerr.Error{
				Message: "bad type",
				Got:     types.RawString(got.Type().String()),
				Expected: types.RawString(
					"string (convertible) OR fmt.Stringer OR error OR []uint8"),
			})
		}
	}

	if r.needCaptures() {
		return r.matchStringCaptures(ctx,
			str, r.re.FindAllStringSubmatch(str, r.numMatches))
	}
	return r.matchBool(ctx, str, r.re.MatchString(str))
}

func (r *tdRe) String() string {
	return r.re.String()
}
