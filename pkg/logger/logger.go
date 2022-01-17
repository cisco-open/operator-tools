// Copyright © 2020 Banzai Cloud
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package logger

import (
	"fmt"

	"github.com/go-logr/logr"
)

var (
	GlobalLogLevel = 1

	Log = New()
)

type GroupedLogger interface {
	Grouped(state bool)
}

type Logger interface {
	Enabled() bool
	Info(msg string, keysAndValues ...interface{})
	Error(err error, msg string, keysAndValues ...interface{})
	V(level int) Logger
	WithValues(keysAndValues ...interface{}) Logger
	WithName(name string) Logger

	GetLogrLogger() logr.Logger
}

type logger struct {
	level int
	sink  *spinnerLogSink
}

func New(options ...Option) Logger {
	sink := newSpinnerLogSink(options...)
	return &logger{
		level: 0,
		sink:  sink,
	}
}

func (log *logger) Info(msg string, keysAndValues ...interface{}) {
	log.sink.Info(log.level, msg, keysAndValues)
}

func (log *logger) Enabled() bool {
	return log.sink.Enabled(log.level)
}

func (log *logger) Error(e error, msg string, keysAndValues ...interface{}) {
	log.sink.Error(e, msg, keysAndValues)
}

func (log *logger) V(level int) Logger {
	return &logger{
		level: level,
		sink:  log.sink.copyLogger(),
	}
}

func (log *logger) WithName(name string) Logger {
	sink := log.sink.copyLogger()
	sink.names = append(sink.names, name)

	return &logger{
		level: log.level,
		sink:  sink,
	}
}

func (log *logger) WithValues(keysAndValues ...interface{}) Logger {
	sink := log.sink.copyLogger()
	for k, v := range keysAndValues {
		sink.values[k] = v
	}

	return &logger{
		level: log.level,
		sink:  sink,
	}
}

func (log *logger) GetLogrLogger() logr.Logger {
	sink := log.sink.copyLogger()
	return logr.New(sink).V(log.level)
}

func (log *logger) Plain(msg string) {
	if GlobalLogLevel >= log.level {
		fmt.Println(fmt.Sprint(msg))
	}
}

func (log *logger) Plainf(format string, args ...interface{}) {
	if GlobalLogLevel >= log.level {
		fmt.Println(fmt.Sprintf(format, args...))
	}
}

func (log *logger) SetOptions(options ...Option) {
	log.sink.SetOptions(options...)
}

func (log *logger) ShowTime(f bool) Logger {
	sink := log.sink.copyLogger()
	sink.showTime = f

	return &logger{
		level: log.level,
		sink:  sink,
	}
}

func (log *logger) Grouped(state bool) {
	if !log.sink.grouppable {
		return
	}

	if state && log.sink.spinner == nil {
		log.sink.initSpinner()
	} else if !state && log.sink.spinner != nil {
		log.sink.stopSpinner()
	}

	log.sink.grouped = state
}

func EnableGroupSession(logger interface{}) func() {
	if l, ok := logger.(interface{ Grouped(state bool) }); ok {
		l.Grouped(true)
		return func() { l.Grouped(false) }
	}
	return func() {}
}
