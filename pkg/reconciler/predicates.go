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
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	"github.com/cisco-open/k8s-objectmatcher/patch"
	"github.com/cisco-open/operator-tools/pkg/types"
)

type SkipCreatePredicate struct {
	predicate.Funcs
}

func (SkipCreatePredicate) Create(e event.CreateEvent) bool {
	return false
}

type SkipUpdatePredicate struct {
	predicate.Funcs
}

func (SkipUpdatePredicate) Update(e event.UpdateEvent) bool {
	return false
}

type SkipDeletePredicate struct {
	predicate.Funcs
}

func (SkipDeletePredicate) Delete(e event.DeleteEvent) bool {
	return false
}

type PendingStatusPredicate struct {
	predicate.Funcs
}

func (PendingStatusPredicate) Update(e event.UpdateEvent) bool {
	if o, ok := e.ObjectNew.(interface {
		IsAnyInState(state types.ReconcileStatus) bool
	}); ok {
		return o.IsAnyInState(types.ReconcileStatusPending)
	}

	if o, ok := e.ObjectNew.(interface {
		IsPending() bool
	}); ok {
		return o.IsPending()
	}

	return false
}

type SpecChangePredicate struct {
	predicate.Funcs

	patchMaker       patch.Maker
	calculateOptions []patch.CalculateOption
}

func (p SpecChangePredicate) Update(e event.UpdateEvent) bool {
	if p.patchMaker == nil {
		p.patchMaker = patch.DefaultPatchMaker
	}
	if p.calculateOptions == nil {
		p.calculateOptions = []patch.CalculateOption{patch.IgnoreStatusFields(), IgnoreManagedFields()}
	}

	oldRV := e.ObjectOld.GetResourceVersion()
	e.ObjectOld.SetResourceVersion(e.ObjectNew.GetResourceVersion())
	defer e.ObjectOld.SetResourceVersion(oldRV)

	patchResult, err := p.patchMaker.Calculate(e.ObjectOld, e.ObjectNew, p.calculateOptions...)
	if err != nil {
		return true
	} else if patchResult.IsEmpty() {
		return false
	}

	return true
}
