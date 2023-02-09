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

package reconciler

import (
	"encoding/json"

	"emperror.dev/errors"

	"github.com/cisco-open/k8s-objectmatcher/patch"
)

func IgnoreManagedFields() patch.CalculateOption {
	return func(current, modified []byte) ([]byte, []byte, error) {
		current, err := deleteManagedFields(current)
		if err != nil {
			return []byte{}, []byte{}, errors.WrapIf(err, "could not delete managed fields from modified byte sequence")
		}

		modified, err = deleteManagedFields(modified)
		if err != nil {
			return []byte{}, []byte{}, errors.WrapIf(err, "could not delete managed fields from modified byte sequence")
		}

		return current, modified, nil
	}
}

func deleteManagedFields(obj []byte) ([]byte, error) {
	var objectMap map[string]interface{}
	err := json.Unmarshal(obj, &objectMap)
	if err != nil {
		return []byte{}, errors.WrapIf(err, "could not unmarshal byte sequence")
	}
	if metadata, ok := objectMap["metadata"].(map[string]interface{}); ok {
		delete(metadata, "managedFields")
		objectMap["metadata"] = metadata
	}
	obj, err = json.Marshal(objectMap)
	if err != nil {
		return []byte{}, errors.WrapIf(err, "could not marshal byte sequence")
	}

	return obj, nil
}
