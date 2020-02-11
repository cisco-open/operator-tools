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
