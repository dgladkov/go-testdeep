// Copyright (c) 2018, Maxime Soulé
// All rights reserved.
//
// This source code is licensed under the BSD-style license found in the
// LICENSE file in the root directory of this source tree.

package test

import (
	"fmt"
	"runtime"
	"testing"
)

// TestingT is a type implementing td.TestingT intended to be used in
// tests.
type TestingT struct {
	Messages  []string
	IsFatal   bool
	HasFailed bool
}

// NewTestingT returns a new instance of *TestingT.
func NewTestingT() *TestingT {
	return &TestingT{}
}

// Error mocks testing.T Error method.
func (t *TestingT) Error(args ...interface{}) {
	t.Messages = append(t.Messages, fmt.Sprint(args...))
	t.IsFatal = false
	t.HasFailed = true
}

// Fatal mocks testing.T Fatal method.
func (t *TestingT) Fatal(args ...interface{}) {
	t.Messages = append(t.Messages, fmt.Sprint(args...))
	t.IsFatal = true
	t.HasFailed = true
}

// Helper mocks testing.T Helper method.
func (t *TestingT) Helper() {
	// Do nothing
}

// LastMessage returns the last message.
func (t *TestingT) LastMessage() string {
	if len(t.Messages) == 0 {
		return ""
	}
	return t.Messages[len(t.Messages)-1]
}

// ResetMessages resets the messages.
func (t *TestingT) ResetMessages() {
	t.Messages = t.Messages[:0]
}

// TestingTB is a type implementing testing.TB intended to be used in
// tests.
type TestingTB struct {
	TestingT
	name string
	testing.TB
	cleanup func()
}

// NewTestingTB returns a new instance of *TestingTB.
func NewTestingTB(name string) *TestingTB {
	return &TestingTB{name: name}
}

// Cleanup mocks testing.T Cleanup method. Not thread-safe but we
// don't care in tests.
func (t *TestingTB) Cleanup(fn func()) {
	old := t.cleanup
	t.cleanup = func() {
		if old != nil {
			defer old()
		}
		fn()
	}
	runtime.SetFinalizer(t, func(t *TestingTB) { t.cleanup() })
}

// Fatal mocks testing.T Error method.
func (t *TestingTB) Error(args ...interface{}) {
	t.TestingT.Error(args...)
}

// Errorf mocks testing.T Errorf method.
func (t *TestingTB) Errorf(format string, args ...interface{}) {
	t.TestingT.Error(fmt.Sprintf(format, args...))
}

// Fail mocks testing.T Fail method.
func (t *TestingTB) Fail() {
	t.HasFailed = true
}

// FailNow mocks testing.T FailNow method.
func (t *TestingTB) FailNow() {
	t.HasFailed = true
	t.IsFatal = true
}

// Failed mocks testing.T Failed method.
func (t *TestingTB) Failed() bool {
	return t.HasFailed
}

// Fatal mocks testing.T Fatal method.
func (t *TestingTB) Fatal(args ...interface{}) {
	t.TestingT.Fatal(args...)
}

// Fatalf mocks testing.T Fatalf method.
func (t *TestingTB) Fatalf(format string, args ...interface{}) {
	t.TestingT.Fatal(fmt.Sprintf(format, args...))
}

// Helper mocks testing.T Helper method.
func (t *TestingTB) Helper() {
	// Do nothing
}

// Log mocks testing.T Log method.
func (t *TestingTB) Log(args ...interface{}) {
	t.Messages = append(t.Messages, fmt.Sprint(args...))
}

// Logf mocks testing.T Logf method.
func (t *TestingTB) Logf(format string, args ...interface{}) {
	t.Log(fmt.Sprintf(format, args...))
}

// Name mocks testing.T Name method.
func (t *TestingTB) Name() string {
	return t.name
}

// Skip mocks testing.T Skip method.
func (t *TestingTB) Skip(args ...interface{}) {}

// SkipNow mocks testing.T SkipNow method.
func (t *TestingTB) SkipNow() {}

// Skipf mocks testing.T Skipf method.
func (t *TestingTB) Skipf(format string, args ...interface{}) {}

// Skipped mocks testing.T Skipped method.
func (t *TestingTB) Skipped() bool {
	return false
}
