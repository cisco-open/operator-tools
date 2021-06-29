# Secret abstraction
## Overview
 Provides an abstraction to refer to secrets from other types
 SecretLoader facilitates loading the secrets from an operator.
 Leverages core types from kubernetes/api/core/v1

## Configuration
## Secret

### value (string, optional) {#secret-value}

Refers to a non-secret value 

Default: -

### valueFrom (*ValueFrom, optional) {#secret-valuefrom}

Refers to a secret value to be used directly 

Default: -

### mountFrom (*ValueFrom, optional) {#secret-mountfrom}

Refers to a secret value to be used through a volume mount 

Default: -


## ValueFrom

### secretKeyRef (*corev1.SecretKeySelector, optional) {#valuefrom-secretkeyref}

Default: -


