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

package docgen_test

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/MakeNowJust/heredoc"
	"github.com/andreyvit/diff"
	"github.com/go-logr/logr"

	"github.com/cisco-open/operator-tools/pkg/docgen"
	"github.com/cisco-open/operator-tools/pkg/utils"
)

var logger logr.Logger

func init() {
	logger = utils.Log
}

func TestGenParse(t *testing.T) {
	_, filename, _, _ := runtime.Caller(0)
	currentDir := filepath.Dir(filename)

	var testData = []struct {
		docItem  docgen.DocItem
		expected string
	}{
		{
			docItem: docgen.DocItem{
				Name:       "sample",
				SourcePath: filepath.Join(currentDir, "testdata", "sample.go"),
				DestPath:   filepath.Join(currentDir, "../../build/_test/docgen"),
			},
			expected: heredoc.Doc(`
					## Sample

					### field1 (string, optional) {#sample-field1}
					
					Default: -
			`),
		},
		{
			docItem: docgen.DocItem{
				Name:       "sample-default",
				SourcePath: filepath.Join(currentDir, "testdata", "sample_default.go"),
				DestPath:   filepath.Join(currentDir, "../../build/_test/docgen"),
				DefaultValueFromTagExtractor: func(tag string) string {
					return docgen.GetPrefixedValue(tag, `asd:\"default:(.*)\"`)
				},
			},
			expected: heredoc.Doc(`
				## SampleDefault

				### field1 (string, optional) {#sampledefault-field1}
				
				Default: testval
			`),
		},
		{
			docItem: docgen.DocItem{
				Name:       "sample-default-comment",
				SourcePath: filepath.Join(currentDir, "testdata", "sample_default_comment.go"),
				DestPath:   filepath.Join(currentDir, "../../build/_test/docgen"),
			},
			expected: heredoc.Doc(`
				## SampleDefaultComment

				### field1 (string, optional) {#sampledefaultcomment-field1}

				Field1 is a good field.

				Default: testval
			`),
		},
		{
			docItem: docgen.DocItem{
				Name:       "sample-codeblock",
				SourcePath: filepath.Join(currentDir, "testdata", "sample_codeblock.go"),
				DestPath:   filepath.Join(currentDir, "../../build/_test/docgen"),
				DefaultValueFromTagExtractor: func(tag string) string {
					return docgen.GetPrefixedValue(tag, `asd:\"default:(.*)\"`)
				},
			},
			expected: heredoc.Doc(`
				## Sample

				### field1 (string, optional) {#sample-field1}

				Description
				{{< highlight yaml >}}
				test: code block
				some: more lines
					indented: line
				{{< /highlight >}}


				Default: -
			`),
		},
	}

	for _, item := range testData {
		parser := docgen.GetDocumentParser(item.docItem, logger)
		err := parser.Generate()
		if err != nil {
			t.Fatalf("%+v", err)
		}

		bytes, err := os.ReadFile(filepath.Join(item.docItem.DestPath, item.docItem.Name+".md"))
		if err != nil {
			t.Fatalf("%+v", err)
		}

		if a, e := diff.TrimLinesInString(string(bytes)), diff.TrimLinesInString(item.expected); a != e {
			t.Errorf("Result does not match (-actual vs +expected):\n%v\nActual: %s", diff.LineDiff(a, e), string(bytes))
		}
	}
}
