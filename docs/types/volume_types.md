# Kubernetes volume abstraction
## Overview
 Refers to different types of volumes to be mounted to pods: emptyDir, hostPath, pvc

 Leverages core types from kubernetes/api/core/v1

## Configuration
## KubernetesVolume

### host_path (*corev1.HostPathVolumeSource, optional) {#kubernetesvolume-host_path}

Deprecated, use hostPath<br>

Default: -

### hostPath (*corev1.HostPathVolumeSource, optional) {#kubernetesvolume-hostpath}

Default: -

### emptyDir (*corev1.EmptyDirVolumeSource, optional) {#kubernetesvolume-emptydir}

Default: -

### pvc (*PersistentVolumeClaim, optional) {#kubernetesvolume-pvc}

PersistentVolumeClaim defines the Spec and the Source at the same time.<br>The PVC will be created with the configured spec and the name defined in the source.<br>

Default: -


## PersistentVolumeClaim

### spec (corev1.PersistentVolumeClaimSpec, optional) {#persistentvolumeclaim-spec}

Default: -

### source (corev1.PersistentVolumeClaimVolumeSource, optional) {#persistentvolumeclaim-source}

Default: -


