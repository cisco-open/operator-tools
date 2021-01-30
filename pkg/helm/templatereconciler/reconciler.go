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

package templatereconciler

import (
	"context"
	"net/http"
	"time"

	"emperror.dev/errors"
	"github.com/banzaicloud/operator-tools/pkg/inventory"
	"github.com/banzaicloud/operator-tools/pkg/logger"
	"github.com/banzaicloud/operator-tools/pkg/reconciler"
	"github.com/banzaicloud/operator-tools/pkg/resources"
	"github.com/banzaicloud/operator-tools/pkg/types"
	"github.com/go-logr/logr"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/discovery"
	controllerruntime "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type ReleaseData struct {
	Chart       http.FileSystem
	Values      map[string]interface{}
	Namespace   string
	ChartName   string
	ReleaseName string
	// Layers can be embedded into CRDs directly to provide flexible override mechanisms
	Layers []resources.K8SResourceOverlay
	// Modifiers can be used from client code to modify resources before being applied
	Modifiers []resources.ObjectModifierFunc
}

type Component interface {
	Name() string
	Skipped(runtime.Object) bool
	Enabled(runtime.Object) bool
	PreChecks(runtime.Object) error
	ReleaseData(runtime.Object) (*ReleaseData, error)
	UpdateStatus(object runtime.Object, status types.ReconcileStatus, message string) error
}

type HelmReconciler struct {
	client       client.Client
	scheme       *runtime.Scheme
	logger       logr.Logger
	inventory    *inventory.Inventory
	opts         []reconciler.NativeReconcilerOpt
	objectParser *resources.ObjectParser
}

func NewHelmReconciler(
	client client.Client,
	scheme *runtime.Scheme,
	logger logr.Logger,
	discovery discovery.DiscoveryInterface,
	opts []reconciler.NativeReconcilerOpt,
) *HelmReconciler {
	r := &HelmReconciler{
		client:       client,
		scheme:       scheme,
		logger:       logger,
		inventory:    inventory.NewDiscoveryInventory(client, logger, discovery),
		objectParser: resources.NewObjectParser(scheme),
		opts:         opts,
	}
	return r
}

func (rec *HelmReconciler) Reconcile(object runtime.Object, component Component) (*reconcile.Result, error) {
	var ok bool
	var parent reconciler.ResourceOwner
	if parent, ok = object.(reconciler.ResourceOwner); !ok {
		return nil, errors.New("cannot convert object to ResourceOwner interface")
	}

	if component.Skipped(object) {
		return &reconcile.Result{}, component.UpdateStatus(object, types.ReconcileStatusUnmanaged, "")
	}

	if err := component.UpdateStatus(object, types.ReconcileStatusReconciling, ""); err != nil {
		rec.logger.Error(err, "status update failed")
	}
	rec.logger.Info("reconciling")

	if component.Enabled(object) {
		if err := component.PreChecks(object); err != nil {
			if err := component.UpdateStatus(object, types.ReconcileStatusReconciling, "waiting for precondition checks to pass"); err != nil {
				rec.logger.Error(err, "status update failed")
			}
			rec.logger.Error(err, "precondition checks failed")
			return &reconcile.Result{
				RequeueAfter: time.Second * 5,
			}, nil
		}
	}

	defer logger.EnableGroupSession(rec.logger)()

	rec.logger.Info("syncing resources")

	releaseData, err := component.ReleaseData(object)
	if err != nil {
		return nil, errors.WrapIf(err, "failed to get release data")
	}

	result, err := rec.reconcile(parent, component, releaseData)
	if err != nil {
		uerr := component.UpdateStatus(object, types.ReconcileStatusFailed, err.Error())
		if uerr != nil {
			rec.logger.Error(uerr, "status update failed")
		}
		return result, err
	} else {
		if component.Skipped(object) {
			err = component.UpdateStatus(object, types.ReconcileStatusUnmanaged, "")
			if err != nil {
				return result, err
			}
		} else if component.Enabled(object) {
			err = component.UpdateStatus(object, types.ReconcileStatusAvailable, "")
			if err != nil {
				return result, err
			}
		} else {
			err = component.UpdateStatus(object, types.ReconcileStatusRemoved, "")
			if err != nil {
				return result, err
			}
		}
	}

	return result, err
}

func (rec *HelmReconciler) reconcile(parent reconciler.ResourceOwner, component Component, releaseData *ReleaseData) (*reconcile.Result, error) {
	resourceBuilders, err := reconciler.GetResourceBuildersFromObjects([]runtime.Object{
		&v1.Namespace{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Namespace",
				APIVersion: "v1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: releaseData.Namespace,
			},
		},
	}, reconciler.StateCreated)

	if err != nil {
		return nil, err
	}

	if component.Enabled(parent) {
		objects, state, err := orderedChartObjectsWithState(releaseData)
		if err != nil {
			return nil, err
		}

		modifiers := releaseData.Modifiers

		for _, layer := range releaseData.Layers {
			modifier, err := resources.PatchYAMLModifier(layer, rec.objectParser)
			if err != nil {
				return nil, errors.WrapIf(err, "failed to create modifier from layer")
			}
			modifiers = append(modifiers, modifier)
		}

		chartResourceBuilders, err := reconciler.GetResourceBuildersFromObjects(objects, state, modifiers...)
		if err != nil {
			return nil, err
		}

		resourceBuilders = rec.inventory.Append(releaseData.Namespace, releaseData.ReleaseName, parent, append(resourceBuilders, chartResourceBuilders...))
	} else {
		resourceBuilders = rec.inventory.Append(releaseData.Namespace, releaseData.ReleaseName, parent, resourceBuilders)
	}

	r := reconciler.NewNativeReconcilerWithDefaults(
		component.Name(),
		rec.client,
		rec.scheme,
		rec.logger,
		func(_ reconciler.ResourceOwner, _ interface{}) []reconciler.ResourceBuilder {
			return resourceBuilders
		},
		rec.inventory.TypesToPurge,
		func(_ runtime.Object) (reconciler.ResourceOwner, interface{}) {
			return nil, nil
		},
		append(rec.opts, reconciler.NativeReconcilerWithScheme(rec.scheme))...)

	result, err := r.Reconcile(parent)
	if err != nil {
		return result, err
	}

	if !component.Enabled(parent) {
		// cleanup orphaned pods left from removed jobs
		if err := rec.client.DeleteAllOf(context.TODO(), &v1.Pod{},
			client.MatchingLabels{"release": releaseData.ReleaseName},
			client.HasLabels{"job-name"},
			client.InNamespace(releaseData.Namespace),
		); err != nil {
			return result, errors.WrapIf(err, "failed to remove pods left from the release")
		}
	}

	rec.logger.Info("reconciled")

	return result, nil
}

func (rec HelmReconciler) RegisterWatches(_ *controllerruntime.Builder) {}
