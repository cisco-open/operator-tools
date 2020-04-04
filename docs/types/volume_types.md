# Kubernetes volume abstraction
## Overview
 Refers to different types of volumes to be mounted to pods: emptyDir, hostPath, pvc

 Leverages core types from kubernetes/api/core/v1

## Configuration
### KubernetesVolume
| Variable Name | Type | Required | Default | Description |
|---|---|---|---|---|
| host_path | *corev1.HostPathVolumeSource | No | - | Deprecated, use hostPath<br> |
| hostPath | *corev1.HostPathVolumeSource | No | - |  |
| emptyDir | *corev1.EmptyDirVolumeSource | No | - |  |
| pvc | *PersistentVolumeClaim | No | - | PersistentVolumeClaim defines the Spec and the Source at the same time.<br>The PVC will be created with the configured spec and the name defined in the source.<br> |
### PersistentVolumeClaim
| Variable Name | Type | Required | Default | Description |
|---|---|---|---|---|
| spec | corev1.PersistentVolumeClaimSpec | No | - |  |
| source | corev1.PersistentVolumeClaimVolumeSource | No | - |  |
