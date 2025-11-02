package v1

// +kubebuilder:webhook:webhookVersions=v1,verbs=create;update,mutating=false,failurePolicy=fail,matchPolicy=Equivalent,groups=testdata.kubebuilder.io,resources=cronjobs,versions=v1,name=validation.cronjob.testdata.kubebuilder.io,sideEffects=None,admissionReviewVersions=v1;v1beta1,path=/validate-testdata-kubebuilder-io-v1-cronjob
