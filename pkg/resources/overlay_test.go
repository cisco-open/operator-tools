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
	"testing"

	"github.com/go-test/deep"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"

	"github.com/cisco-open/operator-tools/pkg/types"
	"github.com/cisco-open/operator-tools/pkg/utils"
)

func TestPatchYAMLModifier(t *testing.T) {
	objectName := "test-object"
	testNamespace := "test-ns"

	baseObject := &v1.Service{
		TypeMeta: v12.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: v12.ObjectMeta{
			Name:      objectName,
			Namespace: testNamespace,
		},
		Spec: v1.ServiceSpec{
			LoadBalancerIP: "1.2.3.4",
			Ports: []v1.ServicePort{
				{
					Port: 123,
					Name: "port1",
				},
				{
					Port: 456,
					Name: "port-to-delete",
				},
			},
		},
	}

	baseWantObject := &v1.Service{
		TypeMeta: v12.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: v12.ObjectMeta{
			Name:      objectName,
			Namespace: testNamespace,
		},
		Spec: v1.ServiceSpec{
			LoadBalancerIP: "5.6.7.8",
			Ports: []v1.ServicePort{
				{
					Port: 123,
					Name: "port2",
				},
			},
		},
	}

	parser := NewObjectParser(clientgoscheme.Scheme)
	tests := map[string]struct {
		overlay           K8SResourceOverlay
		object            runtime.Object
		want              runtime.Object
		assertErr         func(error)
		assertModifierErr func(error)
	}{
		"matching patching": {
			overlay: K8SResourceOverlay{
				ObjectKey: types.ObjectKey{
					Name:      objectName,
					Namespace: testNamespace,
				},
				Patches: []K8SResourceOverlayPatch{
					{
						Type:  ReplaceOverlayPatchType,
						Path:  utils.StringPointer("/spec/loadBalancerIP"),
						Value: utils.StringPointer("5.6.7.8"),
					},
					{
						Type:  ReplaceOverlayPatchType,
						Path:  utils.StringPointer("/spec/ports/0/name"),
						Value: utils.StringPointer("port2"),
					},
					{
						Type: DeleteOverlayPatchType,
						Path: utils.StringPointer("/spec/ports/1"),
					},
				},
			},
			object: baseObject,
			want:   baseWantObject,
		},
		"matching wo objectkey patching": {
			overlay: K8SResourceOverlay{
				Patches: []K8SResourceOverlayPatch{
					{
						Type:  ReplaceOverlayPatchType,
						Path:  utils.StringPointer("/spec/loadBalancerIP"),
						Value: utils.StringPointer("5.6.7.8"),
					},
					{
						Type:  ReplaceOverlayPatchType,
						Path:  utils.StringPointer("/spec/ports/0/name"),
						Value: utils.StringPointer("port2"),
					},
					{
						Type: DeleteOverlayPatchType,
						Path: utils.StringPointer("/spec/ports/1"),
					},
				},
			},
			object: baseObject,
			want:   baseWantObject,
		},
		"matching wo objectkey namespace patching": {
			overlay: K8SResourceOverlay{
				ObjectKey: types.ObjectKey{
					Name: objectName,
				},
				Patches: []K8SResourceOverlayPatch{
					{
						Type:  ReplaceOverlayPatchType,
						Path:  utils.StringPointer("/spec/loadBalancerIP"),
						Value: utils.StringPointer("5.6.7.8"),
					},
					{
						Type:  ReplaceOverlayPatchType,
						Path:  utils.StringPointer("/spec/ports/0/name"),
						Value: utils.StringPointer("port2"),
					},
					{
						Type: DeleteOverlayPatchType,
						Path: utils.StringPointer("/spec/ports/1"),
					},
				},
			},
			object: baseObject,
			want:   baseWantObject,
		},
		"matching wo objectkey name patching": {
			overlay: K8SResourceOverlay{
				ObjectKey: types.ObjectKey{
					Namespace: testNamespace,
				},
				Patches: []K8SResourceOverlayPatch{
					{
						Type:  ReplaceOverlayPatchType,
						Path:  utils.StringPointer("/spec/loadBalancerIP"),
						Value: utils.StringPointer("5.6.7.8"),
					},
					{
						Type:  ReplaceOverlayPatchType,
						Path:  utils.StringPointer("/spec/ports/0/name"),
						Value: utils.StringPointer("port2"),
					},
					{
						Type: DeleteOverlayPatchType,
						Path: utils.StringPointer("/spec/ports/1"),
					},
				},
			},
			object: baseObject,
			want:   baseWantObject,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := PatchYAMLModifier(tt.overlay, parser)
			if tt.assertErr != nil {
				tt.assertErr(err)
			} else {
				assert.NoError(t, err)
			}
			patched, err := got(tt.object)
			if tt.assertModifierErr != nil {
				tt.assertModifierErr(err)
			} else {
				assert.NoError(t, err)
			}
			if diff := deep.Equal(patched, tt.want); diff != nil {
				t.Error(diff)
			}
		})
	}
}
