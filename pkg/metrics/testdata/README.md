# Testdata for generator tests

The files in this directory are used for testing the `controller-gen metrics` generator and to provide an example.

## Expected output

The generator creates the configuration for one target, which is selected with the `target` option. The default target is `resource-state-metrics`. The expected output for the types in this directory is in one directory per target:

* [resource-state-metrics/](resource-state-metrics/): a `ResourceMetricsMonitor` using the CEL resolver and a ClusterRole carrying the `resource-state-metrics.instrumentation.k8s-sigs.io/aggregate-to-exporter` label.
* [kube-state-metrics/](kube-state-metrics/): the custom resource state configuration and a ClusterRole carrying the `kube-state-metrics/aggregate-to-manager` label.

Next to them [bar.example.com_foos.yaml](bar.example.com_foos.yaml) is the CustomResourceDefinition of the types.

These files are used in the test at [generator_integration_test.go](../generator_integration_test.go) to verify that the resulting output does not change during changes in the codebase.

If there are intended changes these files need to get regenerated to make the test succeed again, by running the `go:generate` markers in [generate.go](generate.go):

```sh
go generate ./pkg/metrics/testdata/
```

## Example files: example-foo.yaml, example-metrics-ksm.txt and example-metrics-rsm.txt

There is also an example CR ([example-foo.yaml](example-foo.yaml)) and the resulting example metrics for each target:
[example-metrics-ksm.txt](example-metrics-ksm.txt) for kube-state-metrics and
[example-metrics-rsm.txt](example-metrics-rsm.txt) for resource-state-metrics.

### kube-state-metrics: example-metrics-ksm.txt

The example metrics file got created by:

1. Generating a CustomResourceDefinition and Kube-State-Metrics configuration file:

    ```sh
    go generate ./pkg/metrics/testdata/
    ```

2. Creating a cluster using [kind](https://kind.sigs.k8s.io/)

    ```sh
    kind create cluster
    ```

3. Applying the CRD and example CR to the cluster. The CRD has `v2` as storage version, which has no status, so the
   status of the example CR would be pruned. Make `v1` the storage version of the CRD (`kubectl edit crd foos.bar.example.com`)
   before applying the example CR:

    ```sh
    kubectl apply -f ./pkg/metrics/testdata/bar.example.com_foos.yaml
    kubectl apply -f ./pkg/metrics/testdata/example-foo.yaml
    ```

4. Running kube-state-metrics v2.20.0 (the version the generated configuration is based on) with the provided configuration file.
   The container has to reach the API server of the kind cluster, so it runs on the `kind` docker network with the internal kubeconfig.
   Host networking does not work with docker on macOS, and the files are copied into the container instead of mounted:

    ```sh
    kind get kubeconfig --internal > /tmp/kubeconfig-internal
    docker create --name ksm --network kind -p 8080:8080 \
        registry.k8s.io/kube-state-metrics/kube-state-metrics:v2.20.0 \
        --kubeconfig /tmp/config --custom-resource-state-only \
        --custom-resource-state-config-file /tmp/metrics.yaml
    docker cp /tmp/kubeconfig-internal ksm:/tmp/config
    docker cp ./pkg/metrics/testdata/kube-state-metrics/metrics.yaml ksm:/tmp/metrics.yaml
    docker start ksm
    ```

5. Querying the metrics endpoint in a second terminal:

    ```sh
    curl localhost:8080/metrics > ./pkg/metrics/testdata/example-metrics-ksm.txt
    ```

### resource-state-metrics: example-metrics-rsm.txt

The example metrics file got created by:

1. Generating the CustomResourceDefinition and the resource-state-metrics configuration:

    ```sh
    go generate ./pkg/metrics/testdata/
    ```

2. Creating a cluster using [kind](https://kind.sigs.k8s.io/)

    ```sh
    kind create cluster
    ```

3. Deploying the ServiceAccount, ClusterRoles, bindings and CRD from the `manifests` directory of
   [resource-state-metrics](https://github.com/kubernetes-sigs/resource-state-metrics) at commit `14c3b15` (`v0.1.0-66-g14c3b15`).
   Its deployment is not needed, as resource-state-metrics is run locally in step 5.

4. Applying the CRD, the generated ClusterRole, the example CR and the generated ResourceMetricsMonitor:

    ```sh
    kubectl apply -f ./pkg/metrics/testdata/bar.example.com_foos.yaml
    kubectl apply -f ./pkg/metrics/testdata/resource-state-metrics/rbac.yaml
    kubectl apply -f ./pkg/metrics/testdata/example-foo.yaml
    kubectl apply -f ./pkg/metrics/testdata/resource-state-metrics/resource-metrics-monitor.yaml
    ```

    The CRD has `v2` as storage version, which has no status, so the status of the example CR is pruned. Make `v1`
    the storage version of the CRD (`kubectl edit crd foos.bar.example.com`) and apply the example CR again.

5. Running resource-state-metrics against the cluster. The binary is built from the commit of step 3 (`make resource-state-metrics`).
   The release v0.1.0 is not used, as it does not keep the values of list metrics aligned with their labels (fixed by
   [#22](https://github.com/kubernetes-sigs/resource-state-metrics/pull/22)):

    ```sh
    ./resource-state-metrics -kubeconfig $HOME/.kube/config
    ```

6. Querying the metrics endpoint (the telemetry metrics are on port 9998) in a second terminal:

    ```sh
    curl localhost:9999/metrics > ./pkg/metrics/testdata/example-metrics-rsm.txt
    ```
