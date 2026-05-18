// Copyright © 2026 Banzai Cloud
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

package reconciler

import (
	"emperror.dev/errors"
	json "github.com/json-iterator/go"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/cisco-open/k8s-objectmatcher/patch"
)

// FallbackPatchMaker falls back to a three-way JSON merge patch when the
// primary patch.Maker errors — e.g. on types whose custom MarshalJSON
// flattens map entries to top-level JSON keys that reflective strategic-merge
// cannot resolve (opentelemetry-operator's AnyConfig).
type FallbackPatchMaker struct {
	primary          patch.Maker
	annotator        *patch.Annotator
	jsonMergePatcher patch.JSONMergePatcher
}

func NewFallbackPatchMaker(primary patch.Maker, annotator *patch.Annotator, jsonMergePatcher patch.JSONMergePatcher) *FallbackPatchMaker {
	return &FallbackPatchMaker{primary: primary, annotator: annotator, jsonMergePatcher: jsonMergePatcher}
}

func (m *FallbackPatchMaker) Calculate(current, modified runtime.Object, opts ...patch.CalculateOption) (*patch.PatchResult, error) {
	if result, err := m.primary.Calculate(current, modified, opts...); err == nil {
		return result, nil
	}
	return m.calculateJSONMerge(current, modified, opts...)
}

func (m *FallbackPatchMaker) calculateJSONMerge(currentObj, modifiedObj runtime.Object, opts ...patch.CalculateOption) (*patch.PatchResult, error) {
	current, err := json.ConfigCompatibleWithStandardLibrary.Marshal(currentObj)
	if err != nil {
		return nil, errors.Wrap(err, "marshal current")
	}
	modified, err := json.ConfigCompatibleWithStandardLibrary.Marshal(modifiedObj)
	if err != nil {
		return nil, errors.Wrap(err, "marshal modified")
	}

	for _, opt := range opts {
		if current, modified, err = opt(current, modified); err != nil {
			return nil, errors.Wrap(err, "apply CalculateOption")
		}
	}

	if current, _, err = patch.DeleteNullInJson(current); err != nil {
		return nil, errors.Wrap(err, "strip nulls from current")
	}
	if modified, _, err = patch.DeleteNullInJson(modified); err != nil {
		return nil, errors.Wrap(err, "strip nulls from modified")
	}

	original, err := m.annotator.GetOriginalConfiguration(currentObj)
	if err != nil {
		return nil, errors.Wrap(err, "read last-applied annotation")
	}

	patchBytes, err := m.jsonMergePatcher.CreateThreeWayJSONMergePatch(original, modified, current)
	if err != nil {
		return nil, errors.Wrap(err, "json merge fallback")
	}

	return &patch.PatchResult{Patch: patchBytes, Current: current, Modified: modified, Original: original}, nil
}
