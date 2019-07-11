package controller

// +kubebuilder:rbac:groups=batch.io,resources=cronjobs,verbs=get;watch;create
// +kubebuilder:rbac:groups=batch.io,resources=cronjobs/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=art,resources=jobs,verbs=get
// +kubebuilder:rbac:groups=batch;batch;batch,resources=jobs/status,verbs=watch
// +kubebuilder:rbac:groups=batch;cron,resources=jobs/status,verbs=create;get
// +kubebuilder:rbac:groups=cron;batch,resources=jobs/status,verbs=get;create
// +kubebuilder:rbac:groups=batch,resources=jobs/status,verbs=watch;watch
