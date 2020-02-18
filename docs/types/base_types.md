### MetaBase
| Variable Name | Type | Required | Default | Description |
|---|---|---|---|---|
| annotations | map[string]string | No | - |  |
| labels | map[string]string | No | - |  |
### PodSpecBase
| Variable Name | Type | Required | Default | Description |
|---|---|---|---|---|
| tolerations | []corev1.Toleration | No | - |  |
| nodeSelector | map[string]string | No | - |  |
| serviceAccountName | string | No | - |  |
| affinity | *corev1.Affinity | No | - |  |
| securityContext | *corev1.PodSecurityContext | No | - |  |
| volumes | []corev1.Volume | No | - |  |
### ContainerBase
| Variable Name | Type | Required | Default | Description |
|---|---|---|---|---|
| resources | *corev1.ResourceRequirements | No | - |  |
| image | string | No | - |  |
| pullPolicy | corev1.PullPolicy | No | - |  |
| command | []string | No | - |  |
| volumeMounts | []corev1.VolumeMount | No | - |  |
| securityContext | *corev1.SecurityContext | No | - |  |
