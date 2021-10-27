# CSIAddonsJob

CSIAddonsJob is a namespaced custom resource designed to invoke operation to be performed on a volume.

```yaml
apiVersion: csiaddons.openshift.io/v1beta1
kind: CSIAddonsJob
metadata:
  name: sample-1
spec:
  operation:
    name: "reclaimSpace"
    dataSource:
      name: pvc-1
      kind: PersistentVolumeClaim
      apiGroup: ""
    parameters:
      key1: value1
      key2: value2
  backOffLimit: 4
  activeDeadlineSeconds: 500
```

+ `operation` contains the required information to invoke operation on target volume.
    + `name` indicates the name of the operation that will be invoked by the CR.
    + `dataSource` contains typed reference to the volume on which the operation will be performed.
        + `apiGroup` is the group for the resource being referenced. If apiGroup is not specified, the specified Kind must
        be in the core API group. For any other third-party types, apiGroup is required.
        + `kind` is the kind of resource being replicated. For eg. `PersistentVolumeClaim` or `PersistentVolume`.
        + `name` is the name of the resource.
    + `parameters` contains key-value pairs that are passed down to the driver.
+ `backOfflimit` specifies the number of retries before marking this operation failed.
+ `activeDeadlineSeconds` specifies the duration in seconds relative to the startTime that the operation might be active; value must be positive integer.