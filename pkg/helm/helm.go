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
	"github.com/ghodss/yaml"
	"helm.sh/helm/v3/pkg/releaseutil"

	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/chartutil"
	"helm.sh/helm/v3/pkg/engine"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
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

func Render(fs http.FileSystem, values map[string]interface{}, releaseOptions ReleaseOptions, chartName string) ([]runtime.Object, error) {
	files, err := getFiles(fs)
	if err != nil {
		return nil, err
	}

	// Create chart and render templates
	chrt, err := loader.LoadFiles(files)
	if err != nil {
		return nil, err
	}

	renderOpts := chartutil.ReleaseOptions{
		Name:      releaseOptions.Name,
		IsInstall: true,
		IsUpgrade: false,
		Namespace: releaseOptions.Namespace,
	}

	if err := chartutil.ProcessDependencies(chrt, values); err != nil {
		return nil, err
	}
	renderedValues, err := chartutil.ToRenderValues(chrt, values, renderOpts, nil)
	if err != nil {
		return nil, err
	}
	renderedTemplates, err := engine.Render(chrt, renderedValues)
	if err != nil {
		return nil, err
	}

	// Merge templates and inject
	var objects []runtime.Object
	for _, tmpl := range files {
		if !strings.HasSuffix(tmpl.Name, "yaml") && !strings.HasSuffix(tmpl.Name, "yml") && !strings.HasSuffix(tmpl.Name, "tpl") {
			continue
		}
		t := path.Join(chartName, tmpl.Name)
		if renderedTemplate, ok := renderedTemplates[t]; ok {
			renderedTemplate = strings.TrimSpace(renderedTemplate)
			if renderedTemplate == "" {
				continue
			}

			manifests := releaseutil.SplitManifests(renderedTemplate)
			for _, manifest := range manifests {
				yamlDoc := strings.TrimSpace(manifest)
				if yamlDoc == "" {
					continue
				}

				// convert yaml to json
				json, err := yaml.YAMLToJSON([]byte(yamlDoc))
				if err != nil {
					return nil, errors.WrapIfWithDetails(err, "unable to convert yaml to json", map[string]interface{}{"templatePath": t})
				}

				if string(json) == "null" {
					continue
				}

				// deserialize json into unstructured
				o, _, err := unstructured.UnstructuredJSONScheme.Decode(json, nil, nil)
				if err != nil {
					return nil, errors.WrapIfWithDetails(err, "unable to create unstructured", map[string]interface{}{"templatePath": t})
				}
				objects = append(objects, o)
			}
		}
	}

	return objects, nil
}

func getFiles(fs http.FileSystem) ([]*loader.BufferedFile, error) {
	files := []*loader.BufferedFile{
		{
			Name: chartutil.ChartfileName,
		},
		{
			// Without requirements.yaml legacy charts's subdependencies will be processed but cannot be disabled
			// See https://github.com/helm/helm/blob/e2442699fa4703456b16884990c5218c16ed16fc/pkg/chart/loader/load.go#L105
			Name: "requirements.yaml",
		},
	}

	// if the Helm chart templates use some resource files (like dashboards), those should be put under resources
	for _, dirName := range []string{"resources", "crds", chartutil.TemplatesDir, chartutil.ChartsDir} {
		dir, err := fs.Open(dirName)
		if err != nil {
			if !os.IsNotExist(err) {
				return nil, err
			}
		} else {
			// Recursively get the all files from the dir and it's subfolders
			files, err = getFilesFromDir(fs, dir, files, dirName)
			if err != nil {
				return nil, err
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

func getFilesFromDir(fs http.FileSystem, dir http.File, files []*loader.BufferedFile, dirName string) ([]*loader.BufferedFile, error) {
	dirFiles, err := dir.Readdir(-1)
	if err != nil {
		return nil, err
	}

	for _, file := range dirFiles {
		filename := file.Name()
		if strings.HasSuffix(filename, "yaml") || strings.HasSuffix(filename, "yml") || strings.HasSuffix(filename, "tpl") || strings.HasSuffix(filename, "json") {
			files = append(files, &loader.BufferedFile{
				Name: dirName + "/" + filename,
			})
		} else if file.IsDir() {
			dir, err := fs.Open(dirName + "/" + filename)
			if err != nil {
				return nil, err
			}

			files, err = getFilesFromDir(fs, dir, files, dirName+"/"+filename)
			if err != nil {
				return nil, err
			}
		}
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
