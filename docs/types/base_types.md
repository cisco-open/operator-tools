### ObjectKey
| Variable Name | Type | Required | Default | Description |
|---|---|---|---|---|
| name | string | No | - |  |
| namespace | string | No | - |  |
### MetaBase
| Variable Name | Type | Required | Default | Description |
|---|---|---|---|---|
| annotations | map[string]string | No | - |  |
| labels | map[string]string | No | - |  |
### PodTemplateBase
| Variable Name | Type | Required | Default | Description |
|---|---|---|---|---|
| metadata | *MetaBase | No | - |  |
| podSpec | *PodSpecBase | No | - |  |
### ContainerBase
| Variable Name | Type | Required | Default | Description |
|---|---|---|---|---|
| name | string | No | - |  |
| resources | *corev1.ResourceRequirements | No | - |  |
| image | string | No | - |  |
| pullPolicy | corev1.PullPolicy | No | - |  |
| command | []string | No | - |  |
| volumeMounts | []corev1.VolumeMount | No | - |  |
| securityContext | *corev1.SecurityContext | No | - |  |
| livenessProbe | *corev1.Probe | No | - |  |
| readinessProbe | *corev1.Probe | No | - |  |
### PodSpecBase
| Variable Name | Type | Required | Default | Description |
|---|---|---|---|---|
| tolerations | []corev1.Toleration | No | - |  |
| nodeSelector | map[string]string | No | - |  |
| serviceAccountName | string | No | - |  |
| affinity | *corev1.Affinity | No | - |  |
| securityContext | *corev1.PodSecurityContext | No | - |  |
| volumes | []corev1.Volume | No | - |  |
| priorityClassName | string | No | - |  |
| containers | []ContainerBase | No | - |  |
| imagePullSecrets | []corev1.LocalObjectReference | No | - |  |
### DeploymentSpecBase
| Variable Name | Type | Required | Default | Description |
|---|---|---|---|---|
| replicas | *int32 | No | - |  |
| selector | *metav1.LabelSelector | No | - |  |
| strategy | *appsv1.DeploymentStrategy | No | - |  |
### StatefulsetSpecBase
| Variable Name | Type | Required | Default | Description |
|---|---|---|---|---|
| replicas | *int32 | No | - |  |
| selector | *metav1.LabelSelector | No | - |  |
| podManagementPolicy | appsv1.PodManagementPolicyType | No | - |  |
| updateStrategy | *appsv1.StatefulSetUpdateStrategy | No | - |  |
### DaemonSetBase
| Variable Name | Type | Required | Default | Description |
|---|---|---|---|---|
|  | *MetaBase | Yes | - |  |
| spec | *DaemonSetSpecBase | No | - |  |
### DaemonSetSpecBase
| Variable Name | Type | Required | Default | Description |
|---|---|---|---|---|
| selector | *metav1.LabelSelector | No | - |  |
| updateStrategy | *appsv1.DaemonSetUpdateStrategy | No | - |  |
| minReadySeconds | int32 | No | - |  |
| revisionHistoryLimit | *int32 | No | - |  |
| template | *PodTemplateBase | No | - |  |
