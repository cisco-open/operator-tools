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

type Strimap = map[string]interface{}

type StrimapBuilder Strimap

func (s StrimapBuilder) Getin(strs ...string) Strimap {
	if s == nil || len(strs) == 0 {
		return nil
	}

	submap, ok := s[strs[0]].(Strimap)
	if !ok {
		return nil
	}

	if len(strs[1:]) > 0 {
		return StrimapBuilder(submap).Getin(strs[1:]...)
	}
	return submap
}

func MergeMaps(a, b Strimap) Strimap {
	out := make(Strimap, len(a))
	for k, v := range a {
		out[k] = v
	}
	for k, v := range b {
		if v, ok := v.(Strimap); ok {
			if bv, ok := out[k]; ok {
				if bv, ok := bv.(Strimap); ok {
					out[k] = MergeMaps(bv, v)
					continue
				}
			}
		}
		out[k] = v
	}
	return out
}