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

package docgen

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"emperror.dev/errors"
	"github.com/go-logr/logr"
	"github.com/iancoleman/orderedmap"
)

type SourceLister struct {
	Logger                       logr.Logger
	Sources                      map[string]SourceDir
	IgnoredSources               []string
	IncludeSources               []string
	DefaultValueFromTagExtractor func(string) string
	Index                        *Doc
	DocGeneratedHook             func(doc *Doc) error
	Header                       string
	Footer                       string
}

type DocIndex struct {
	Path string
}

type SourceDir struct {
	Path     string
	DestPath string
}

func NewSourceLister(sources map[string]SourceDir, logger logr.Logger) *SourceLister {
	return &SourceLister{
		Logger:  logger,
		Sources: sources,
	}
}

func (sl *SourceLister) ListSources() ([]DocItem, error) {
	sourceList := []DocItem{}

	for _, key := range orderedMap(sl.Sources).Keys() {
		p := sl.Sources[key]
		files, err := os.ReadDir(p.Path)
		if err != nil {
			return nil, errors.WrapIff(err, "failed to read files from %s", p.Path)
		}
		for _, file := range files {
			fname := strings.Replace(file.Name(), ".go", "", 1)
			if filepath.Ext(file.Name()) == ".go" && (sl.isWhitelisted(fname) || !sl.isBlacklisted(fname)) {
				fullPath := filepath.Join(p.Path, file.Name())
				sl.Logger.V(2).Info("included", "source", fname)
				sourceList = append(sourceList, DocItem{
					Name:                         fname,
					SourcePath:                   fullPath,
					DestPath:                     p.DestPath,
					DefaultValueFromTagExtractor: sl.DefaultValueFromTagExtractor,
					Category:                     key,
				},
				)
			}
		}
	}

	return sourceList, nil
}

func (sl *SourceLister) isBlacklisted(source string) bool {
	for _, p := range sl.IgnoredSources {
		r := regexp.MustCompile(p)
		if r.MatchString(source) {
			sl.Logger.V(2).Info("blacklisted", "source", source)
			return true
		}
	}
	return false
}

func (sl *SourceLister) isWhitelisted(source string) bool {
	for _, p := range sl.IncludeSources {
		r := regexp.MustCompile(p)
		if r.MatchString(source) {
			sl.Logger.V(2).Info("whitelisted", "source", source)
			return true
		}
	}
	return false
}

func (lister *SourceLister) Generate() error {
	lister.Index.Append(lister.Header)

	sources, err := lister.ListSources()
	if err != nil {
		return errors.WrapIf(err, "failed to get plugin list")
	}

	for _, source := range sources {
		document := GetDocumentParser(source, lister.Logger.WithName("docgen"))
		if err := document.Generate(); err != nil {
			return err
		}

		if lister.DocGeneratedHook != nil {
			if err := lister.DocGeneratedHook(document); err != nil {
				return err
			}
		}
	}

	lister.Index.Append(lister.Footer)

	if err := lister.Index.Generate(); err != nil {
		return err
	}

	return nil
}

func orderedMap(original map[string]SourceDir) *orderedmap.OrderedMap {
	o := orderedmap.New()
	for k, v := range original {
		o.Set(k, v)
	}
	o.SortKeys(sort.Strings)
	return o
}
