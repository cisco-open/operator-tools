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
	"emperror.dev/errors"
	ypatch "github.com/cppforlife/go-patch/patch"
	"gopkg.in/yaml.v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	k8syaml "sigs.k8s.io/yaml"

	"github.com/cisco-open/operator-tools/pkg/types"
	"github.com/cisco-open/operator-tools/pkg/utils"
)

type GroupVersionKind struct {
	Group   string `json:"group,omitempty"`
	Version string `json:"version,omitempty"`
	Kind    string `json:"kind,omitempty"`
}

// +kubebuilder:object:generate=true
type K8SResourceOverlay struct {
	GVK       *GroupVersionKind         `json:"groupVersionKind,omitempty"`
	ObjectKey types.ObjectKey           `json:"objectKey,omitempty"`
	Patches   []K8SResourceOverlayPatch `json:"patches,omitempty"`
}

// +kubebuilder:object:generate=true
type OverlayPatchType string

const (
	ReplaceOverlayPatchType OverlayPatchType = "replace"
	DeleteOverlayPatchType  OverlayPatchType = "remove"
)

// +kubebuilder:object:generate=true
type K8SResourceOverlayPatch struct {
	Type       OverlayPatchType `json:"type,omitempty"`
	Path       *string          `json:"path,omitempty"`
	Value      *string          `json:"value,omitempty"`
	ParseValue bool             `json:"parseValue,omitempty"`
}

func PatchYAMLModifier(overlay K8SResourceOverlay, parser *ObjectParser) (ObjectModifierFunc, error) {
	if len(overlay.Patches) == 0 {
		return func(o runtime.Object) (runtime.Object, error) {
			return o, nil
		}, nil
	}

	var opsDefinitions []ypatch.OpDefinition
	for _, patch := range overlay.Patches {
		var value interface{}
		if patch.ParseValue {
			err := yaml.Unmarshal([]byte(utils.PointerToString(patch.Value)), &value)
			if err != nil {
				return nil, errors.WrapIf(err, "could not unmarshal value")
			}
		} else {
			value = interface{}(patch.Value)
		}

		op := ypatch.OpDefinition{
			Type: string(patch.Type),
			Path: patch.Path,
		}
		if patch.Type == ReplaceOverlayPatchType {
			op.Value = &value
		}
		opsDefinitions = append(opsDefinitions, op)
	}

	return func(o runtime.Object) (runtime.Object, error) {
		var ok bool
		var meta metav1.Object

		if overlay.GVK != nil {
			gvk := o.GetObjectKind().GroupVersionKind()
			if overlay.GVK.Group != "" && overlay.GVK.Group != gvk.Group {
				return o, nil
			}
			if overlay.GVK.Version != "" && overlay.GVK.Version != gvk.Version {
				return o, nil
			}
			if overlay.GVK.Kind != "" && overlay.GVK.Kind != gvk.Kind {
				return o, nil
			}
		}

		if meta, ok = o.(metav1.Object); !ok {
			return o, nil
		}

		if (overlay.ObjectKey.Name != "" && meta.GetName() != overlay.ObjectKey.Name) || (overlay.ObjectKey.Namespace != "" && meta.GetNamespace() != overlay.ObjectKey.Namespace) {
			return o, nil
		}

		y, err := k8syaml.Marshal(o)
		if err != nil {
			return o, errors.WrapIf(err, "could not marshal runtime object")
		}

		ops, err := ypatch.NewOpsFromDefinitions(opsDefinitions)
		if err != nil {
			return o, errors.WrapIf(err, "could not init patch ops from definitions")
		}

		var in interface{}
		err = yaml.Unmarshal(y, &in)
		if err != nil {
			return o, errors.WrapIf(err, "could not unmarshal resource yaml")
		}

		res, err := ops.Apply(in)
		if err != nil {
			return o, errors.WrapIf(err, "could not apply patch ops")
		}

		y, err = yaml.Marshal(res)
		if err != nil {
			return o, errors.WrapIf(err, "could not marshal patched object to yaml")
		}

		o, err = parser.ParseYAMLToK8sObject(y)
		if err != nil {
			return o, errors.WrapIf(err, "could not parse runtime object from yaml")
		}

		return o, nil
	}, nil
}

func ConvertGVK(gvk schema.GroupVersionKind) GroupVersionKind {
	return GroupVersionKind{
		Group:   gvk.Group,
		Version: gvk.Version,
		Kind:    gvk.Kind,
	}
}
