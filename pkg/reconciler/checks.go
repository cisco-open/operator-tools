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
	"context"

	"emperror.dev/errors"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/banzaicloud/operator-tools/pkg/wait"
)

func IstioSidecarInjectorExistsCheck(c client.Client, namespace string) wait.CustomResourceConditionCheck {
	return func() (bool, error) {
		var pods corev1.PodList
		err := c.List(context.Background(), &pods, client.InNamespace(namespace), client.MatchingLabels(map[string]string{
			"istio": "sidecar-injector",
		}))
		if err != nil {
			return false, errors.WrapIf(err, "could not list pods")
		}
		if len(pods.Items) == 0 {
			err = c.List(context.Background(), &pods, client.InNamespace(namespace), client.MatchingLabels(map[string]string{
				"istio": "pilot",
			}))
			if err != nil {
				return false, errors.WrapIf(err, "could not list pods")
			}
		}

		for _, pod := range pods.Items {
			if pod.Status.Phase == corev1.PodRunning {
				for _, cs := range pod.Status.ContainerStatuses {
					if !cs.Ready {
						return false, nil
					}
				}
				return true, nil
			}
		}

		return false, nil
	}
}
