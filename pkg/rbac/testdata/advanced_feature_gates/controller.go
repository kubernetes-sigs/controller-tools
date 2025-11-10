package testdata

// Always included RBAC rule
// +kubebuilder:rbac:groups="",resources=pods,verbs=get;list;watch

// Another always included rule
// +kubebuilder:rbac:groups="",resources=configmaps,verbs=get;list

// OR logic: alpha OR beta feature gate RBAC rule
// +kubebuilder:rbac:featureGate=alpha|beta,groups="",resources=secrets,verbs=get;list;create

// AND logic: alpha AND beta feature gate RBAC rule
// +kubebuilder:rbac:featureGate=alpha&beta,groups="",resources=services,verbs=get;list;create;update;delete

// Complex precedence: (alpha AND beta) OR gamma
// +kubebuilder:rbac:featureGate=(alpha&beta)|gamma,groups=batch,resources=jobs,verbs=get;list;create

// Complex precedence: (alpha OR beta) AND gamma
// +kubebuilder:rbac:featureGate=(alpha|beta)&gamma,groups=apps,resources=replicasets,verbs=get;list;watch

func main() {
	// Test file for advanced RBAC feature gates with complex expressions
}
