## ObjectKey

### name (string, optional) {#objectkey-name}


### namespace (string, optional) {#objectkey-namespace}



## EnabledComponent

EnabledComponent implements the "enabled component" pattern
Embed this type into other component types to avoid unnecessary code duplication
NOTE: Don't forget to annotate the embedded field with `json:",inline"` tag for controller-gen

### enabled (*bool, optional) {#enabledcomponent-enabled}



## MetaBase

Deprecated
Consider using ObjectMeta in the typeoverrides package combined with the merge package

### annotations (map[string]string, optional) {#metabase-annotations}


### labels (map[string]string, optional) {#metabase-labels}



## PodTemplateBase

Deprecated
Consider using PodTemplateSpec in the typeoverrides package combined with the merge package

### metadata (*MetaBase, optional) {#podtemplatebase-metadata}


### spec (*PodSpecBase, optional) {#podtemplatebase-spec}



## ContainerBase

Deprecated
Consider using Container in the typeoverrides package combined with the merge package

### command ([]string, optional) {#containerbase-command}


### image (string, optional) {#containerbase-image}


### livenessProbe (*corev1.Probe, optional) {#containerbase-livenessprobe}


### name (string, optional) {#containerbase-name}


### pullPolicy (corev1.PullPolicy, optional) {#containerbase-pullpolicy}


### readinessProbe (*corev1.Probe, optional) {#containerbase-readinessprobe}


### resources (*corev1.ResourceRequirements, optional) {#containerbase-resources}


### securityContext (*corev1.SecurityContext, optional) {#containerbase-securitycontext}


### volumeMounts ([]corev1.VolumeMount, optional) {#containerbase-volumemounts}



## PodSpecBase

Deprecated
Consider using PodSpec in the typeoverrides package combined with the merge package

### affinity (*corev1.Affinity, optional) {#podspecbase-affinity}


### containers ([]ContainerBase, optional) {#podspecbase-containers}


### imagePullSecrets ([]corev1.LocalObjectReference, optional) {#podspecbase-imagepullsecrets}


### initContainers ([]ContainerBase, optional) {#podspecbase-initcontainers}


### nodeSelector (map[string]string, optional) {#podspecbase-nodeselector}


### priorityClassName (string, optional) {#podspecbase-priorityclassname}


### securityContext (*corev1.PodSecurityContext, optional) {#podspecbase-securitycontext}


### serviceAccountName (string, optional) {#podspecbase-serviceaccountname}


### tolerations ([]corev1.Toleration, optional) {#podspecbase-tolerations}


### volumes ([]corev1.Volume, optional) {#podspecbase-volumes}



## DeploymentBase

Deprecated
Consider using Deployment in the typeoverrides package combined with the merge package

###  (*MetaBase, required) {#deploymentbase-}


### spec (*DeploymentSpecBase, optional) {#deploymentbase-spec}



## DeploymentSpecBase

Deprecated
Consider using DeploymentSpec in the typeoverrides package combined with the merge package

### replicas (*int32, optional) {#deploymentspecbase-replicas}


### selector (*metav1.LabelSelector, optional) {#deploymentspecbase-selector}


### strategy (*appsv1.DeploymentStrategy, optional) {#deploymentspecbase-strategy}


### template (*PodTemplateBase, optional) {#deploymentspecbase-template}



## StatefulSetBase

Deprecated
Consider using StatefulSet in the typeoverrides package combined with the merge package

###  (*MetaBase, required) {#statefulsetbase-}


### spec (*StatefulsetSpecBase, optional) {#statefulsetbase-spec}



## StatefulsetSpecBase

Deprecated
Consider using StatefulSetSpec in the typeoverrides package combined with the merge package

### podManagementPolicy (appsv1.PodManagementPolicyType, optional) {#statefulsetspecbase-podmanagementpolicy}


### replicas (*int32, optional) {#statefulsetspecbase-replicas}


### selector (*metav1.LabelSelector, optional) {#statefulsetspecbase-selector}


### template (*PodTemplateBase, optional) {#statefulsetspecbase-template}


### updateStrategy (*appsv1.StatefulSetUpdateStrategy, optional) {#statefulsetspecbase-updatestrategy}



## DaemonSetBase

Deprecated
Consider using DaemonSet in the typeoverrides package combined with the merge package

###  (*MetaBase, required) {#daemonsetbase-}


### spec (*DaemonSetSpecBase, optional) {#daemonsetbase-spec}



## DaemonSetSpecBase

Deprecated
Consider using DaemonSetSpec in the typeoverrides package combined with the merge package

### minReadySeconds (int32, optional) {#daemonsetspecbase-minreadyseconds}


### revisionHistoryLimit (*int32, optional) {#daemonsetspecbase-revisionhistorylimit}


### selector (*metav1.LabelSelector, optional) {#daemonsetspecbase-selector}


### template (*PodTemplateBase, optional) {#daemonsetspecbase-template}


### updateStrategy (*appsv1.DaemonSetUpdateStrategy, optional) {#daemonsetspecbase-updatestrategy}



