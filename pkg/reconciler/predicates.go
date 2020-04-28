package reconciler

import (
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
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

func (SkipCreatePredicate) Update(e event.UpdateEvent) bool {
	return false
}

type SkipDeletePredicate struct {
	predicate.Funcs
}

func (SkipCreatePredicate) Delete(e event.DeleteEvent) bool {
	return false
}
