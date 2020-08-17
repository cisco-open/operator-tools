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

package helm

import (
	"reflect"
	"strings"
	"testing"

	"k8s.io/utils/diff"
)

func Test_splitMultiYamlDoc(t *testing.T) {
	multiYamlDoc := `
a: b
---
c: d
---
e: f ---- hh
  ---
e: f ----xx
`
	expected := []string{
		"a: b",
		"c: d",
		"e: f ---- hh",
		"e: f ----xx",
	}

	docs := splitMultiYamlDoc(strings.TrimSpace(multiYamlDoc))

	if !reflect.DeepEqual(expected, docs) {
		t.Error(diff.ObjectDiff(expected, docs))
	}
}
