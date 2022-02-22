// Copyright © 2022 Banzai Cloud
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
	"github.com/go-logr/logr"
)

type LogrAdapter struct {
	logr.Logger
}

func NewWithLogrLogger(logger logr.Logger) Logger {
	return &LogrAdapter{
		Logger: logger,
	}
}

func (log *LogrAdapter) GetLogrLogger() logr.Logger {
	return log.Logger
}

func (log *LogrAdapter) V(level int) Logger {
	return &LogrAdapter{
		Logger: log.Logger.V(level),
	}
}

func (log *LogrAdapter) WithValues(keysAndValues ...interface{}) Logger {
	return &LogrAdapter{
		Logger: log.Logger.WithValues(keysAndValues...),
	}
}

func (log *LogrAdapter) WithName(name string) Logger {
	return &LogrAdapter{
		Logger: log.Logger.WithName(name),
	}
}
