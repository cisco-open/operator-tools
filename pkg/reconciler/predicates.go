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

	"github.com/banzaicloud/k8s-objectmatcher/patch"
	"github.com/banzaicloud/operator-tools/pkg/types"
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

	return false
}

type SpecChangePredicate struct {
	predicate.Funcs
}

func (SpecChangePredicate) Update(e event.UpdateEvent) bool {
	e.MetaNew.SetResourceVersion(e.MetaOld.GetResourceVersion())
	patchResult, err := patch.DefaultPatchMaker.Calculate(e.ObjectOld, e.ObjectNew, patch.IgnoreStatusFields(), IgnoreManagedFields())
	if err != nil {
		return true
	} else if patchResult.IsEmpty() {
		return false
	}

	return true
}
