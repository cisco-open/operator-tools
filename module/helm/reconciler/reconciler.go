// Copyright © 2019 Banzai Cloud
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

import (
	"fmt"
	"net/http"
	"time"

	"emperror.dev/errors"
	"github.com/go-logr/logr"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/storage/driver"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// HelmReleaseHooks implements a custom helm release strategy that can be used
// to fine tune the installation and removal of components based on helm charts
type HelmReleaseHooks interface {
	GetName() string
	GetNamespace() string
	GetValues(runtime.Object) map[string]interface{}
	IsReady() (bool, error)
	ShouldUninstall(object runtime.Object) bool
	ConfigureUpgrade(*action.Upgrade)
	ConfigureInstall(*action.Install)
	ConfigureUninstall(*action.Uninstall)
	RegisterWatches(*builder.Builder)
}

// GenericHelmReconciler implements reconciler.ComponentReconciler
// from github.com/banzaicloud/operator-tools/pkg/reconciler without depending on it explicitly
type GenericHelmReconciler struct {
	helmChart      *chart.Chart
	reconcileHooks HelmReleaseHooks
	actionConfig   *action.Configuration
	log            logr.Logger
}

type NonCachedDiscovery struct {
	discovery.DiscoveryInterface
}

func (n *NonCachedDiscovery) Fresh() bool {
	return false
}

func (n *NonCachedDiscovery) Invalidate() {
}

type Initializer struct {
	ClientConfig    clientcmd.ClientConfig
	RestConfig      *rest.Config
	DiscoveryClient discovery.CachedDiscoveryInterface
	RestMapper      meta.RESTMapper
	Namespace       string
	Log             logr.Logger
}

func (i *Initializer) ToRESTConfig() (*rest.Config, error) {
	return i.RestConfig, nil
}

func (i *Initializer) ToDiscoveryClient() (discovery.CachedDiscoveryInterface, error) {
	return i.DiscoveryClient, nil
}

func (i *Initializer) ToRESTMapper() (meta.RESTMapper, error) {
	return i.RestMapper, nil
}

func (i *Initializer) ToRawKubeConfigLoader() clientcmd.ClientConfig {
	return i.ClientConfig
}
func SetClientConfig(c clientcmd.ClientConfig) InitializerOption {
	return func(i *Initializer) error {
		i.ClientConfig = c
		return nil
	}
}
func SetRestConfig(r *rest.Config) InitializerOption {
	return func(i *Initializer) error {
		i.RestConfig = r
		return nil
	}
}
func SetCachedDiscovery(d discovery.CachedDiscoveryInterface) InitializerOption {
	return func(i *Initializer) error {
		i.DiscoveryClient = d
		return nil
	}
}
func SetNonCachedDiscovery(d discovery.DiscoveryInterface) InitializerOption {
	return func(i *Initializer) error {
		i.DiscoveryClient = &NonCachedDiscovery{
			DiscoveryInterface: d,
		}
		return nil
	}
}
func SetMapper(m meta.RESTMapper) InitializerOption {
	return func(i *Initializer) error {
		i.RestMapper = m
		return nil
	}
}
func SetLog(logger logr.Logger) InitializerOption {
	return func(initializer *Initializer) error {
		initializer.Log = logger
		return nil
	}
}
func WithLog(name string) InitializerOption {
	return func(initializer *Initializer) error {
		if initializer.Log == nil {
			return errors.New("logger has not yet been initialized")
		}
		initializer.Log.WithName(name)
		return nil
	}
}
func WithNamespace(namespace string) InitializerOption {
	return func(initializer *Initializer) error {
		initializer.Namespace = namespace
		return nil
	}
}
func CreateAllFromClientConfig() InitializerOption {
	return func(initializer *Initializer) (err error) {
		if initializer.ClientConfig == nil {
			return errors.New("client config has not yet been initialized")
		}
		initializer.Namespace, _, err = initializer.ClientConfig.Namespace()
		if err != nil {
			return errors.WrapIf(err, "failed to get namespace from client config")
		}
		initializer.RestConfig, err = initializer.ClientConfig.ClientConfig()
		if err != nil {
			return errors.WrapIf(err, "failed to get rest config")
		}
		initializer.RestMapper, err = apiutil.NewDynamicRESTMapper(initializer.RestConfig, apiutil.WithLazyDiscovery)
		if err != nil {
			return errors.WrapIf(err, "failed to initialize rest mapper from rest config")
		}
		d, err := discovery.NewDiscoveryClientForConfig(initializer.RestConfig)
		if err != nil {
			return errors.WrapIf(err, "failed to initialize discovery client from rest config")
		}
		initializer.DiscoveryClient = &NonCachedDiscovery{
			DiscoveryInterface: d,
		}
		return nil
	}
}

func Init(helmChart http.File, opts ...InitializerOption) (*GenericHelmReconciler, error) {
	i := &Initializer{}
	for _, o := range opts {
		err := o(i)
		if err != nil {
			return nil, err
		}
	}
	actionConfig := &action.Configuration{}
	if err := actionConfig.Init(i, i.Namespace, "secret", func(format string, v ...interface{}) {
		i.Log.Info(fmt.Sprintf(format, v...))
	}); err != nil {
		return nil, errors.WrapIf(err, "failed to initialize helm action config")
	}
	archive, err := loader.LoadArchive(helmChart)
	if err != nil {
		return nil, errors.WrapIf(err, "failed to load chart")
	}
	return &GenericHelmReconciler{
		actionConfig: actionConfig,
		helmChart:    archive,
		log:          i.Log,
	}, nil
}

type InitializerOption func(*Initializer) error

func InitializerOptions(opts ...InitializerOption) InitializerOption {
	return func(initializer *Initializer) error {
		for _, o := range opts {
			if err := o(initializer); err != nil {
				return err
			}
		}
		return nil
	}
}

func Preset(c clientcmd.ClientConfig) InitializerOption {
	return InitializerOptions(
		SetClientConfig(c), CreateAllFromClientConfig(), SetLog(log.Log.WithName("helm")),
	)
}

func (hr *GenericHelmReconciler) SetReconcileHooks(hooks HelmReleaseHooks) *GenericHelmReconciler {
	hr.reconcileHooks = hooks
	return hr
}

func (hr *GenericHelmReconciler) Reconcile(object runtime.Object) (*reconcile.Result, error) {
	vals := map[string]interface{}{}
	if hr.reconcileHooks != nil {
		vals = hr.reconcileHooks.GetValues(object)
	}

	name := hr.helmChart.Name()
	if hr.reconcileHooks != nil {
		name = hr.reconcileHooks.GetName()
	}

	meta, err := meta.Accessor(object)
	if err != nil {
		return nil, errors.WrapIff(err, "failed to access metadata from object %+v", object)
	}

	namespace := meta.GetNamespace()
	if hr.reconcileHooks != nil {
		namespace = hr.reconcileHooks.GetNamespace()
	}

	if hr.reconcileHooks != nil {
		if hr.reconcileHooks.ShouldUninstall(object) {
			uninstall := action.NewUninstall(hr.actionConfig)
			uninstall.Timeout = time.Minute * 5
			uninstall.KeepHistory = true
			if hr.reconcileHooks != nil {
				hr.reconcileHooks.ConfigureUninstall(uninstall)
			}
			_, err := uninstall.Run(name)
			if err != nil {
				return nil, errors.WrapIff(err, "failed to uninstall chart %s", name)
			}
			return nil, nil
		}
	}

	history := action.NewHistory(hr.actionConfig)
	history.Max = 1
	_, err = history.Run(name)
	if err == driver.ErrReleaseNotFound {
		install := action.NewInstall(hr.actionConfig)
		hr.log.Info(fmt.Sprintf("release %s will be installed", name))
		install.Timeout = time.Minute * 5
		install.Wait = true
		install.Namespace = namespace
		install.ReleaseName = name
		if hr.reconcileHooks != nil {
			hr.reconcileHooks.ConfigureInstall(install)
		}
		_, err := install.Run(hr.helmChart, vals)
		if err != nil {
			return nil, errors.WrapIff(err, "failed to install chart %s", name)
		}
		hr.log.Info(fmt.Sprintf("release %s has been installed", name))
	} else if err != nil {
		return nil, errors.WrapIf(err, "failed to get release history")
	} else {
		upgrade := action.NewUpgrade(hr.actionConfig)
		hr.log.Info(fmt.Sprintf("release %s will be upgraded", name))
		upgrade.Namespace = namespace
		upgrade.Wait = true
		upgrade.Timeout = time.Minute * 5
		if hr.reconcileHooks != nil {
			hr.reconcileHooks.ConfigureUpgrade(upgrade)
		}
		_, err := upgrade.Run(name, hr.helmChart, vals)
		if err != nil {
			return nil, errors.WrapIff(err, "failed to upgrade chart %s", name)
		}
	}

	if hr.reconcileHooks != nil {
		ready, err := hr.reconcileHooks.IsReady()
		if err != nil {
			return nil, errors.WrapIff(err, "failed to detect ready state for chart %s", hr.helmChart.Name())
		}
		if ready {
			return nil, nil
		} else {
			return &reconcile.Result{Requeue: true}, nil
		}
	}

	return nil, nil
}

func (hr *GenericHelmReconciler) RegisterWatches(b *builder.Builder) {
	if hr.reconcileHooks != nil {
		hr.reconcileHooks.RegisterWatches(b)
	}
}
