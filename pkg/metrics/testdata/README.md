# Testdata for generator tests

The files in this directory are used for testing the `kube-state-metrics generate` command and to provide an example.

## foo-config.yaml

This file is used in the test at [generate_integration_test.go](../generate_integration_test.go) to verify that the resulting configuration does not change during changes in the codebase.

If there are intended changes this file needs to get regenerated to make the test succeed again.
This could be done via:

```sh
go run ./cmd/controller-gen metrics crd \
    paths=./pkg/metrics/testdata \
    output:dir=./pkg/metrics/testdata
```

Or by using the go:generate marker inside [foo_types.go](foo_types.go):

```sh
go generate ./pkg/metrics/testdata/
```

## Example files: metrics.yaml, rbac.yaml and example-metrics.txt

There is also an example CR ([example-foo.yaml](example-foo.yaml)) and resulting example metrics ([example-metrics.txt](example-metrics.txt)).

The example metrics file got created by:

1. Generating a CustomResourceDefinition and Kube-State-Metrics configration file:

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
        --custom-resource-state-config-file /data/pkg/metrics/testdata/foo-config.yaml
    ```

5. Querying the metrics endpoint in a second terminal:

    ```sh
    curl localhost:8080/metrics > ./pkg/metrics/testdata/foo-cr-example-metrics.txt
    ```
