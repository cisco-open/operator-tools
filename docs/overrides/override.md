## ObjectMeta

ObjectMeta contains only a [subset of the fields included in k8s.io/apimachinery/pkg/apis/meta/v1.ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.19/#objectmeta-v1-meta).

### annotations (map[string]string, optional) {#objectmeta-annotations}

Default: -

### labels (map[string]string, optional) {#objectmeta-labels}

Default: -


## Service

Service is a subset of [Service in k8s.io/api/apps/v1](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.19/#service-v1-core) for embedding.

### metadata (ObjectMeta, optional) {#service-metadata}

Default: -

### spec (v1.ServiceSpec, optional) {#service-spec}

Kubernetes [Service Specification](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.19/#servicespec-v1-core) 

Default: -


## IngressExtensionsV1beta1

IngressExtensionsV1beta1 is a subset of Ingress k8s.io/api/extensions/v1beta1 but is already deprecated

### metadata (ObjectMeta, optional) {#ingressextensionsv1beta1-metadata}

Default: -

### spec (v1beta1.IngressSpec, optional) {#ingressextensionsv1beta1-spec}

Default: -


## IngressNetworkingV1beta1

IngressExtensionsV1beta1 is a subset of [Ingress in k8s.io/api/networking/v1beta1](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.19/#ingress-v1-networking-k8s-io).

### metadata (ObjectMeta, optional) {#ingressnetworkingv1beta1-metadata}

Default: -

### spec (networkingv1beta1.IngressSpec, optional) {#ingressnetworkingv1beta1-spec}

Kubernetes [Ingress Specification](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.19/#ingressclassspec-v1-networking-k8s-io) 

Default: -


## DaemonSet

DaemonSet is a subset of [DaemonSet in k8s.io/api/apps/v1](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.19/#daemonset-v1-apps), with [DaemonSetSpec replaced by the local variant](#daemonset-spec).

### metadata (ObjectMeta, optional) {#daemonset-metadata}

Default: -

### spec (DaemonSetSpec, optional) {#daemonset-spec}

[Local DaemonSet specification](#daemonset-spec) 

Default: -


## DaemonSetSpec

DaemonSetSpec is a subset of [DaemonSetSpec in k8s.io/api/apps/v1](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.19/#daemonsetspec-v1-apps) but with required fields declared as optional
and [PodTemplateSpec replaced by the local variant](#podtemplatespec).

### selector (*metav1.LabelSelector, optional) {#daemonsetspec-selector}

A label query over pods that are managed by the daemon set. 

Default: -

### template (PodTemplateSpec, optional) {#daemonsetspec-template}

An object that describes the pod that will be created. Note that this is a [local PodTemplateSpec](#podtemplatespec) 

Default: -

### updateStrategy (appsv1.DaemonSetUpdateStrategy, optional) {#daemonsetspec-updatestrategy}

An update strategy to replace existing DaemonSet pods with new pods. 

Default: -

### minReadySeconds (int32, optional) {#daemonsetspec-minreadyseconds}

The minimum number of seconds for which a newly created DaemonSet pod should be ready without any of its container crashing, for it to be considered available. Defaults to 0 (pod will be considered available as soon as it is ready).  

Default:  0

### revisionHistoryLimit (*int32, optional) {#daemonsetspec-revisionhistorylimit}

The number of old history to retain to allow rollback. This is a pointer to distinguish between explicit zero and not specified. Defaults to 10.  

Default:  10


## Deployment

Deployment is a subset of [Deployment in k8s.io/api/apps/v1](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.19/#deployment-v1-apps), with [DeploymentSpec replaced by the local variant](#deployment-spec).

### metadata (ObjectMeta, optional) {#deployment-metadata}

Default: -

### spec (DeploymentSpec, optional) {#deployment-spec}

The desired behavior of [this deployment](#deploymentspec). 

Default: -


## DeploymentSpec

DeploymentSpec is a subset of [DeploymentSpec in k8s.io/api/apps/v1](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.19/#deploymentspec-v1-apps) but with required fields declared as optional
and [PodTemplateSpec replaced by the local variant](#podtemplatespec).

### replicas (*int32, optional) {#deploymentspec-replicas}

Number of desired pods. This is a pointer to distinguish between explicit zero and not specified. Defaults to 1.  

Default:  1

### selector (*metav1.LabelSelector, optional) {#deploymentspec-selector}

Label selector for pods. Existing ReplicaSets whose pods are selected by this will be the ones affected by this deployment. It must match the pod template's labels. 

Default: -

### template (PodTemplateSpec, optional) {#deploymentspec-template}

An object that describes the pod that will be created. Note that this is a [local PodTemplateSpec](#podtemplatespec) 

Default: -

### strategy (appsv1.DeploymentStrategy, optional) {#deploymentspec-strategy}

The deployment strategy to use to replace existing pods with new ones. +patchStrategy=retainKeys 

Default: -

### minReadySeconds (int32, optional) {#deploymentspec-minreadyseconds}

Minimum number of seconds for which a newly created pod should be ready without any of its container crashing, for it to be considered available. Defaults to 0 (pod will be considered available as soon as it is ready)  

Default:  0

### revisionHistoryLimit (*int32, optional) {#deploymentspec-revisionhistorylimit}

The number of old ReplicaSets to retain to allow rollback. This is a pointer to distinguish between explicit zero and not specified. Defaults to 10.  

Default:  10

### paused (bool, optional) {#deploymentspec-paused}

Indicates that the deployment is paused. 

Default: -

### progressDeadlineSeconds (*int32, optional) {#deploymentspec-progressdeadlineseconds}

The maximum time in seconds for a deployment to make progress before it is considered to be failed. The deployment controller will continue to process failed deployments and a condition with a ProgressDeadlineExceeded reason will be surfaced in the deployment status. Note that progress will not be estimated during the time a deployment is paused. Defaults to 600s.  

Default:  600


## StatefulSet

StatefulSet is a subset of [StatefulSet in k8s.io/api/apps/v1](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.19/#statefulset-v1-apps), with [StatefulSetSpec replaced by the local variant](#statefulset-spec).

### metadata (ObjectMeta, optional) {#statefulset-metadata}

Default: -

### spec (StatefulSetSpec, optional) {#statefulset-spec}

Default: -


## StatefulSetSpec

StatefulSetSpec is a subset of [StatefulSetSpec in k8s.io/api/apps/v1](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.19/#statefulsetspec-v1-apps) but with required fields declared as optional
and [PodTemplateSpec](#podtemplatespec) and [PersistentVolumeClaim replaced by the local variant](#persistentvolumeclaim).

### replicas (*int32, optional) {#statefulsetspec-replicas}

Replicas is the desired number of replicas of the given Template. These are replicas in the sense that they are instantiations of the same Template, but individual replicas also have a consistent identity. If unspecified, defaults to 1.  +optional 

Default:  1

### selector (*metav1.LabelSelector, optional) {#statefulsetspec-selector}

Selector is a label query over pods that should match the replica count. It must match the pod template's labels. 

Default: -

### template (PodTemplateSpec, optional) {#statefulsetspec-template}

template is the object that describes the pod that will be created if insufficient replicas are detected. Each pod stamped out by the StatefulSet will fulfill this Template, but have a unique identity from the rest of the StatefulSet. 

Default: -

### volumeClaimTemplates ([]PersistentVolumeClaim, optional) {#statefulsetspec-volumeclaimtemplates}

volumeClaimTemplates is a list of claims that pods are allowed to reference. The StatefulSet controller is responsible for mapping network identities to claims in a way that maintains the identity of a pod. Every claim in this list must have at least one matching (by name) volumeMount in one container in the template. A claim in this list takes precedence over any volumes in the template, with the same name. +optional 

Default: -

### serviceName (string, optional) {#statefulsetspec-servicename}

serviceName is the name of the service that governs this StatefulSet. This service must exist before the StatefulSet, and is responsible for the network identity of the set. Pods get DNS/hostnames that follow the pattern: pod-specific-string.serviceName.default.svc.cluster.local where "pod-specific-string" is managed by the StatefulSet controller. 

Default: -

### podManagementPolicy (appsv1.PodManagementPolicyType, optional) {#statefulsetspec-podmanagementpolicy}

podManagementPolicy controls how pods are created during initial scale up, when replacing pods on nodes, or when scaling down. The default policy is `OrderedReady`, where pods are created in increasing order (pod-0, then pod-1, etc) and the controller will wait until each pod is ready before continuing. When scaling down, the pods are removed in the opposite order. The alternative policy is `Parallel` which will create pods in parallel to match the desired scale without waiting, and on scale down will delete all pods at once.  +optional 

Default:  OrderedReady

### updateStrategy (appsv1.StatefulSetUpdateStrategy, optional) {#statefulsetspec-updatestrategy}

updateStrategy indicates the StatefulSetUpdateStrategy that will be employed to update Pods in the StatefulSet when a revision is made to Template. 

Default: -

### revisionHistoryLimit (*int32, optional) {#statefulsetspec-revisionhistorylimit}

revisionHistoryLimit is the maximum number of revisions that will be maintained in the StatefulSet's revision history. The revision history consists of all revisions not represented by a currently applied StatefulSetSpec version. The default value is 10.  

Default:  10


## PersistentVolumeClaim

PersistentVolumeClaim is a subset of [PersistentVolumeClaim in k8s.io/api/core/v1](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.19/#persistentvolumeclaim-v1-core).

### metadata (EmbeddedPersistentVolumeClaimObjectMeta, optional) {#persistentvolumeclaim-metadata}

Default: -

### spec (v1.PersistentVolumeClaimSpec, optional) {#persistentvolumeclaim-spec}

Default: -


## EmbeddedPersistentVolumeClaimObjectMeta

ObjectMeta contains only a subset of the fields included in k8s.io/apimachinery/pkg/apis/meta/v1.ObjectMeta
Only fields which are relevant to embedded PVCs are included
controller-gen discards embedded ObjectMetadata type fields, so we have to overcome this.

### name (string, optional) {#embeddedpersistentvolumeclaimobjectmeta-name}

Default: -

### annotations (map[string]string, optional) {#embeddedpersistentvolumeclaimobjectmeta-annotations}

Default: -

### labels (map[string]string, optional) {#embeddedpersistentvolumeclaimobjectmeta-labels}

Default: -


## PodTemplateSpec

PodTemplateSpec describes the data a pod should have when created from a template
It's the same as [PodTemplateSpec in k8s.io/api/core/v1](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.19/#podtemplatespec-v1-core) but with the [local ObjectMeta](#objectmeta) and [PodSpec](#podspec) types embedded.

### metadata (ObjectMeta, optional) {#podtemplatespec-metadata}

Default: -

### spec (PodSpec, optional) {#podtemplatespec-spec}

Default: -


## PodSpec

PodSpec is a subset of [PodSpec in k8s.io/api/corev1](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.19/#podspec-v1-core). It's the same as the original PodSpec expect it allows for containers to be missing.

### volumes ([]v1.Volume, optional) {#podspec-volumes}

List of volumes that can be mounted by containers belonging to the pod. +patchMergeKey=name +patchStrategy=merge,retainKeys 

Default: -

### initContainers ([]v1.Container, optional) {#podspec-initcontainers}

List of initialization containers belonging to the pod. +patchMergeKey=name +patchStrategy=merge 

Default: -

### containers ([]v1.Container, optional) {#podspec-containers}

List of containers belonging to the pod. +patchMergeKey=name +patchStrategy=merge 

Default: -

### ephemeralContainers ([]v1.EphemeralContainer, optional) {#podspec-ephemeralcontainers}

List of ephemeral containers run in this pod. +patchMergeKey=name +patchStrategy=merge 

Default: -

### restartPolicy (v1.RestartPolicy, optional) {#podspec-restartpolicy}

Restart policy for all containers within the pod. One of Always, OnFailure, Never. Default to Always.  

Default:  Always

### terminationGracePeriodSeconds (*int64, optional) {#podspec-terminationgraceperiodseconds}

Optional duration in seconds the pod needs to terminate gracefully. Defaults to 30 seconds.  

Default:  30

### activeDeadlineSeconds (*int64, optional) {#podspec-activedeadlineseconds}

Optional duration in seconds the pod may be active on the node relative to 

Default: -

### dnsPolicy (v1.DNSPolicy, optional) {#podspec-dnspolicy}

Set DNS policy for the pod. Defaults to "ClusterFirst".  Valid values are 'ClusterFirstWithHostNet', 'ClusterFirst', 'Default' or 'None'. 

Default:  ClusterFirst

### nodeSelector (map[string]string, optional) {#podspec-nodeselector}

NodeSelector is a selector which must be true for the pod to fit on a node. 

Default: -

### serviceAccountName (string, optional) {#podspec-serviceaccountname}

ServiceAccountName is the name of the ServiceAccount to use to run this pod. 

Default: -

### automountServiceAccountToken (*bool, optional) {#podspec-automountserviceaccounttoken}

AutomountServiceAccountToken indicates whether a service account token should be automatically mounted. 

Default: -

### nodeName (string, optional) {#podspec-nodename}

NodeName is a request to schedule this pod onto a specific node. 

Default: -

### hostNetwork (bool, optional) {#podspec-hostnetwork}

Host networking requested for this pod. Use the host's network namespace. If this option is set, the ports that will be used must be specified. Default to false.  

Default:  false

### hostPID (bool, optional) {#podspec-hostpid}

Use the host's pid namespace. Optional: Default to false.  

Default:  false

### hostIPC (bool, optional) {#podspec-hostipc}

Use the host's ipc namespace. Optional: Default to false.  

Default:  false

### shareProcessNamespace (*bool, optional) {#podspec-shareprocessnamespace}

Share a single process namespace between all of the containers in a pod. HostPID and ShareProcessNamespace cannot both be set. Optional: Default to false.  

Default:  false

### securityContext (*v1.PodSecurityContext, optional) {#podspec-securitycontext}

SecurityContext holds pod-level security attributes and common container settings. Optional: Defaults to empty.  See type description for default values of each field. 

Default: -

### imagePullSecrets ([]v1.LocalObjectReference, optional) {#podspec-imagepullsecrets}

ImagePullSecrets is an optional list of references to secrets in the same namespace to use for pulling any of the images used by this PodSpec. +patchMergeKey=name +patchStrategy=merge 

Default: -

### hostname (string, optional) {#podspec-hostname}

Specifies the hostname of the Pod If not specified, the pod's hostname will be set to a system-defined value. 

Default: -

### subdomain (string, optional) {#podspec-subdomain}

If specified, the fully qualified Pod hostname will be "<hostname>.<subdomain>.<pod namespace>.svc.<cluster domain>". If not specified, the pod will not have a domainname at all. 

Default: -

### affinity (*v1.Affinity, optional) {#podspec-affinity}

If specified, the pod's scheduling constraints 

Default: -

### schedulerName (string, optional) {#podspec-schedulername}

If specified, the pod will be dispatched by specified scheduler. If not specified, the pod will be dispatched by default scheduler. 

Default: -

### tolerations ([]v1.Toleration, optional) {#podspec-tolerations}

If specified, the pod's tolerations. 

Default: -

### hostAliases ([]v1.HostAlias, optional) {#podspec-hostaliases}

HostAliases is an optional list of hosts and IPs that will be injected into the pod's hosts file if specified. This is only valid for non-hostNetwork pods. +patchMergeKey=ip +patchStrategy=merge 

Default: -

### priorityClassName (string, optional) {#podspec-priorityclassname}

If specified, indicates the pod's priority. 

Default: -

### priority (*int32, optional) {#podspec-priority}

The priority value. Various system components use this field to find the priority of the pod. 

Default: -

### dnsConfig (*v1.PodDNSConfig, optional) {#podspec-dnsconfig}

Specifies the DNS parameters of a pod. 

Default: -

### readinessGates ([]v1.PodReadinessGate, optional) {#podspec-readinessgates}

If specified, all readiness gates will be evaluated for pod readiness. 

Default: -

### runtimeClassName (*string, optional) {#podspec-runtimeclassname}

RuntimeClassName refers to a RuntimeClass object in the node.k8s.io group, which should be used to run this pod. 

Default: -

### enableServiceLinks (*bool, optional) {#podspec-enableservicelinks}

EnableServiceLinks indicates whether information about services should be injected into pod's environment variables, matching the syntax of Docker links. Optional: Defaults to true.  

Default:  true

### preemptionPolicy (*v1.PreemptionPolicy, optional) {#podspec-preemptionpolicy}

PreemptionPolicy is the Policy for preempting pods with lower priority. One of Never, PreemptLowerPriority. Defaults to PreemptLowerPriority if unset.  

Default:  PreemptLowerPriority

### overhead (v1.ResourceList, optional) {#podspec-overhead}

Overhead represents the resource overhead associated with running a pod for a given RuntimeClass. 

Default: -

### topologySpreadConstraints ([]v1.TopologySpreadConstraint, optional) {#podspec-topologyspreadconstraints}

TopologySpreadConstraints describes how a group of pods ought to spread across topology domains. +patchMergeKey=topologyKey +patchStrategy=merge +listType=map +listMapKey=topologyKey +listMapKey=whenUnsatisfiable 

Default: -

### setHostnameAsFQDN (*bool, optional) {#podspec-sethostnameasfqdn}

If true the pod's hostname will be configured as the pod's FQDN, rather than the leaf name (the default). Default to false.  +optional 

Default:  false


## ServiceAccount

ServiceAccount is a subset of [ServiceAccount in k8s.io/api/core/v1](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.19/#serviceaccount-v1-core).

### metadata (ObjectMeta, optional) {#serviceaccount-metadata}

+optional 

Default: -

### secrets ([]v1.ObjectReference, optional) {#serviceaccount-secrets}

+optional +patchMergeKey=name +patchStrategy=merge 

Default: -

### imagePullSecrets ([]v1.LocalObjectReference, optional) {#serviceaccount-imagepullsecrets}

+optional 

Default: -

### automountServiceAccountToken (*bool, optional) {#serviceaccount-automountserviceaccounttoken}

+optional 

Default: -


