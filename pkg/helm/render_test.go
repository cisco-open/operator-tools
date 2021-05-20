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
	"net/http"
	"testing"

	"github.com/ghodss/yaml"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
)

func TestRenderChartWithCrdsAndTemplates(t *testing.T) {
	chart := http.Dir("testdata/crds-and-templates/logging-operator")

	defaultValues, err := GetDefaultValues(chart)
	require.NoError(t, err)

	valuesMap := map[string]interface{}{}
	err = yaml.Unmarshal(defaultValues, &valuesMap)
	require.NoError(t, err)

	// custom resources in templates must be disabled explicitly
	valuesMap["createCustomResource"] = false

	objects, err := Render(chart, valuesMap, ReleaseOptions{
		Name:      "release-name",
		Namespace: "release-namespace",
	}, "logging-operator")
	require.NoError(t, err)

	assert.Len(t, objects, 1)

	o, ok := objects[0].(*unstructured.Unstructured)
	assert.True(t, ok, "object should be unstructured")

	assert.Equal(t, "loggings.logging.banzaicloud.io", o.GetName())
}


func TestRenderChartWithCrdsOnly(t *testing.T) {
	chart := http.Dir("testdata/crds-only/logging-operator")

	defaultValues, err := GetDefaultValues(chart)
	require.NoError(t, err)

	valuesMap := map[string]interface{}{}
	err = yaml.Unmarshal(defaultValues, &valuesMap)
	require.NoError(t, err)

	objects, err := Render(chart, valuesMap, ReleaseOptions{
		Name:      "release-name",
		Namespace: "release-namespace",
	}, "logging-operator")
	require.NoError(t, err)

	assert.Len(t, objects, 1)

	o, ok := objects[0].(*unstructured.Unstructured)
	assert.True(t, ok, "object should be unstructured")

	assert.Equal(t, "loggings.logging.banzaicloud.io", o.GetName())
}

func TestRenderWithScheme(t *testing.T) {
	chart := http.Dir("testdata/templates/logging-operator")

	defaultValues, err := GetDefaultValues(chart)
	require.NoError(t, err)

	valuesMap := map[string]interface{}{}
	err = yaml.Unmarshal(defaultValues, &valuesMap)
	require.NoError(t, err)

	scheme := runtime.NewScheme()
	_ = clientgoscheme.AddToScheme(scheme)

	objects, err := Render(chart, valuesMap, ReleaseOptions{
		Name:      "release-name",
		Namespace: "release-namespace",
		Scheme: scheme,
	}, "logging-operator")
	require.NoError(t, err)

	assert.Len(t, objects, 1)

	_, ok := objects[0].(*v1.ServiceAccount)
	assert.True(t, ok, "object should be a ServiceAccount")
}
