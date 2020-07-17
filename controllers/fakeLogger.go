package controllers

import (
	"github.com/go-logr/logr"
)

type FakeLogger struct {
	FakeInfoLogger
	ErrorFn        func(err error, msg string, keysAndValues ...interface{})
	ErrorCallCount int
}

type FakeInfoLogger struct {
	InfoFn        func(msg string, keysAndValues ...interface{})
	InfoCallCount int
}

func (f *FakeInfoLogger) Info(msg string, keysAndValues ...interface{}) {
	f.InfoCallCount++
	f.InfoFn(msg, keysAndValues)
}

func (*FakeInfoLogger) Enabled() bool {
	panic("NI")
}

func (f *FakeLogger) Error(err error, msg string, keysAndValues ...interface{}) {
	f.ErrorCallCount++
	f.ErrorFn(err, msg, keysAndValues)
}

func (*FakeLogger) V(level int) logr.InfoLogger {
	panic("NI")
}

func (*FakeLogger) WithValues(keysAndValues ...interface{}) logr.Logger {
	panic("NI")
}

func (*FakeLogger) WithName(name string) logr.Logger {
	panic("NI")
}
