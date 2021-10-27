# CSIAddonsJob

CSIAddonsJob is a namespaced custom resource designed to invoke operation to be performed on a volume.

```yaml
apiVersion: csiaddons.openshift.io/v1beta1
kind: CSIAddonsJob
metadata:
  name: sample-1
spec:
  operation:
    reclaimSpace:
      target:
        pvc: pvc-1
      parameters:
        key1: value1
        key2: value2
  backOffLimit: 4
  activeDeadlineSeconds: 500
---
apiVersion: csiaddons.openshift.io/v1beta1
kind: CSIAddonsJob
metadata:
  name: sample-2
spec:
  operation:
    reclaimSpace:
      target:
        pv: pvc-xxxx-xxxx-xxxx
      parameters:
        key1: value1
        key2: value2
  backOffLimit: 4
  activeDeadlineSeconds: 500
```

+ `operation` contains the required information to invoke operation on target volume.
    + `reclaimSpace` contains the required information to invoke `reclaimSpace`  operation on target volume.
      + `target` represents the volume target on which the operation can be performed. Only one target needs to be specified.
        + `pvc` contains a string indicating the name of `PersistentVolumeClaim`.
        + `pv` contains a string indicating the name of `PersistentVolume`.
    + `parameters` contains key-value pairs that are passed down to the driver.
+ `backOfflimit` specifies the number of retries before marking this operation failed.
+ `activeDeadlineSeconds` specifies the duration in seconds relative to the startTime that the operation might be active; value must be positive integer.