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

package utils

import (
	"fmt"
	"io"
	"os"

	"emperror.dev/errors"
	"github.com/go-logr/logr"
	"github.com/spf13/cast"
)

type logger struct {
	level  int
	name   string
	values []interface{}
	out    io.Writer
	err    io.Writer
}

var GlobalLogLevel = 0

var Log logr.Logger = logr.New(logger{
	level: 0,
	out:   os.Stderr,
	err:   os.Stderr,
})

func NewLogger(name string, out, err io.Writer, level int) logr.Logger {
	return logr.New(&logger{name: name, err: err, out: out, level: level})
}

// Info implements logr.InfoLogger
func (log logger) Info(level int, msg string, vals ...interface{}) {
	if GlobalLogLevel >= log.level {
		allVal := append(vals, log.values...)
		if len(allVal) == 0 {
			fmt.Fprintf(log.out, "%s> %s\n", log.name, msg)
		} else {
			fmt.Fprintf(log.out, "%s> %s %s\n", log.name, msg, joinAndSeparatePairs(allVal))
		}
	}
}

func (logger) Init(logr.RuntimeInfo) {}

// Enabled implements logr.InfoLogger
func (logger) Enabled(level int) bool {
	return true
}

// Error implements logr.logger
func (log logger) Error(e error, msg string, vals ...interface{}) {
	allVal := append(vals, log.values...)
	if len(allVal) == 0 {
		fmt.Fprintf(log.err, "%s> %s %s\n", log.name, msg, getDetailedErr(e))
	} else {
		fmt.Fprintf(log.err, "%s> %s %s %s\n", log.name, msg, getDetailedErr(e), joinAndSeparatePairs(allVal))
	}
}

// V implements logr.logger
func (log logger) V(level int) logr.LogSink {
	return logger{
		name:   log.name,
		level:  level,
		values: log.values,
		out:    log.out,
		err:    log.err,
	}
}

// WithName implements logr.logger
func (log logger) WithName(name string) logr.LogSink {
	return logger{
		name:   name,
		level:  log.level,
		values: log.values,
		out:    log.out,
		err:    log.err,
	}
}

// WithValues implements logr.logger
func (log logger) WithValues(values ...interface{}) logr.LogSink {
	return logger{
		name:   log.name,
		level:  log.level,
		values: values,
		out:    log.out,
		err:    log.err,
	}
}

func joinAndSeparatePairs(vals []interface{}) string {
	joined := ""
	for i, v := range vals {
		joined += cast.ToString(v)
		if i%2 == 0 {
			joined = joined + ": "
		} else {
			if i < len(vals)-1 {
				joined = joined + ", "
			}
		}
	}
	return joined
}

func getDetailedErr(err error) string {
	details := errors.GetDetails(err)
	if len(details) == 0 {
		return err.Error()
	}
	return fmt.Sprintf("%s (%s)", err.Error(), joinAndSeparatePairs(details))
}
