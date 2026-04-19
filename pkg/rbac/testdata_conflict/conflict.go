package controller

// Multiple markers with same roleName but different ServiceAccounts - should fail
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get,serviceAccountName=sa-one,serviceAccountNamespace=ns-one
// +kubebuilder:rbac:groups=batch,resources=jobs,verbs=list,serviceAccountName=sa-two,serviceAccountNamespace=ns-two
