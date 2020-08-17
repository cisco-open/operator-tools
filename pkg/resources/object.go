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

package resources

import (
	"bufio"
	"bytes"
	"strings"

	"emperror.dev/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer/json"
	k8syaml "k8s.io/apimachinery/pkg/util/yaml"

	"github.com/banzaicloud/operator-tools/pkg/logger"
)

var log = logger.Log

const (
	// YAMLSeparator is a separator for multi-document YAML files.
	YAMLSeparator = "---"
)

type Objects []runtime.Object

// ToMap returns a map of K8sObject hash to K8sObject.
func (os Objects) ToMap() map[string]runtime.Object {
	ret := make(map[string]runtime.Object)
	for _, oo := range os {
		if isValid(oo) {
			ret[GetHash(oo)] = oo
		}
	}
	return ret
}

func GetHash(o runtime.Object) string {
	var name, namespace string
	if m, ok := o.(interface {
		GetName() string
		GetNamespace() string
	}); ok {
		name = m.GetName()
		namespace = m.GetNamespace()
	}

	return strings.Join([]string{o.GetObjectKind().GroupVersionKind().String(), namespace, name}, ":")
}

// Valid checks returns true if Kind and ComponentName of K8sObject are both not empty.
func isValid(o runtime.Object) bool {
	if m, ok := o.(metav1.Object); ok {
		if o.GetObjectKind().GroupVersionKind().Kind == "" || m.GetName() == "" {
			return false
		}

		return true
	}

	return false
}

type ObjectParser struct {
	scheme *runtime.Scheme
}

func NewObjectParser(scheme *runtime.Scheme) *ObjectParser {
	return &ObjectParser{
		scheme: scheme,
	}
}

func (p *ObjectParser) ParseYAMLManifest(manifest string, modifiers ...YAMLModifierFuncs) ([]runtime.Object, error) {
	var b bytes.Buffer

	var yamls []string
	scanner := bufio.NewScanner(strings.NewReader(manifest))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == YAMLSeparator {
			yamls = append(yamls, b.String())
			b.Reset()
		} else {
			if _, err := b.WriteString(line); err != nil {
				return nil, err
			}
			if _, err := b.WriteString("\n"); err != nil {
				return nil, err
			}
		}
	}
	yamls = append(yamls, b.String())

	var objects []runtime.Object

	for _, yaml := range yamls {
		yaml = p.removeNonYAMLLines(yaml)
		if yaml == "" {
			continue
		}
		o, err := p.ParseYAMLToK8sObject([]byte(yaml), modifiers...)
		if err != nil {
			log.Error(err, "failed to parse YAML to a k8s object")
			continue
		}

		objects = append(objects, o)
	}

	return objects, nil
}

func (p *ObjectParser) ParseYAMLToK8sObject(yaml []byte, yamlModifiers ...YAMLModifierFuncs) (runtime.Object, error) {
	for _, modifierFunc := range yamlModifiers {
		yaml = modifierFunc(yaml)
	}

	s := json.NewYAMLSerializer(json.DefaultMetaFactory, p.scheme, p.scheme)
	o, _, err := s.Decode(yaml, nil, nil)
	if err != nil {
		r := bytes.NewReader(yaml)
		decoder := k8syaml.NewYAMLOrJSONDecoder(r, 1024)

		out := &unstructured.Unstructured{}
		err := decoder.Decode(out)
		if err != nil {
			return nil, errors.WrapIf(err, "error decoding object")
		}
		o = out
	}

	return o, nil
}

func (p *ObjectParser) removeNonYAMLLines(yms string) string {
	out := ""
	for _, s := range strings.Split(yms, "\n") {
		if strings.HasPrefix(s, "#") {
			continue
		}
		out += s + "\n"
	}

	return strings.TrimSpace(out)
}
