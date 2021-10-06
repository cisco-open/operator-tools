## ObjectKey

### name (string, optional) {#objectkey-name}

Default: -

### namespace (string, optional) {#objectkey-namespace}

Default: -


## EnabledComponent

EnabledComponent implements the "enabled component" pattern
Embed this type into other component types to avoid unnecessary code duplication

### enabled (*bool, optional) {#enabledcomponent-enabled}

Default: -


## MetaBase

Deprecated
Consider using ObjectMeta in the typeoverrides package combined with the merge package

### annotations (map[string]string, optional) {#metabase-annotations}

Default: -

### labels (map[string]string, optional) {#metabase-labels}

Default: -


## PodTemplateBase

Deprecated
Consider using PodTemplateSpec in the typeoverrides package combined with the merge package

### metadata (*MetaBase, optional) {#podtemplatebase-metadata}

Default: -

### spec (*PodSpecBase, optional) {#podtemplatebase-spec}

Default: -


## ContainerBase

Deprecated
Consider using Container in the typeoverrides package combined with the merge package

### name (string, optional) {#containerbase-name}

Default: -

### resources (*corev1.ResourceRequirements, optional) {#containerbase-resources}

Default: -

### image (string, optional) {#containerbase-image}

Default: -

### pullPolicy (corev1.PullPolicy, optional) {#containerbase-pullpolicy}

Default: -

### command ([]string, optional) {#containerbase-command}

Default: -

### volumeMounts ([]corev1.VolumeMount, optional) {#containerbase-volumemounts}

Default: -

### securityContext (*corev1.SecurityContext, optional) {#containerbase-securitycontext}

Default: -

### livenessProbe (*corev1.Probe, optional) {#containerbase-livenessprobe}

Default: -

### readinessProbe (*corev1.Probe, optional) {#containerbase-readinessprobe}

Default: -


## PodSpecBase

Deprecated
Consider using PodSpec in the typeoverrides package combined with the merge package

### tolerations ([]corev1.Toleration, optional) {#podspecbase-tolerations}

Default: -

### nodeSelector (map[string]string, optional) {#podspecbase-nodeselector}

Default: -

### serviceAccountName (string, optional) {#podspecbase-serviceaccountname}

Default: -

### affinity (*corev1.Affinity, optional) {#podspecbase-affinity}

Default: -

### securityContext (*corev1.PodSecurityContext, optional) {#podspecbase-securitycontext}

Default: -

### volumes ([]corev1.Volume, optional) {#podspecbase-volumes}

Default: -

### priorityClassName (string, optional) {#podspecbase-priorityclassname}

Default: -

### containers ([]ContainerBase, optional) {#podspecbase-containers}

Default: -

### initContainers ([]ContainerBase, optional) {#podspecbase-initcontainers}

Default: -

### imagePullSecrets ([]corev1.LocalObjectReference, optional) {#podspecbase-imagepullsecrets}

Default: -


## DeploymentBase

Deprecated
Consider using Deployment in the typeoverrides package combined with the merge package

###  (*MetaBase, required) {#deploymentbase-}

Default: -

### spec (*DeploymentSpecBase, optional) {#deploymentbase-spec}

Default: -


## DeploymentSpecBase

Deprecated
Consider using DeploymentSpec in the typeoverrides package combined with the merge package

### replicas (*int32, optional) {#deploymentspecbase-replicas}

Default: -

### selector (*metav1.LabelSelector, optional) {#deploymentspecbase-selector}

Default: -

### strategy (*appsv1.DeploymentStrategy, optional) {#deploymentspecbase-strategy}

Default: -

### template (*PodTemplateBase, optional) {#deploymentspecbase-template}

Default: -


## StatefulSetBase

Deprecated
Consider using StatefulSet in the typeoverrides package combined with the merge package

###  (*MetaBase, required) {#statefulsetbase-}

Default: -

### spec (*StatefulsetSpecBase, optional) {#statefulsetbase-spec}

Default: -


## StatefulsetSpecBase

Deprecated
Consider using StatefulSetSpec in the typeoverrides package combined with the merge package

### replicas (*int32, optional) {#statefulsetspecbase-replicas}

Default: -

### selector (*metav1.LabelSelector, optional) {#statefulsetspecbase-selector}

Default: -

### podManagementPolicy (appsv1.PodManagementPolicyType, optional) {#statefulsetspecbase-podmanagementpolicy}

Default: -

### updateStrategy (*appsv1.StatefulSetUpdateStrategy, optional) {#statefulsetspecbase-updatestrategy}

Default: -

### template (*PodTemplateBase, optional) {#statefulsetspecbase-template}

Default: -


## DaemonSetBase

Deprecated
Consider using DaemonSet in the typeoverrides package combined with the merge package

###  (*MetaBase, required) {#daemonsetbase-}

Default: -

### spec (*DaemonSetSpecBase, optional) {#daemonsetbase-spec}

Default: -


## DaemonSetSpecBase

Deprecated
Consider using DaemonSetSpec in the typeoverrides package combined with the merge package

### selector (*metav1.LabelSelector, optional) {#daemonsetspecbase-selector}

Default: -

### updateStrategy (*appsv1.DaemonSetUpdateStrategy, optional) {#daemonsetspecbase-updatestrategy}

Default: -

### minReadySeconds (int32, optional) {#daemonsetspecbase-minreadyseconds}

Default: -

### revisionHistoryLimit (*int32, optional) {#daemonsetspecbase-revisionhistorylimit}

Default: -

### template (*PodTemplateBase, optional) {#daemonsetspecbase-template}

Default: -


