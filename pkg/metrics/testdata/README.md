# Testdata for generator tests

The files in this directory are used for testing the `controller-gen metrics` generator and to provide an example.

## Expected output

The generator creates the configuration for one target, which is selected with the `target` option. The default target is `resource-state-metrics`. The expected output for the types in this directory is in one directory per target:

* [resource-state-metrics/](resource-state-metrics/): a `ResourceMetricsMonitor` using the CEL resolver and a ClusterRole carrying the `resource-state-metrics/aggregate-to-manager` label.
* [kube-state-metrics/](kube-state-metrics/): the custom resource state configuration and a ClusterRole carrying the `kube-state-metrics/aggregate-to-manager` label.

Next to them [bar.example.com_foos.yaml](bar.example.com_foos.yaml) is the CustomResourceDefinition of the types.

These files are used in the test at [generate_integration_test.go](../generate_integration_test.go) to verify that the resulting output does not change during changes in the codebase.

If there are intended changes these files need to get regenerated to make the test succeed again, by running the `go:generate` markers in [generate.go](generate.go):

```sh
go generate ./pkg/metrics/testdata/
```

## Example files: example-foo.yaml and example-metrics.txt

There is also an example CR ([example-foo.yaml](example-foo.yaml)) and resulting example metrics ([example-metrics.txt](example-metrics.txt)).

The example metrics file got created by:

1. Generating a CustomResourceDefinition and Kube-State-Metrics configuration file:

    ```sh
    go generate ./pkg/metrics/testdata/
    ```

2. Creating a cluster using [kind](https://kind.sigs.k8s.io/)

    ```sh
    kind create cluster
    ```

3. Applying the CRD and example CR to the cluster:

    ```sh
    kubectl apply -f ./pkg/metrics/testdata/bar.example.com_foos.yaml
    kubectl apply -f ./pkg/metrics/testdata/example-foo.yaml
    ```

4. Running kube-state-metrics with the provided configuration file:

    ```sh
    docker run --net=host -ti --rm \
        -v $HOME/.kube/config:/config \
        -v $(pwd):/data \
        registry.k8s.io/kube-state-metrics/kube-state-metrics:v2.13.0 \
        --kubeconfig /config --custom-resource-state-only \
        --custom-resource-state-config-file /data/pkg/metrics/testdata/kube-state-metrics/metrics.yaml
    ```

5. Querying the metrics endpoint in a second terminal:

    ```sh
    curl localhost:8080/metrics > ./pkg/metrics/testdata/example-metrics.txt
    ```
