# Metrics markers

The experimental `metrics` generator of controller-gen creates the configuration to expose metrics
for custom resources, using the markers described on this page. The configuration can be generated for
one of the following targets, which is selected with the `target` option:

* `resource-state-metrics` (default): a `ResourceMetricsMonitor` of
  [resource-state-metrics](https://github.com/kubernetes-sigs/resource-state-metrics) using the CEL resolver.
* `kube-state-metrics`: the custom resource state configuration of
  [kube-state-metrics](https://github.com/kubernetes/kube-state-metrics).

```sh
controller-gen metrics:experimental=true paths=./... output:dir=./config/metrics
controller-gen metrics:experimental=true,target=kube-state-metrics paths=./... output:dir=./config/metrics
```

Metrics are only generated for custom resources having the `store` marker. The `path`
and `value` fields use the JSON field names of the types and support simple paths only, e.g.
`.status.conditions`, without array notation. A path which does not exist in the Go types is an error.

The example metrics below are shown without the labels which the targets add themselves. The names of the metrics
are shown with the prefix of the resource-state-metrics target for a custom resource of the kind `Foo`.

## Markers

* [`k8s:controller-gen:metrics:store`](#k8scontroller-genmetricsstore)
* [`k8s:controller-gen:metrics:label`](#k8scontroller-genmetricslabel)
* [`k8s:controller-gen:metrics:gauge`](#k8scontroller-genmetricsgauge)
* [`k8s:controller-gen:metrics:info`](#k8scontroller-genmetricsinfo)
* [`k8s:controller-gen:metrics:stateset`](#k8scontroller-genmetricsstateset)

### `k8s:controller-gen:metrics:store`

Can be set on: type.

Enables the creation of a custom resource configuration entry and uses the given prefix for the metrics if configured.

| Field | Type | Required | Description |
|---|---|---|---|
| `namePrefix` | `string` | no | Specifies the prefix for all metrics of this resource.<br>Note: This field is only used by the kube-state-metrics target, where it directly maps to the metricNamePrefix field and replaces the default prefix kube_customresource, e.g. foo_created.<br>It is ignored by the resource-state-metrics target, which always uses the lowercased kind as prefix and adds it after the prefix kube_customresource_, e.g. kube_customresource_machinedeployment_created for the kind MachineDeployment. |

**Examples**

Minimal: enables the metrics generation for the custom resource. With the resource-state-metrics target the metrics get the lowercased kind as prefix, e.g. `kube_customresource_foo_created`.

```go
// +k8s:controller-gen:metrics:store
type Foo struct {
```

`namePrefix`: with the kube-state-metrics target all metrics get the prefix `foo` instead of `kube_customresource`, e.g. `foo_created`. It is ignored by the resource-state-metrics target.

```go
// +k8s:controller-gen:metrics:store:namePrefix="foo"
type Foo struct {
```

### `k8s:controller-gen:metrics:label`

Can be set on: type.

Specifies additional labels for all metrics of the custom resource.
It is only considered on the type of the custom resource, which is the type having the store marker.
Note: The label names group, version, kind, name and namespace can't be used, as the targets add labels with these names to every metric themselves. Using one of them is an error.

| Field | Type | Required | Description |
|---|---|---|---|
| `name` | `string` | yes | Specifies the name of the label. |
| `value` | `string` | yes | Specifies the JSONPath to the value of the label, relative to the custom resource. |

**Examples**

Adds the label `name` with the name of the custom resource to all metrics.

```go
// +k8s:controller-gen:metrics:store
// +k8s:controller-gen:metrics:label:name="name",value=".metadata.name"
type Foo struct {
```

A `value` with a map key containing dots, which have to be escaped: adds the label `app_name` with the value of the label `app.kubernetes.io/name` to all metrics.

```go
// +k8s:controller-gen:metrics:store
// +k8s:controller-gen:metrics:label:name="app_name",value=.metadata.labels.app\.kubernetes\.io/name
type Foo struct {
```

### `k8s:controller-gen:metrics:gauge`

Can be set on: field or type.

Defines a Gauge metric and uses the path to the field joined by the provided path as path of the metric.
Gauge is a metric which targets a Path that may be a single value, array, or object.
Lists and maps generate a metric per element. If the elements are objects, Value has to point to the numeric field of each element.
Ref: https://github.com/OpenObservability/OpenMetrics/blob/main/specification/OpenMetrics.md#gauge

| Field | Type | Required | Description |
|---|---|---|---|
| `name` | `string` | yes | Specifies the Name of the metric. |
| `help` | `string` | no | Specifies the help text for the metric. |
| `labels` | `map[string]string` | no | Specifies additional labels where the value is taken from the given JSONPath, relative to the metric path.<br>Note: The label names group, version, kind, name and namespace can't be used, as the targets add labels with these names to every metric themselves. Using one of them is an error.<br>Note: If the path points to a map, labels for its entries are not supported, as they can't be aligned with the values by all targets. |
| `path` | `string` | no | Specifies the JSONPath of the metric, relative to the field or type the marker is set on.<br>With the kube-state-metrics target the full path ends up in the path field of the metric. |
| `value` | `*string` | no | Specifies the JSONPath to a numeric field that will be the metric value, relative to the metric path.<br>Note: With the resource-state-metrics target, the value is converted to a unix timestamp if the field is a metav1.Time or metav1.MicroTime. |
| `missingAsZero` | `bool` | no | Specifies to expose a not-existing field as zero value instead of omitting the metric.<br>Note: With the resource-state-metrics target this also applies to the entries of lists and maps having no value, which are not left out then. |
| `ignoreMissing` | `bool` | no | Specifies to leave the metric out silently if a field on its path doesn't exist, e.g. because the status is not set yet.<br>Note: Only used by the resource-state-metrics target. Without it the evaluation of the expression fails for such a resource, which resource-state-metrics logs and counts as an error. |

**Examples**

Minimal gauge on a field (`name`, `help`). Without `path`, the metric uses the field the marker is set on.

```go
type FooSpec struct {
	// +k8s:controller-gen:metrics:gauge:name="replicas",help="The desired number of replicas."
	Replicas *int32 `json:"replicas,omitempty"`
}
```

Resulting metric for `replicas: 3`:

```
# HELP kube_customresource_foo_replicas The desired number of replicas.
# TYPE kube_customresource_foo_replicas gauge
kube_customresource_foo_replicas 3
```

`path`: the marker on the type of the custom resource with a path relative to the custom resource. A `metav1.Time` is exposed as unix timestamp.

```go
// +k8s:controller-gen:metrics:store
// +k8s:controller-gen:metrics:gauge:name="created",help="Unix creation timestamp.",path=".metadata.creationTimestamp"
type Foo struct {
```

Resulting metric:

```
# HELP kube_customresource_foo_created Unix creation timestamp.
# TYPE kube_customresource_foo_created gauge
kube_customresource_foo_created 1.724940463e+09
```

`value`: the metric value is taken from a field of the object the path points to.

```go
type FooSpec struct {
	// +k8s:controller-gen:metrics:gauge:name="settings_replicas",help="The replicas of the settings.",value=".replicas"
	Settings Settings `json:"settings"`
}
```

Resulting metric for `settings: {replicas: 2}`:

```
# HELP kube_customresource_foo_settings_replicas The replicas of the settings.
# TYPE kube_customresource_foo_settings_replicas gauge
kube_customresource_foo_settings_replicas 2
```

`labels`: a metric is created for each element of a list. The path of the labels is relative to each element. As the elements are objects, `value` selects the numeric field.

```go
type FooStatus struct {
	// +k8s:controller-gen:metrics:gauge:name="status_condition_last_transition_time",help="The condition last transition time of a foo.",value=".lastTransitionTime",labels={"type":".type","status":".status"}
	Conditions []Condition `json:"conditions,omitempty"`
}
```

Resulting metrics for the conditions `SomeType` (status `True`) and `AnotherType` (status `False`):

```
# HELP kube_customresource_foo_status_condition_last_transition_time The condition last transition time of a foo.
# TYPE kube_customresource_foo_status_condition_last_transition_time gauge
kube_customresource_foo_status_condition_last_transition_time{status="False",type="AnotherType"} 1.697119142e+09
kube_customresource_foo_status_condition_last_transition_time{status="True",type="SomeType"} 1.697119142e+09
```

`missingAsZero`: an unset field is exposed as 0 instead of omitting the metric.

```go
type FooStatus struct {
	// +k8s:controller-gen:metrics:gauge:name="ready_replicas",help="The number of ready replicas.",missingAsZero=true
	ReadyReplicas *int32 `json:"readyReplicas,omitempty"`
}
```

Resulting metric if `readyReplicas` is not set:

```
# HELP kube_customresource_foo_ready_replicas The number of ready replicas.
# TYPE kube_customresource_foo_ready_replicas gauge
kube_customresource_foo_ready_replicas 0
```

`ignoreMissing` (resource-state-metrics target only): the generated expressions check that the fields on the path exist, instead of failing the evaluation of the expression.

```go
type FooSpec struct {
	// +k8s:controller-gen:metrics:gauge:name="replicas",help="The desired number of replicas.",ignoreMissing=true
	Replicas *int32 `json:"replicas,omitempty"`
}
```

Without `ignoreMissing` the expression is `o.spec.replicas`. Evaluating it fails if `spec` or `replicas` is not set, so resource-state-metrics doesn't expose the metric, logs the failure at verbosity 1 and counts the evaluation with the result `error` in its counter of expression evaluations. With `ignoreMissing` the expression is `has(o.spec) && has(o.spec.replicas) ? [o.spec.replicas] : []`.
The option is also available for the `info` and `stateset` markers.

The marker on a type other than the custom resource. The path is relative to the type and the metric is generated for the field of the custom resource using the type, here `.spec.settings.replicas`.

```go
// +k8s:controller-gen:metrics:gauge:name="settings_replicas",help="The replicas of the settings.",path=".replicas"
type Settings struct {
	Replicas int32 `json:"replicas"`
}

type FooSpec struct {
	Settings Settings `json:"settings"`
}
```

### `k8s:controller-gen:metrics:info`

Can be set on: field or type.

Defines a Info metric and uses the implicit path to the field as path for the metric configuration.
Info is a metric which is used to expose textual information.
Ref: https://github.com/OpenObservability/OpenMetrics/blob/main/specification/OpenMetrics.md#info

| Field | Type | Required | Description |
|---|---|---|---|
| `name` | `string` | yes | Specifies the Name of the metric. |
| `help` | `string` | no | Specifies the help text for the metric. |
| `labels` | `map[string]string` | no | Specifies additional labels where the value is taken from the given JSONPath, relative to the metric path.<br>Note: The label names group, version, kind, name and namespace can't be used, as the targets add labels with these names to every metric themselves. Using one of them is an error.<br>Note: If the path points to a map, at most one label for its entries is supported, as the labels can't be aligned with each other by all targets. |
| `path` | `string` | no | Specifies the JSONPath of the metric, relative to the field or type the marker is set on.<br>With the kube-state-metrics target the full path ends up in the path field of the metric. |
| `ignoreMissing` | `bool` | no | Specifies to leave the metric out silently if a field on its path doesn't exist, e.g. because the status is not set yet.<br>Note: Only used by the resource-state-metrics target. Without it the evaluation of the expression fails for such a resource, which resource-state-metrics logs and counts as an error. |

**Examples**

`labels`: an info metric with the value 1 is created for each element of a list, the information is exposed by the labels. The path `.` of a label refers to the element itself.

```go
type FooSpec struct {
	// +k8s:controller-gen:metrics:info:name="tags",help="The tags of a foo.",labels={tag:"."}
	Tags []string `json:"tags,omitempty"`
}
```

Resulting metrics for `tags: [a, b]`:

```
# HELP kube_customresource_foo_tags The tags of a foo.
# TYPE kube_customresource_foo_tags info
kube_customresource_foo_tags{tag="a"} 1
kube_customresource_foo_tags{tag="b"} 1
```

`path`: the marker on the type of the custom resource with a path relative to the custom resource. The path of the labels is relative to each owner reference.

```go
// +k8s:controller-gen:metrics:store
// +k8s:controller-gen:metrics:info:name="owner",help="Owner references.",path=".metadata.ownerReferences",labels={owner_kind:".kind",owner_name:".name"}
type Foo struct {
```

Resulting metric for an owner reference:

```
# HELP kube_customresource_foo_owner Owner references.
# TYPE kube_customresource_foo_owner info
kube_customresource_foo_owner{owner_kind="foo",owner_name="foo"} 1
```

### `k8s:controller-gen:metrics:stateset`

Can be set on: field or type.

Defines a StateSet metric and uses the implicit path to the field as path for the metric configuration.
A StateSet is a metric which represent a series of related boolean values, also called a bitset.
Ref: https://github.com/OpenObservability/OpenMetrics/blob/main/specification/OpenMetrics.md#stateset

| Field | Type | Required | Description |
|---|---|---|---|
| `name` | `string` | yes | Specifies the Name of the metric. |
| `help` | `string` | no | Specifies the help text for the metric. |
| `labels` | `map[string]string` | no | Specifies additional labels where the value is taken from the given JSONPath, relative to the metric path.<br>Note: The label names group, version, kind, name and namespace can't be used, as the targets add labels with these names to every metric themselves. Using one of them is an error.<br>Note: If the path points to a map, labels for its entries are not supported, as they can't be aligned with the values by all targets. |
| `path` | `string` | no | Specifies the JSONPath of the metric, relative to the field or type the marker is set on.<br>With the kube-state-metrics target the full path ends up in the path field of the metric. |
| `states` | `[]string` | yes | Specifies a list of values to compare the given Value against. |
| `labelName` | `string` | no | Specifies the key of the label which is used for each entry in States to expose the value.<br>Note: It is required by the resource-state-metrics target. |
| `ignoreMissing` | `bool` | no | Specifies to leave the metric out silently if a field on its path doesn't exist, e.g. because the status is not set yet.<br>Note: Only used by the resource-state-metrics target. Without it the evaluation of the expression fails for such a resource, which resource-state-metrics logs and counts as an error. |
| `value` | `*string` | no | Specifies the JSONPath to the field which gets used as value to compare against the states for equality, relative to the metric path. |

**Examples**

Minimal stateset on a string field (`name`, `help`, `states`, `labelName`). A metric is created for each state, which has the value 1 if the field equals the state, otherwise 0. The state is exposed by the label `labelName`.

```go
type FooStatus struct {
	// +k8s:controller-gen:metrics:stateset:name="phase",help="The phase of a foo.",states={"Pending","Running","Failed"},labelName="phase"
	Phase string `json:"phase,omitempty"`
}
```

Resulting metrics for `phase: Running`:

```
# HELP kube_customresource_foo_phase The phase of a foo.
# TYPE kube_customresource_foo_phase stateset
kube_customresource_foo_phase{phase="Failed"} 0
kube_customresource_foo_phase{phase="Pending"} 0
kube_customresource_foo_phase{phase="Running"} 1
```

`value`: the field compared to the states is taken from the object the path points to.

```go
type FooStatus struct {
	// +k8s:controller-gen:metrics:stateset:name="rollout_phase",help="The phase of the rollout.",states={"Progressing","Complete"},labelName="phase",value=".phase"
	Rollout RolloutStatus `json:"rollout,omitempty"`
}
```

Resulting metrics for `rollout: {phase: Complete}`:

```
# HELP kube_customresource_foo_rollout_phase The phase of the rollout.
# TYPE kube_customresource_foo_rollout_phase stateset
kube_customresource_foo_rollout_phase{phase="Complete"} 1
kube_customresource_foo_rollout_phase{phase="Progressing"} 0
```

`labels`: the metrics are created for each element of a list and each state. The path of the labels is relative to each element. As the elements are objects, `value` selects the field compared to the states.

```go
type FooStatus struct {
	// +k8s:controller-gen:metrics:stateset:name="status_condition",help="The condition of a foo.",states={"True","False","Unknown"},labelName="status",value=".status",labels={"type":".type"}
	Conditions []Condition `json:"conditions,omitempty"`
}
```

Resulting metrics for the conditions `SomeType` (status `True`) and `AnotherType` (status `False`):

```
# HELP kube_customresource_foo_status_condition The condition of a foo.
# TYPE kube_customresource_foo_status_condition stateset
kube_customresource_foo_status_condition{status="False",type="AnotherType"} 1
kube_customresource_foo_status_condition{status="False",type="SomeType"} 0
kube_customresource_foo_status_condition{status="True",type="AnotherType"} 0
kube_customresource_foo_status_condition{status="True",type="SomeType"} 1
kube_customresource_foo_status_condition{status="Unknown",type="AnotherType"} 0
kube_customresource_foo_status_condition{status="Unknown",type="SomeType"} 0
```

## Restrictions and side cases

* **Paths** have to exist in the Go types, otherwise the generation fails. Paths below fields of the types
  `any`, `runtime.RawExtension` and `apiextensionsv1.JSON` can't be used, as their content is unknown.
* **Lists and maps** generate a metric per element. With the resource-state-metrics target, elements without
  the field of `value` are left out, unless `missingAsZero` is set.
* **Missing fields**: with the resource-state-metrics target the generated expressions don't check if the fields
  exist, unless `ignoreMissing` is set on the marker. Use `ignoreMissing` for metrics on optional fields, otherwise
  every evaluation for a resource without the field is counted as `error` by resource-state-metrics.
  A `label` of an optional field without `ignoreMissing` is also exposed with the text of its expression as value.
  The labels of the `label` marker are always checked. The kube-state-metrics target ignores the option.
* **Labels on maps**: as the targets can't align the labels of the entries of a map with each other and with
  the value, gauge and stateset metrics on a map don't support `labels`, and info metrics on a map support
  at most one label. The key of a map entry can't be exposed.
* **Timestamps**: with the resource-state-metrics target, values of the types `metav1.Time` and
  `metav1.MicroTime` are converted to unix timestamps.
* **Labels added by the targets**: resource-state-metrics adds the labels `group`, `version`, `kind`, `name` and
  `namespace` to every metric, kube-state-metrics adds `name` (generated as `labelsFromPath` of the resource).
  To have the same markers work for all targets, using one of the five names as name of a label of the `label`
  marker, of `labels` or as `labelName` is an error for every target.
* **Metric names**: with the resource-state-metrics target the name of the metric including the prefix has to
  match `^[a-zA-Z_][a-zA-Z0-9_]*$`. The label names have to be unique.
* **The `label` marker** is only considered on the type of the custom resource. Defining the same label
  twice is only allowed with the same `value`.
