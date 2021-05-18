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
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	runtimeClient "sigs.k8s.io/controller-runtime/pkg/client"
)

type BeforeCreateFn func(runtime.Object) error
type BeforeDeleteFn func(runtime.Object) error
type BeforeUpdateFn func(runtime.Object, runtime.Object) error

type DynamicDesiredState struct {
	StaticDesiredState DesiredState
	BeforeCreateFunc   func(desired runtime.Object) error
	BeforeUpdateFunc   []func(current, desired runtime.Object) error
	BeforeDeleteFunc   func(current runtime.Object) error
	CreateOptions      []runtimeClient.CreateOption
	UpdateOptions      []runtimeClient.UpdateOption
	DeleteOptions      []runtimeClient.DeleteOption
	ShouldCreateFunc   func(desired runtime.Object) (bool, error)
	ShouldUpdateFunc   func(current, desired runtime.Object) (bool, error)
	ShouldDeleteFunc   func(desired runtime.Object) (bool, error)
}

func NewDynamicDesiredState() *DynamicDesiredState {
	return &DynamicDesiredState{}
}

func (s *DynamicDesiredState) WithBeforeUpdateFns(fns ...BeforeUpdateFn) *DynamicDesiredState {
	if s == nil {
		return nil
	}
	for _, fn := range fns {
		s.BeforeUpdateFunc = append(s.BeforeUpdateFunc, fn)
	}
	return s
}

func (s *DynamicDesiredState) WithDesiredState(d DesiredState) *DynamicDesiredState {
	s.StaticDesiredState = d
	return s
}

func (s DynamicDesiredState) DesiredState() DesiredState {
	return s.StaticDesiredState
}

func (s DynamicDesiredState) GetDesiredState() DesiredState {
	return s.StaticDesiredState
}

func (s DynamicDesiredState) ShouldCreate(desired runtime.Object) (bool, error) {
	if s.ShouldCreateFunc != nil {
		return s.ShouldCreateFunc(desired)
	}

	return true, nil
}

func (s DynamicDesiredState) ShouldUpdate(current, desired runtime.Object) (bool, error) {
	if s.ShouldUpdateFunc != nil {
		return s.ShouldUpdateFunc(current, desired)
	}

	return true, nil
}

func (s DynamicDesiredState) ShouldDelete(desired runtime.Object) (bool, error) {
	if s.ShouldDeleteFunc != nil {
		return s.ShouldDeleteFunc(desired)
	}

	return true, nil
}

func (s DynamicDesiredState) BeforeCreate(desired runtime.Object) error {
	if s.BeforeCreateFunc != nil {
		return s.BeforeCreateFunc(desired)
	}

	return nil
}

func (s DynamicDesiredState) BeforeUpdate(current, desired runtime.Object) error {
	for _, fn := range s.BeforeUpdateFunc {
		if err := fn(current, desired); err != nil {
			return err
		}
	}

	return nil
}

func (s DynamicDesiredState) BeforeDelete(current runtime.Object) error {
	if s.BeforeDeleteFunc != nil {
		return s.BeforeDeleteFunc(current)
	}

	return nil
}

func (s DynamicDesiredState) GetCreateptions() []runtimeClient.CreateOption {
	return s.CreateOptions
}

func (s DynamicDesiredState) GetUpdateOptions() []runtimeClient.UpdateOption {
	return s.UpdateOptions
}

func (s DynamicDesiredState) GetDeleteOptions() []runtimeClient.DeleteOption {
	return s.DeleteOptions
}

// TODO place into another file

// BeforeUpdateFn
// func DesiredStateUpdateServiceIPModifier(current, desired runtime.Object) error {
// 	if co, ok := current.(*corev1.Service); ok {
// 		do := desired.(*corev1.Service)
// 		do.Spec.ClusterIP = co.Spec.ClusterIP
// 	}

// 	return nil
// }
func DesiredStateUpdateServiceIPModifier(current, desired runtime.Object) error {
	if co, ok := current.(*corev1.Service); ok {
		switch dt := desired.(type) {
		case *unstructured.Unstructured:
			spec, ok := dt.Object["spec"].(map[string]interface{})
			if ok {
				spec["clusterIP"] = co.Spec.ClusterIP
			}
		case *corev1.Service:
			dt.Spec.ClusterIP = co.Spec.ClusterIP
		default:
			// TODO ? Return error of mismatching types?
			return nil
		}
	}

	return nil
}

// BeforeUpdateFn
// func DesiredStateUpdateKeepServiceAccountTokenReferences(current, desired runtime.Object) error {
// 	if co, ok := current.(*corev1.ServiceAccount); ok {
// 		if do, ok := desired.(*corev1.ServiceAccount); ok {
// 			do.Secrets = co.Secrets
// 		} else {
// 			fmt.Printf("DEBUG>>\ncurrent:\n%#v\ndesired:\n%#v\n", current, desired)
// 		}
// 	}

// 	return nil
// }
func DesiredStateUpdateKeepServiceAccountTokenReferences(current, desired runtime.Object) error {
	if co, ok := current.(*corev1.ServiceAccount); ok {
		switch dt := desired.(type) {
		case *unstructured.Unstructured:
			dt.Object["secrets"] = co.Secrets
		case *corev1.ServiceAccount:
			dt.Secrets = co.Secrets
		default:
			// TODO ? Return error of mismatching types?
			return nil
		}
	}
	return nil
}
