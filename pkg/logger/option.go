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
	"io"
)

type Option func(*spinnerLogSink)

func Out(w io.Writer) Option {
	return func(l *spinnerLogSink) {
		l.out = w
	}
}

func Err(w io.Writer) Option {
	return func(l *spinnerLogSink) {
		l.err = w
	}
}

func Grouppable() Option {
	return func(l *spinnerLogSink) {
		l.grouppable = true
	}
}

func Truncate() Option {
	return func(l *spinnerLogSink) {
		l.truncate = true
	}
}

func Color(colors Colors) Option {
	return func(l *spinnerLogSink) {
		l.colors = colors
	}
}

func CheckMarkCharacter(m rune) Option {
	return func(l *spinnerLogSink) {
		l.checkMark = m
	}
}

func ErrorMarkCharacter(m rune) Option {
	return func(l *spinnerLogSink) {
		l.errorMark = m
	}
}

func SeparatorCharacter(m rune) Option {
	return func(l *spinnerLogSink) {
		l.separatorCharacter = m
	}
}

func WithName(name string) Option {
	return func(l *spinnerLogSink) {
		l.names = append(l.names, name)
	}
}

func WithTime(format string) Option {
	return func(l *spinnerLogSink) {
		l.timeFormat = format
		l.showTime = true
	}
}
