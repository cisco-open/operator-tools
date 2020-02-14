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
	"time"

	"emperror.dev/errors"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type ReconcileRetry struct {
	MaxRetries  int
	DefaultWait time.Duration
}

func (r *ReconcileRetry) Reconcile(component func() (*reconcile.Result, error)) error {
	i := 0
	for ; i < r.MaxRetries || r.MaxRetries == -1; i++ {
		result, err := component()
		if err != nil {
			return err
		}
		if result == nil {
			return nil
		}
		if !result.Requeue && result.RequeueAfter == 0 {
			break
		} else {
			if result.RequeueAfter > 0 {
				time.Sleep(result.RequeueAfter)
			} else {
				time.Sleep(r.DefaultWait)
			}
		}
	}
	return errors.Errorf("reconciliation did not complete after %d retries", i)
}
