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
	"bytes"
	"testing"

	"github.com/go-logr/logr"
	"github.com/stretchr/testify/require"
)

func TestInfoLog(t *testing.T) {
	type foo string
	var buf bytes.Buffer

	sink := NewSpinnerLogSink(Out(&buf))
	sink.ShowTime(false)
	sink.Info(0, "test", "foo", foo("bar"))
	expected := []byte("\r\x1b[K✓ test foo=bar\n")
	require.Equal(t, expected, buf.Bytes())
}

func TestSpinnerLogSinkWithLogrLogger(t *testing.T) {
	type foo string
	var buf bytes.Buffer

	sink := NewSpinnerLogSink(Out(&buf))
	sink.ShowTime(false)

	log := logr.New(sink)
	log.Info("test", "foo", foo("bar"))
	expected := []byte("\r\x1b[K✓ test foo=bar\n")
	require.Equal(t, expected, buf.Bytes())
}
