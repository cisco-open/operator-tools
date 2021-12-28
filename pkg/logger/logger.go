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

type GroupedLogger interface {
	Grouped(state bool)
}

type Colors struct {
	Info  color.Attribute
	Error color.Attribute
	Key   color.Attribute
}

type logger struct {
	level  int
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

var GlobalLogLevel = 1

var Log logr.Logger = New()

type Option func(*logger)

func Out(w io.Writer) Option {
	return func(l *logger) {
		l.out = w
	}
}

func Err(w io.Writer) Option {
	return func(l *logger) {
		l.err = w
	}
}

func Grouppable() Option {
	return func(l *logger) {
		l.grouppable = true
	}
}

func Truncate() Option {
	return func(l *logger) {
		l.truncate = true
	}
}

func Color(colors Colors) Option {
	return func(l *logger) {
		l.colors = colors
	}
}

func CheckMarkCharacter(m rune) Option {
	return func(l *logger) {
		l.checkMark = m
	}
}

func ErrorMarkCharacter(m rune) Option {
	return func(l *logger) {
		l.errorMark = m
	}
}

func SeparatorCharacter(m rune) Option {
	return func(l *logger) {
		l.separatorCharacter = m
	}
}

func WithName(name string) Option {
	return func(l *logger) {
		l.names = append(l.names, name)
	}
}

func WithTime(format string) Option {
	return func(l *logger) {
		l.timeFormat = format
		l.showTime = true
	}
}

func New(options ...Option) logr.Logger {
	l := &logger{
		names: []string{},
		level: 0,
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

	return logr.New(l)
}

func (log *logger) SetOptions(options ...Option) {
	for _, opt := range options {
		opt(log)
	}
}

// Info implements logr.InfoLogger.
func (log *logger) Info(level int, msg string, vals ...interface{}) {
	if GlobalLogLevel >= log.level {
		allVal := append(vals, log.values...)
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
}

func (log *logger) Init(logr.RuntimeInfo) {}

// Enabled implements logr.InfoLogger.
func (log *logger) Enabled(level int) bool {
	return true
}

// Error implements logr.logger.
func (log *logger) Error(e error, msg string, vals ...interface{}) {
	allVal := append(vals, log.values...)
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

func (log *logger) printNames() string {
	return strings.Join(log.names, "/")
}

// V implements logr.logger
func (log *logger) V(level int) logr.LogSink {
	l := log.copyLogger()
	l.level = level

	return &l
}

// WithName implements logr.logger
func (log *logger) WithName(name string) logr.LogSink {
	l := log.copyLogger()
	l.names = append(l.names, name)

	return &l
}

// WithValues implements logr.logger
func (log *logger) WithValues(values ...interface{}) logr.LogSink {
	l := log.copyLogger()
	l.values = values

	return &l
}

func (log *logger) ShowTime(f bool) logr.LogSink {
	l := log.copyLogger()
	l.showTime = f

	return &l
}

func (log *logger) Grouped(state bool) {
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

func (log *logger) initSpinner() {
	log.mux.Lock()
	defer log.mux.Unlock()
	log.spinner = spinner.New(spinner.CharSets[21], 100*time.Millisecond, spinner.WithHiddenCursor(false), spinner.WithWriter(log.out))
	_ = log.spinner.Color("green")
	log.spinner.Start()
}

func (log *logger) stopSpinner() {
	log.mux.Lock()
	defer log.mux.Unlock()
	if log.spinner != nil {
		log.spinner.Stop()
		log.spinner = nil
	}
}

func (log *logger) copyLogger() logger {
	names := make([]string, len(log.names))
	copy(names, log.names)

	values := make([]interface{}, len(log.values))
	copy(values, log.values)

	return logger{
		level:              log.level,
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

func (*logger) truncateString(str string, num int) string {
	bnoden := str
	if len(str) > num {
		if num > 3 {
			num -= 3
		}
		bnoden = str[0:num] + "..."
	}
	return bnoden
}

func (log *logger) joinAndSeparatePairs(vals []interface{}) string {
	joined := ""
	c := log.colors.Key
	for i, v := range vals {
		s, err := cast.ToStringE(v)
		if err != nil {
			s = fmt.Sprintf("%v", v)
		}
		joined += color.New(c).Sprint(s)
		if i%2 == 0 {
			c = 0
			joined = joined + "="
		} else {
			c = log.colors.Key
			if i < len(vals)-1 {
				joined = joined + ", "
			}
		}
	}
	return joined
}

func (log *logger) getDetailedErr(err error) string {
	if err == nil {
		return ""
	}

	details := errors.GetDetails(err)
	if len(details) == 0 {
		return err.Error()
	}
	return fmt.Sprintf("%s (%s)", err.Error(), log.joinAndSeparatePairs(details))
}

func EnableGroupSession(logger interface{}) func() {
	if l, ok := logger.(interface{ Grouped(state bool) }); ok {
		l.Grouped(true)
		return func() { l.Grouped(false) }
	}
	return func() {}
}
