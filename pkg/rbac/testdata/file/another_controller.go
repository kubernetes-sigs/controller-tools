package controller

// +kubebuilder:rbac:groups=ocean,resources=jobs,verbs=get
// +kubebuilder:rbac:groups=land,resources=jobs,verbs=get,namespace=zoo
// +kubebuilder:rbac:groups=ocean,resources=jobs,verbs=get,namespace=zoo
// +kubebuilder:rbac:groups=ocean,resources=jobs,verbs=get,namespace=park
// +kubebuilder:rbac:groups=batch.io,resources=cronjobs,resourceNames=abc;def;xyz,verbs=get;watch
