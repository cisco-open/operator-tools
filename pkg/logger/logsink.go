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
	"io"
	"os"
	"strings"
	"sync"
	"time"

	"emperror.dev/errors"
	"github.com/briandowns/spinner"
	"github.com/fatih/color"
	"github.com/go-logr/logr"
	"github.com/spf13/cast"
	terminal "github.com/wayneashleyberry/terminal-dimensions"
)

type spinnerLogSink struct {
	names  []string
	values []interface{}
	out    io.Writer
	err    io.Writer

	grouppable         bool
	truncate           bool
	showTime           bool
	grouped            bool
	checkMark          rune
	errorMark          rune
	separatorCharacter rune
	timeFormat         string
	colors             Colors

	spinner *spinner.Spinner

	mux sync.Mutex
}

func NewSpinnerLogSink(options ...Option) *spinnerLogSink {
	l := &spinnerLogSink{
		names: []string{},
		out:   os.Stderr,
		err:   os.Stderr,

		checkMark:          '✓',
		errorMark:          '✗',
		separatorCharacter: '❯',
		colors: Colors{
			Info:  color.FgGreen,
			Error: color.FgRed,
			Key:   color.FgHiGreen,
		},

		mux: sync.Mutex{},
	}

	for _, opt := range options {
		opt(l)
	}

	return l
}

func (log *spinnerLogSink) Init(_ logr.RuntimeInfo) {}

// Info implements logr.LogSink interface
func (log *spinnerLogSink) Info(level int, msg string, keysAndValues ...interface{}) {
	if !log.Enabled(level) {
		return
	}
	allVal := append(keysAndValues, log.values...)
	if len(allVal) > 0 {
		msg = fmt.Sprintf("%s %s", msg, log.joinAndSeparatePairs(allVal))
	}

	names := log.printNames()
	if names != "" {
		msg = fmt.Sprintf("%s %c %s", names, log.separatorCharacter, msg)
	}

	if log.timeFormat != "" && log.showTime {
		msg = fmt.Sprintf("[%s] %s", time.Now().Format(log.timeFormat), msg)
	}

	if log.truncate {
		w, _ := terminal.Width()
		// vscode debug window returns 0 width (without an error)
		if w > 3 {
			w -= 3 // reduced by 3 (spinner char, space, leeway)
			msg = log.truncateString(msg, int(w))
		}
	}

	if log.spinner == nil {
		log.initSpinner()
	}

	log.mux.Lock()
	if log.spinner != nil {
		log.spinner.Writer = log.out
		log.spinner.Suffix = " " + msg // Append text after the spinner
		log.spinner.FinalMSG = color.New(log.colors.Info).Sprintf("%c", log.checkMark) + log.spinner.Suffix + "\n"
	}
	log.mux.Unlock()

	if log.spinner != nil && !log.grouped {
		log.stopSpinner()
	}
}

// Enabled implements logr.LogSink interface
func (log *spinnerLogSink) Enabled(level int) bool {
	return GlobalLogLevel >= level
}

// Error implements logr.LogSink interface
func (log *spinnerLogSink) Error(e error, msg string, keysAndValues ...interface{}) {
	allVal := append(keysAndValues, log.values...)
	if msg != "" {
		msg = color.New(log.colors.Error).Sprintf("%s: %s", msg, log.getDetailedErr(e))
	} else {
		msg = color.New(log.colors.Error).Sprintf("%s", log.getDetailedErr(e))
	}
	if len(allVal) > 0 {
		msg = fmt.Sprintf("%s %s", msg, log.joinAndSeparatePairs(allVal))
	}

	names := log.printNames()
	if names != "" {
		msg = fmt.Sprintf("%s %c %s", names, log.separatorCharacter, msg)
	}

	if log.timeFormat != "" && log.showTime {
		msg = fmt.Sprintf("[%s] %s", time.Now().Format(log.timeFormat), msg)
	}

	if log.spinner == nil {
		log.initSpinner()
	} else {
		log.spinner.Restart()
	}

	log.spinner.Writer = log.err
	log.spinner.Suffix = " " + msg // Append text after the spinner
	log.spinner.FinalMSG = color.New(log.colors.Error).Sprintf("%c", log.errorMark) + log.spinner.Suffix + "\n"

	log.stopSpinner()
}

