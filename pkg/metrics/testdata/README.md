# Testdata for generator tests

The files in this directory are used for testing the `controller-gen metrics` generator and to provide an example.

## metrics.yaml, rbac.yaml and bar.example.com_foos.yaml

These files are used in the test at [generate_integration_test.go](../generate_integration_test.go) to verify that the resulting output does not change during changes in the codebase.

If there are intended changes these files need to get regenerated to make the test succeed again.
This could be done by running [hack/update-generated.sh](../../../hack/update-generated.sh) from the root of the repository, which also runs the `go:generate` marker inside [foo_types.go](foo_types.go):

```sh
./hack/update-generated.sh
```

## Example files: metrics.yaml, rbac.yaml and example-metrics.txt

There is also an example CR ([example-foo.yaml](example-foo.yaml)) and resulting example metrics ([example-metrics.txt](example-metrics.txt)).

The example metrics file got created by:

1. Generating a CustomResourceDefinition and Kube-State-Metrics configuration file:

    ```sh
    ./hack/update-generated.sh
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
        --custom-resource-state-config-file /data/pkg/metrics/testdata/metrics.yaml
    ```

5. Querying the metrics endpoint in a second terminal:

    ```sh
    curl localhost:8080/metrics > ./pkg/metrics/testdata/example-metrics.txt
    ```
