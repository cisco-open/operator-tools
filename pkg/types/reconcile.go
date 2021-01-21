package types

const (
	NameLabel      = "app.kubernetes.io/name"
	InstanceLabel  = "app.kubernetes.io/instance"
	VersionLabel   = "app.kubernetes.io/version"
	ComponentLabel = "app.kubernetes.io/component"
	ManagedByLabel = "app.kubernetes.io/managed-by"

	BanzaiCloudManagedComponent = "banzaicloud.io/managed-component"
	BanzaiCloudOwnedBy          = "banzaicloud.io/owned-by"
	BanzaiCloudRelatedTo        = "banzaicloud.io/related-to"
)

type ReconcileStatus string

const (
	// Used for components and for aggregated status
	ReconcileStatusFailed ReconcileStatus = "Failed"

	// Used for components and for aggregated status
	ReconcileStatusReconciling ReconcileStatus = "Reconciling"

	// Used for components
	ReconcileStatusAvailable ReconcileStatus = "Available"
	ReconcileStatusUnmanaged ReconcileStatus = "Unmanaged"
	ReconcileStatusRemoved   ReconcileStatus = "Removed"

	// Used for aggregated status if all the components are stableized (Available, Unmanaged or Removed)
	ReconcileStatusSucceeded ReconcileStatus = "Succeeded"

	// Used to trigger reconciliation for a resource that otherwise ignores status changes, but listens to the Pending state
	// See PendingStatusPredicate in pkg/reconciler
	ReconcileStatusPending ReconcileStatus = "Pending"
)

func (s ReconcileStatus) Stable() bool {
	return s == ReconcileStatusUnmanaged || s == ReconcileStatusRemoved || s == ReconcileStatusAvailable
}

func (s ReconcileStatus) Available() bool {
	return s == ReconcileStatusAvailable || s == ReconcileStatusSucceeded
}

func (s ReconcileStatus) Failed() bool {
	return s == ReconcileStatusFailed
}

func (s ReconcileStatus) Pending() bool {
	return s == ReconcileStatusReconciling || s == ReconcileStatusPending
}

// Computes an aggregated state based on component statuses
func AggregatedState(componentStatuses []ReconcileStatus) ReconcileStatus {
	overallStatus := ReconcileStatusReconciling
	statusMap := make(map[ReconcileStatus]bool)
	hasUnstable := false
	for _, cs := range componentStatuses {
		if cs != "" {
			statusMap[cs] = true
		}
		if !(cs == "" || cs.Stable()) {
			hasUnstable = true
		}
	}

	if statusMap[ReconcileStatusFailed] {
		overallStatus = ReconcileStatusFailed
	} else if statusMap[ReconcileStatusReconciling] {
		overallStatus = ReconcileStatusReconciling
	}

	if !hasUnstable {
		overallStatus = ReconcileStatusSucceeded
	}
	return overallStatus
}

