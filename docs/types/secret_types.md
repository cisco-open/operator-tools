# Secret abstraction
## Overview
 Provides an abstraction to refer to secrets from other types
 SecretLoader facilitates loading the secrets from an operator.
 Leverages core types from kubernetes/api/core/v1

## Configuration
### Secret
| Variable Name | Type | Required | Default | Description |
|---|---|---|---|---|
| value | string | No | - | Refers to a non-secret value<br> |
| valueFrom | *ValueFrom | No | - | Refers to a secret value to be used directly<br> |
| mountFrom | *ValueFrom | No | - | Refers to a secret value to be used through a volume mount<br> |
### ValueFrom
| Variable Name | Type | Required | Default | Description |
|---|---|---|---|---|
| secretKeyRef | *corev1.SecretKeySelector | No | - |  |
