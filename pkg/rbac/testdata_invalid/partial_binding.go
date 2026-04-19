package controller

// Only serviceAccountName specified (missing serviceAccountNamespace)
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get,serviceAccountName=controller-manager
