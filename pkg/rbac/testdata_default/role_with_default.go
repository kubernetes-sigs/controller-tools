package controller

// Role with only serviceAccountName - should default serviceAccountNamespace to "myapp"
// +kubebuilder:rbac:groups=apps,namespace=myapp,resources=deployments,verbs=get,serviceAccountName=app-controller
