package reconciler

import (
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type ComponentReconciler interface {
	Reconcile(runtime.Object) (*reconcile.Result, error)
	RegisterWatches(*builder.Builder)
}

// Dispatcher orchestrates reconciliation of multiple ComponentReconciler objects
// focusing on handing off reconciled object to all of its components and calculating an aggregated result to return.
// It requires a ResourceGetter callback and optionally can leverage a ResourceFilter and a CompletionHandler
type Dispatcher struct {
	client.Client
	Log                  logr.Logger
	ResourceGetter       func(req ctrl.Request) (runtime.Object, error)
	ResourceFilter       func(runtime.Object) (bool, error)
	CompletionHandler    func(runtime.Object, ctrl.Result) ctrl.Result
	ComponentReconcilers []ComponentReconciler
}

// Reconcile implements reconcile.Reconciler in a generic way from the controller-runtime library
func (r *Dispatcher) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	object, err := r.ResourceGetter(req)
	if err != nil {
		return reconcile.Result{}, err
	}
	if r.ResourceFilter != nil {
		shouldReconcile, err := r.ResourceFilter(object)
		if err != nil || !shouldReconcile {
			return reconcile.Result{}, err
		}
	}
	result, err := r.Handle(object)
	if err != nil {
		return result, err
	}
	if r.CompletionHandler != nil {
		return r.CompletionHandler(object, result), nil
	}
	return result, nil
}

// Handle receives a single object and dispatches it to all the components
// Components need to understand how to interpret the object
func (r *Dispatcher) Handle(object runtime.Object) (ctrl.Result, error) {
	combinedResult := &CombinedResult{}
	for _, cr := range r.ComponentReconcilers {
		combinedResult.Combine(cr.Reconcile(object))
	}
	return combinedResult.Result, combinedResult.Err
}

// RegisterWatches dispatches the watch registration builder to all its components
func (r *Dispatcher) RegisterWatches(b *builder.Builder) *builder.Builder {
	for _, cr := range r.ComponentReconcilers {
		cr.RegisterWatches(b)
	}
	return b
}
