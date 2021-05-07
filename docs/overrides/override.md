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

Kubernetes [Service Specification](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.19/#servicespec-v1-core)<br>

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

Kubernetes [Ingress Specification](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.19/#ingressclassspec-v1-networking-k8s-io)<br>

Default: -


## DaemonSet

DaemonSet is a subset of [DaemonSet in k8s.io/api/apps/v1](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.19/#daemonset-v1-apps), with [DaemonSetSpec replaced by the local variant](#daemonset-spec).

### metadata (ObjectMeta, optional) {#daemonset-metadata}

Default: -

### spec (DaemonSetSpec, optional) {#daemonset-spec}

[Local DaemonSet specification](#daemonset-spec)<br>

Default: -


## DaemonSetSpec

DaemonSetSpec is a subset of [DaemonSetSpec in k8s.io/api/apps/v1](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.19/#daemonsetspec-v1-apps) but with required fields declared as optional
and [PodTemplateSpec replaced by the local variant](#podtemplatespec).

### selector (*metav1.LabelSelector, optional) {#daemonsetspec-selector}

A label query over pods that are managed by the daemon set.<br>

Default: -

### template (PodTemplateSpec, optional) {#daemonsetspec-template}

An object that describes the pod that will be created. Note that this is a [local PodTemplateSpec](#podtemplatespec)<br>

Default: -

### updateStrategy (appsv1.DaemonSetUpdateStrategy, optional) {#daemonsetspec-updatestrategy}

An update strategy to replace existing DaemonSet pods with new pods.<br>

Default: -

### minReadySeconds (int32, optional) {#daemonsetspec-minreadyseconds}

The minimum number of seconds for which a newly created DaemonSet pod should<br>be ready without any of its container crashing, for it to be considered<br>available. Defaults to 0 (pod will be considered available as soon as it<br>is ready). <br>

Default:  0

### revisionHistoryLimit (*int32, optional) {#daemonsetspec-revisionhistorylimit}

The number of old history to retain to allow rollback.<br>This is a pointer to distinguish between explicit zero and not specified.<br>Defaults to 10. <br>

Default:  10


## Deployment

Deployment is a subset of [Deployment in k8s.io/api/apps/v1](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.19/#deployment-v1-apps), with [DeploymentSpec replaced by the local variant](#deployment-spec).

### metadata (ObjectMeta, optional) {#deployment-metadata}

Default: -

### spec (DeploymentSpec, optional) {#deployment-spec}

The desired behavior of [this deployment](#deploymentspec).<br>

Default: -


## DeploymentSpec

DeploymentSpec is a subset of [DeploymentSpec in k8s.io/api/apps/v1](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.19/#deploymentspec-v1-apps) but with required fields declared as optional
and [PodTemplateSpec replaced by the local variant](#podtemplatespec).

### replicas (*int32, optional) {#deploymentspec-replicas}

Number of desired pods. This is a pointer to distinguish between explicit<br>zero and not specified. Defaults to 1. <br>

Default:  1

### selector (*metav1.LabelSelector, optional) {#deploymentspec-selector}

Label selector for pods. Existing ReplicaSets whose pods are<br>selected by this will be the ones affected by this deployment.<br>It must match the pod template's labels.<br>

Default: -

### template (PodTemplateSpec, optional) {#deploymentspec-template}

An object that describes the pod that will be created. Note that this is a [local PodTemplateSpec](#podtemplatespec)<br>

Default: -

### strategy (appsv1.DeploymentStrategy, optional) {#deploymentspec-strategy}

The deployment strategy to use to replace existing pods with new ones.<br>+patchStrategy=retainKeys<br>

Default: -

### minReadySeconds (int32, optional) {#deploymentspec-minreadyseconds}

Minimum number of seconds for which a newly created pod should be ready<br>without any of its container crashing, for it to be considered available.<br>Defaults to 0 (pod will be considered available as soon as it is ready) <br>

Default:  0

### revisionHistoryLimit (*int32, optional) {#deploymentspec-revisionhistorylimit}

The number of old ReplicaSets to retain to allow rollback.<br>This is a pointer to distinguish between explicit zero and not specified.<br>Defaults to 10. <br>

Default:  10

### paused (bool, optional) {#deploymentspec-paused}

Indicates that the deployment is paused.<br>

Default: -

### progressDeadlineSeconds (*int32, optional) {#deploymentspec-progressdeadlineseconds}

The maximum time in seconds for a deployment to make progress before it<br>is considered to be failed. The deployment controller will continue to<br>process failed deployments and a condition with a ProgressDeadlineExceeded<br>reason will be surfaced in the deployment status. Note that progress will<br>not be estimated during the time a deployment is paused.<br>Defaults to 600s. <br>

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

Replicas is the desired number of replicas of the given Template.<br>These are replicas in the sense that they are instantiations of the<br>same Template, but individual replicas also have a consistent identity.<br>If unspecified, defaults to 1. <br>+optional<br>

Default:  1

### selector (*metav1.LabelSelector, optional) {#statefulsetspec-selector}

Selector is a label query over pods that should match the replica count.<br>It must match the pod template's labels.<br>

Default: -

### template (PodTemplateSpec, optional) {#statefulsetspec-template}

template is the object that describes the pod that will be created if<br>insufficient replicas are detected. Each pod stamped out by the StatefulSet<br>will fulfill this Template, but have a unique identity from the rest<br>of the StatefulSet.<br>

Default: -

### volumeClaimTemplates ([]PersistentVolumeClaim, optional) {#statefulsetspec-volumeclaimtemplates}

volumeClaimTemplates is a list of claims that pods are allowed to reference.<br>The StatefulSet controller is responsible for mapping network identities to<br>claims in a way that maintains the identity of a pod. Every claim in<br>this list must have at least one matching (by name) volumeMount in one<br>container in the template. A claim in this list takes precedence over<br>any volumes in the template, with the same name.<br>+optional<br>

Default: -

### serviceName (string, optional) {#statefulsetspec-servicename}

serviceName is the name of the service that governs this StatefulSet.<br>This service must exist before the StatefulSet, and is responsible for<br>the network identity of the set. Pods get DNS/hostnames that follow the<br>pattern: pod-specific-string.serviceName.default.svc.cluster.local<br>where "pod-specific-string" is managed by the StatefulSet controller.<br>

Default: -

### podManagementPolicy (appsv1.PodManagementPolicyType, optional) {#statefulsetspec-podmanagementpolicy}

podManagementPolicy controls how pods are created during initial scale up,<br>when replacing pods on nodes, or when scaling down. The default policy is<br>`OrderedReady`, where pods are created in increasing order (pod-0, then<br>pod-1, etc) and the controller will wait until each pod is ready before<br>continuing. When scaling down, the pods are removed in the opposite order.<br>The alternative policy is `Parallel` which will create pods in parallel<br>to match the desired scale without waiting, and on scale down will delete<br>all pods at once.<br><br>+optional<br>

Default:  OrderedReady

### updateStrategy (appsv1.StatefulSetUpdateStrategy, optional) {#statefulsetspec-updatestrategy}

updateStrategy indicates the StatefulSetUpdateStrategy that will be<br>employed to update Pods in the StatefulSet when a revision is made to<br>Template.<br>

Default: -

### revisionHistoryLimit (*int32, optional) {#statefulsetspec-revisionhistorylimit}

revisionHistoryLimit is the maximum number of revisions that will<br>be maintained in the StatefulSet's revision history. The revision history<br>consists of all revisions not represented by a currently applied<br>StatefulSetSpec version. The default value is 10. <br>

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

List of volumes that can be mounted by containers belonging to the pod.<br>+patchMergeKey=name<br>+patchStrategy=merge,retainKeys<br>

Default: -

### initContainers ([]v1.Container, optional) {#podspec-initcontainers}

List of initialization containers belonging to the pod.<br>+patchMergeKey=name<br>+patchStrategy=merge<br>

Default: -

### containers ([]v1.Container, optional) {#podspec-containers}

List of containers belonging to the pod.<br>+patchMergeKey=name<br>+patchStrategy=merge<br>

Default: -

### ephemeralContainers ([]v1.EphemeralContainer, optional) {#podspec-ephemeralcontainers}

List of ephemeral containers run in this pod.<br>+patchMergeKey=name<br>+patchStrategy=merge<br>

Default: -

### restartPolicy (v1.RestartPolicy, optional) {#podspec-restartpolicy}

Restart policy for all containers within the pod.<br>One of Always, OnFailure, Never.<br>Default to Always. <br>

Default:  Always

### terminationGracePeriodSeconds (*int64, optional) {#podspec-terminationgraceperiodseconds}

Optional duration in seconds the pod needs to terminate gracefully.<br>Defaults to 30 seconds. <br>

Default:  30

### activeDeadlineSeconds (*int64, optional) {#podspec-activedeadlineseconds}

Optional duration in seconds the pod may be active on the node relative to<br>

Default: -

### dnsPolicy (v1.DNSPolicy, optional) {#podspec-dnspolicy}

Set DNS policy for the pod.<br>Defaults to "ClusterFirst". <br>Valid values are 'ClusterFirstWithHostNet', 'ClusterFirst', 'Default' or 'None'.<br>

Default:  ClusterFirst

### nodeSelector (map[string]string, optional) {#podspec-nodeselector}

NodeSelector is a selector which must be true for the pod to fit on a node.<br>

Default: -

### serviceAccountName (string, optional) {#podspec-serviceaccountname}

ServiceAccountName is the name of the ServiceAccount to use to run this pod.<br>

Default: -

### automountServiceAccountToken (*bool, optional) {#podspec-automountserviceaccounttoken}

AutomountServiceAccountToken indicates whether a service account token should be automatically mounted.<br>

Default: -

### nodeName (string, optional) {#podspec-nodename}

NodeName is a request to schedule this pod onto a specific node.<br>

Default: -

### hostNetwork (bool, optional) {#podspec-hostnetwork}

Host networking requested for this pod. Use the host's network namespace.<br>If this option is set, the ports that will be used must be specified.<br>Default to false. <br>

Default:  false

### hostPID (bool, optional) {#podspec-hostpid}

Use the host's pid namespace.<br>Optional: Default to false. <br>

Default:  false

### hostIPC (bool, optional) {#podspec-hostipc}

Use the host's ipc namespace.<br>Optional: Default to false. <br>

Default:  false

### shareProcessNamespace (*bool, optional) {#podspec-shareprocessnamespace}

Share a single process namespace between all of the containers in a pod.<br>HostPID and ShareProcessNamespace cannot both be set.<br>Optional: Default to false. <br>

Default:  false

### securityContext (*v1.PodSecurityContext, optional) {#podspec-securitycontext}

SecurityContext holds pod-level security attributes and common container settings.<br>Optional: Defaults to empty.  See type description for default values of each field.<br>

Default: -

### imagePullSecrets ([]v1.LocalObjectReference, optional) {#podspec-imagepullsecrets}

ImagePullSecrets is an optional list of references to secrets in the same namespace to use for pulling any of the images used by this PodSpec.<br>+patchMergeKey=name<br>+patchStrategy=merge<br>

Default: -

### hostname (string, optional) {#podspec-hostname}

Specifies the hostname of the Pod<br>If not specified, the pod's hostname will be set to a system-defined value.<br>

Default: -

### subdomain (string, optional) {#podspec-subdomain}

If specified, the fully qualified Pod hostname will be "<hostname>.<subdomain>.<pod namespace>.svc.<cluster domain>".<br>If not specified, the pod will not have a domainname at all.<br>

Default: -

### affinity (*v1.Affinity, optional) {#podspec-affinity}

If specified, the pod's scheduling constraints<br>

Default: -

### schedulerName (string, optional) {#podspec-schedulername}

If specified, the pod will be dispatched by specified scheduler.<br>If not specified, the pod will be dispatched by default scheduler.<br>

Default: -

### tolerations ([]v1.Toleration, optional) {#podspec-tolerations}

If specified, the pod's tolerations.<br>

Default: -

### hostAliases ([]v1.HostAlias, optional) {#podspec-hostaliases}

HostAliases is an optional list of hosts and IPs that will be injected into the pod's hosts<br>file if specified. This is only valid for non-hostNetwork pods.<br>+patchMergeKey=ip<br>+patchStrategy=merge<br>

Default: -

### priorityClassName (string, optional) {#podspec-priorityclassname}

If specified, indicates the pod's priority.<br>

Default: -

### priority (*int32, optional) {#podspec-priority}

The priority value. Various system components use this field to find the<br>priority of the pod.<br>

Default: -

### dnsConfig (*v1.PodDNSConfig, optional) {#podspec-dnsconfig}

Specifies the DNS parameters of a pod.<br>

Default: -

### readinessGates ([]v1.PodReadinessGate, optional) {#podspec-readinessgates}

If specified, all readiness gates will be evaluated for pod readiness.<br>

Default: -

### runtimeClassName (*string, optional) {#podspec-runtimeclassname}

RuntimeClassName refers to a RuntimeClass object in the node.k8s.io group, which should be used<br>to run this pod.<br>

Default: -

### enableServiceLinks (*bool, optional) {#podspec-enableservicelinks}

EnableServiceLinks indicates whether information about services should be injected into pod's<br>environment variables, matching the syntax of Docker links.<br>Optional: Defaults to true. <br>

Default:  true

### preemptionPolicy (*v1.PreemptionPolicy, optional) {#podspec-preemptionpolicy}

PreemptionPolicy is the Policy for preempting pods with lower priority.<br>One of Never, PreemptLowerPriority.<br>Defaults to PreemptLowerPriority if unset. <br>

Default:  PreemptLowerPriority

### overhead (v1.ResourceList, optional) {#podspec-overhead}

Overhead represents the resource overhead associated with running a pod for a given RuntimeClass.<br>

Default: -

### topologySpreadConstraints ([]v1.TopologySpreadConstraint, optional) {#podspec-topologyspreadconstraints}

TopologySpreadConstraints describes how a group of pods ought to spread across topology<br>domains.<br>+patchMergeKey=topologyKey<br>+patchStrategy=merge<br>+listType=map<br>+listMapKey=topologyKey<br>+listMapKey=whenUnsatisfiable<br>

Default: -

### setHostnameAsFQDN (*bool, optional) {#podspec-sethostnameasfqdn}

If true the pod's hostname will be configured as the pod's FQDN, rather than the leaf name (the default).<br>Default to false. <br>+optional<br>

Default:  false


## ServiceAccount

ServiceAccount is a subset of [ServiceAccount in k8s.io/api/core/v1](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.19/#serviceaccount-v1-core).

### metadata (ObjectMeta, optional) {#serviceaccount-metadata}

+optional<br>

Default: -

### secrets ([]v1.ObjectReference, optional) {#serviceaccount-secrets}

+optional<br>+patchMergeKey=name<br>+patchStrategy=merge<br>

Default: -

### imagePullSecrets ([]v1.LocalObjectReference, optional) {#serviceaccount-imagepullsecrets}

+optional<br>

Default: -

### automountServiceAccountToken (*bool, optional) {#serviceaccount-automountserviceaccounttoken}

+optional<br>

Default: -


