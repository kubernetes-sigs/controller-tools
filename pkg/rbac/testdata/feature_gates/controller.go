package testdata

// Always included RBAC rule
// +kubebuilder:rbac:groups="",resources=pods,verbs=get;list;watch

// Another always included rule
// +kubebuilder:rbac:groups="",resources=configmaps,verbs=get;list

// Alpha feature gate RBAC rule
// +kubebuilder:rbac:featureGate=alpha,groups=apps,resources=deployments,verbs=get;list;create;update;delete

// Beta feature gate RBAC rule  
// +kubebuilder:rbac:featureGate=beta,groups=extensions,resources=ingresses,verbs=get;list;create;update;delete

func main() {
	// Test file for RBAC feature gates
}
