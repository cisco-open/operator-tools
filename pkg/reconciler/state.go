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
	"k8s.io/apimachinery/pkg/runtime"
	runtimeClient "sigs.k8s.io/controller-runtime/pkg/client"
)

type DynamicDesiredState struct {
	DesiredState     DesiredState
	BeforeCreateFunc func(desired runtime.Object) error
	BeforeUpdateFunc func(current, desired runtime.Object) error
	BeforeDeleteFunc func(current runtime.Object) error
	CreateOptions    []runtimeClient.CreateOption
	UpdateOptions    []runtimeClient.UpdateOption
	DeleteOptions    []runtimeClient.DeleteOption
	ShouldCreateFunc func(desired runtime.Object) (bool, error)
	ShouldUpdateFunc func(current, desired runtime.Object) (bool, error)
	ShouldDeleteFunc func(desired runtime.Object) (bool, error)
}

func (s DynamicDesiredState) GetDesiredState() DesiredState {
	return s.DesiredState
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
	if s.BeforeUpdateFunc != nil {
		return s.BeforeUpdateFunc(current, desired)
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

type MultipleDesiredStates []DesiredState

func (s MultipleDesiredStates) GetDesiredState() DesiredState {
	ds := s[len(s)-1]
	if ds, ok := ds.(DesiredStateWithStaticState); ok {
		return ds.DesiredState()
	}
	if s, ok := ds.(interface {
		GetDesiredState() DesiredState
	}); ok {
		return s.GetDesiredState()
	}

	return ds
}

func (s MultipleDesiredStates) BeforeCreate(desired runtime.Object) error {
	for _, ds := range s {
		err := ds.BeforeCreate(desired)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s MultipleDesiredStates) BeforeUpdate(current, desired runtime.Object) error {
	for _, ds := range s {
		err := ds.BeforeUpdate(current, desired)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s MultipleDesiredStates) BeforeDelete(desired runtime.Object) error {
	for _, ds := range s {
		err := ds.BeforeDelete(desired)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s MultipleDesiredStates) GetCreateptions() []runtimeClient.CreateOption {
	var createOptions []runtimeClient.CreateOption
	for _, ds := range s {
		if ds, ok := ds.(DesiredStateWithCreateOptions); ok {
			createOptions = append(createOptions, ds.GetCreateOptions()...)
		}
	}

	return createOptions
}

func (s MultipleDesiredStates) GetUpdateOptions() []runtimeClient.UpdateOption {
	var updateOptions []runtimeClient.UpdateOption
	for _, ds := range s {
		if ds, ok := ds.(DesiredStateWithUpdateOptions); ok {
			updateOptions = append(updateOptions, ds.GetUpdateOptions()...)
		}
	}

	return updateOptions
}

func (s MultipleDesiredStates) GetDeleteOptions() []runtimeClient.DeleteOption {
	var deleteOptions []runtimeClient.DeleteOption
	for _, ds := range s {
		if ds, ok := ds.(DesiredStateWithDeleteOptions); ok {
			deleteOptions = append(deleteOptions, ds.GetDeleteOptions()...)
		}
	}

	return deleteOptions
}

func (s MultipleDesiredStates) ShouldCreate(desired runtime.Object) (bool, error) {
	for _, ds := range s {
		if s, ok := ds.(DesiredStateShouldCreate); ok {
			should, err := s.ShouldCreate(desired)
			if err != nil {
				return should, err
			}
			if !should {
				return should, nil
			}
		}
	}

	return true, nil
}

func (s MultipleDesiredStates) ShouldUpdate(current, desired runtime.Object) (bool, error) {
	for _, ds := range s {
		if s, ok := ds.(DesiredStateShouldUpdate); ok {
			should, err := s.ShouldUpdate(current, desired)
			if err != nil {
				return should, err
			}
			if !should {
				return should, nil
			}
		}
	}

	return true, nil
}

func (s MultipleDesiredStates) ShouldDelete(desired runtime.Object) (bool, error) {
	for _, ds := range s {
		if s, ok := ds.(DesiredStateShouldDelete); ok {
			should, err := s.ShouldDelete(desired)
			if err != nil {
				return should, err
			}
			if !should {
				return should, nil
			}
		}
	}

	return true, nil
}
