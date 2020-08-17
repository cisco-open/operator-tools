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

import (
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/builder"
)

// HelmReleaseHooks implements a custom helm release strategy that can be used
// to fine tune the installation and removal of components based on helm charts
type HelmReleaseHooks interface {
	GetName() string
	GetNamespace() string
	GetValues() (map[string]interface{}, error)
	IsReady() (bool, error)
	ShouldUninstall() bool
	ConfigureUpgrade(*action.Upgrade)
	ConfigureInstall(*action.Install)
	ConfigureUninstall(*action.Uninstall)
}

type DefaultReleaseHooks struct {
	Object    v1.Object
	Chart     *chart.Chart
	Uninstall bool
}

func (d *DefaultReleaseHooks) GetName() string {
	return d.Object.GetName() + "-" + d.Chart.Name()
}

func (d *DefaultReleaseHooks) GetNamespace() string {
	return d.Object.GetNamespace()
}

func (d *DefaultReleaseHooks) GetValues() (map[string]interface{}, error) {
	return nil, nil
}

func (d *DefaultReleaseHooks) IsReady() (bool, error) {
	return true, nil
}

func (d *DefaultReleaseHooks) ShouldUninstall() bool {
	return false
}

func (d *DefaultReleaseHooks) ConfigureUpgrade(*action.Upgrade) {
}

func (d *DefaultReleaseHooks) ConfigureInstall(*action.Install) {
}

func (d *DefaultReleaseHooks) ConfigureUninstall(*action.Uninstall) {
}

func (d *DefaultReleaseHooks) RegisterWatches(*builder.Builder) {
}