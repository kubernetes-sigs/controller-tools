package controller

//go:generate ../../../.run-controller-gen.sh rbac:roleName=manager-role paths=. output:dir=.

// +kubebuilder:rbac:groups=batch.io,resources=cronjobs,verbs=get;watch;create
// +kubebuilder:rbac:groups=batch.io,resources=cronjobs/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=art,resources=jobs,verbs=get
// +kubebuilder:rbac:groups=wave,resources=jobs,verbs=get,namespace=zoo
// +kubebuilder:rbac:groups=batch;batch;batch,resources=jobs/status,verbs=watch
// +kubebuilder:rbac:groups=batch;cron,resources=jobs/status,verbs=create;get
// +kubebuilder:rbac:groups=art,resources=jobs,verbs=get,namespace=zoo
// +kubebuilder:rbac:groups=cron;batch,resources=jobs/status,verbs=get;create
// +kubebuilder:rbac:groups=batch,resources=jobs/status,verbs=watch;watch
// +kubebuilder:rbac:groups=art,resources=jobs,verbs=get,namespace=park
// +kubebuilder:rbac:groups=batch.io,resources=cronjobs2,resourceNames=foo;bar;baz,verbs=get;watch
// ensure that "core" is translated to apiGroups="":
// +kubebuilder:rbac:groups=core,resources=configmaps,resourceNames=my-configmap,verbs=update;get
// check that the following merge:
// +kubebuilder:rbac:groups=test,resources=foo,resourceNames=bar,verbs=get
// +kubebuilder:rbac:groups=test,resources=foo,resourceNames=bar,verbs=watch
// make sure that differing URLs are not merged:
// +kubebuilder:rbac:groups=test,resources=diffurl,resourceNames=same,verbs=get,urls="https://example.com/1"
// +kubebuilder:rbac:groups=test,resources=diffurl,resourceNames=same,verbs=watch,urls="https://example.com/2"
// make sure that same URLs are merged:
// +kubebuilder:rbac:groups=test,resources=sameurl,resourceNames=same,verbs=get,urls="https://example.com/x"
// +kubebuilder:rbac:groups=test,resources=sameurl,resourceNames=same,verbs=watch,urls="https://example.com/x"
