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
	"bytes"
	"net/http"
	"os"
	"path"
	"strings"

	"emperror.dev/errors"
	"k8s.io/helm/pkg/chartutil"
	"k8s.io/helm/pkg/proto/hapi/chart"
	"k8s.io/helm/pkg/renderutil"
	"k8s.io/helm/pkg/timeconv"
)

type ReleaseOptions chartutil.ReleaseOptions

func GetDefaultValues(fs http.FileSystem) ([]byte, error) {
	file, err := fs.Open(chartutil.ValuesfileName)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(file)
	if err != nil {
		return nil, errors.WrapIf(err, "could not read default values")
	}

	return buf.Bytes(), nil
}

func Render(fs http.FileSystem, values string, releaseOptions ReleaseOptions, chartName string) (string, error) {
	chrtConfig := &chart.Config{
		Raw:    values,
		Values: map[string]*chart.Value{},
	}

	files, err := getFiles(fs)
	if err != nil {
		return "", err
	}

	// Create chart and render templates
	chrt, err := chartutil.LoadFiles(files)
	if err != nil {
		return "", err
	}

	renderOpts := renderutil.Options{
		ReleaseOptions: chartutil.ReleaseOptions{
			Name:      releaseOptions.Name,
			IsInstall: true,
			IsUpgrade: false,
			Time:      timeconv.Now(),
			Namespace: releaseOptions.Namespace,
		},
		KubeVersion: "",
	}

	renderedTemplates, err := renderutil.Render(chrt, chrtConfig, renderOpts)
	if err != nil {
		return "", err
	}

	// Merge templates and inject
	var buf bytes.Buffer
	for _, tmpl := range files {
		if !strings.HasSuffix(tmpl.Name, "yaml") && !strings.HasSuffix(tmpl.Name, "yml") && !strings.HasSuffix(tmpl.Name, "tpl") {
			continue
		}
		t := path.Join(chartName, tmpl.Name)
		if _, err := buf.WriteString(renderedTemplates[t]); err != nil {
			return "", err
		}
		buf.WriteString("\n---\n")
	}

	return buf.String(), nil
}

func getFiles(fs http.FileSystem) ([]*chartutil.BufferedFile, error) {
	files := []*chartutil.BufferedFile{
		{
			Name: chartutil.ChartfileName,
		},
	}

	// if the Helm chart templates use some resource files (like dashboards), those should be put under resources
	for _, dirName := range []string{"resources", chartutil.TemplatesDir} {
		dir, err := fs.Open(dirName)
		if err != nil {
			if !os.IsNotExist(err) {
				return nil, err
			}
		} else {
			dirFiles, err := dir.Readdir(-1)
			if err != nil {
				return nil, err
			}

			for _, file := range dirFiles {
				filename := file.Name()
				if strings.HasSuffix(filename, "yaml") || strings.HasSuffix(filename, "yml") || strings.HasSuffix(filename, "tpl") || strings.HasSuffix(filename, "json") {
					files = append(files, &chartutil.BufferedFile{
						Name: dirName + "/" + filename,
					})
				}
			}
		}
	}

	for _, f := range files {
		data, err := readIntoBytes(fs, f.Name)
		if err != nil {
			return nil, err
		}

		f.Data = data
	}

	return files, nil
}

func readIntoBytes(fs http.FileSystem, filename string) ([]byte, error) {
	file, err := fs.Open(filename)
	if err != nil {
		return nil, errors.WrapIf(err, "could not open file")
	}
	defer file.Close()

	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(file)
	if err != nil {
		return nil, errors.WrapIf(err, "could not read file")
	}

	return buf.Bytes(), nil
}