// WithName implements logr.LogSink interface
func (log *spinnerLogSink) WithName(name string) logr.LogSink {
	l := log.copyLogger()
	l.names = append(l.names, name)

	return l
}

// WithValues implements logr.LogSink interface
func (log *spinnerLogSink) WithValues(keysAndValues ...interface{}) logr.LogSink {
	l := log.copyLogger()
	l.values = append(l.values, keysAndValues)

	return l
}

func (log *spinnerLogSink) SetOptions(options ...Option) {
	for _, opt := range options {
		opt(log)
	}
}

func (log *spinnerLogSink) ShowTime(f bool) *spinnerLogSink {
	l := log.copyLogger()
	l.showTime = f

	return l
}

func (log *spinnerLogSink) Grouped(state bool) {
	if !log.grouppable {
		return
	}

	if state && log.spinner == nil {
		log.initSpinner()
	} else if !state && log.spinner != nil {
		log.stopSpinner()
	}

	log.grouped = state
}

func (log *spinnerLogSink) printNames() string {
	return strings.Join(log.names, "/")
}

func (log *spinnerLogSink) initSpinner() {
	log.mux.Lock()
	defer log.mux.Unlock()
	log.spinner = spinner.New(
		spinner.CharSets[21],
		100*time.Millisecond,
		spinner.WithHiddenCursor(false),
		spinner.WithWriter(log.out),
	)
	_ = log.spinner.Color("green")
	log.spinner.Start()
}

func (log *spinnerLogSink) stopSpinner() {
	log.mux.Lock()
	defer log.mux.Unlock()
	if log.spinner != nil {
		log.spinner.Stop()
		log.spinner = nil
	}
}

func (log *spinnerLogSink) copyLogger() *spinnerLogSink {
	names := make([]string, len(log.names))
	copy(names, log.names)

	values := make([]interface{}, len(log.values))
	copy(values, log.values)

	return &spinnerLogSink{
		names:              log.names,
		values:             log.values,
		out:                log.out,
		err:                log.err,
		grouppable:         log.grouppable,
		truncate:           log.truncate,
		timeFormat:         log.timeFormat,
		showTime:           log.showTime,
		checkMark:          log.checkMark,
		errorMark:          log.errorMark,
		separatorCharacter: log.separatorCharacter,
		colors:             log.colors,

		grouped: log.grouped,
		spinner: log.spinner,

		mux: sync.Mutex{},
	}
}

func (*spinnerLogSink) truncateString(str string, num int) string {
	bnoden := str
	if len(str) > num {
		if num > 3 { //nolint:gomnd
			num -= 3
		}
		bnoden = str[0:num] + "..."
	}
	return bnoden
}

func (log *spinnerLogSink) joinAndSeparatePairs(values []interface{}) string {
	joined := ""
	c := log.colors.Key
	for i, v := range values {
		s, err := cast.ToStringE(v)
		if err != nil {
			s = fmt.Sprintf("%v", v)
		}
		joined += color.New(c).Sprint(s)
		if i%2 == 0 {
			c = 0
			joined += "="
		} else {
			c = log.colors.Key
			if i < len(values)-1 {
				joined += ", "
			}
		}
	}
	return joined
}

func (log *spinnerLogSink) getDetailedErr(err error) string {
	if err == nil {
		return ""
	}

	details := errors.GetDetails(err)
	if len(details) == 0 {
		return err.Error()
	}
	return fmt.Sprintf("%s (%s)", err.Error(), log.joinAndSeparatePairs(details))
}
