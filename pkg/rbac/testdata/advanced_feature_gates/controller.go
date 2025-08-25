package testdata

// Always included RBAC rule
// +kubebuilder:rbac:groups="",resources=pods,verbs=get;list;watch

// Another always included rule
// +kubebuilder:rbac:groups="",resources=configmaps,verbs=get;list

// OR logic: alpha OR beta feature gate RBAC rule
// +kubebuilder:rbac:featureGate=alpha|beta,groups="",resources=secrets,verbs=get;list;create

// AND logic: alpha AND beta feature gate RBAC rule  
// +kubebuilder:rbac:featureGate=alpha&beta,groups="",resources=services,verbs=get;list;create;update;delete

func main() {
	// Test file for advanced RBAC feature gates
}
