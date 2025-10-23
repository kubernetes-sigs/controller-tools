package v1alpha1

// +kubebuilder:webhook:webhookVersions=v1,verbs=create;update,mutating=false,failurePolicy=fail,matchPolicy=Equivalent,groups=testdata.kubebuilder.io,resources=cronjobs,versions=v1alpha1,name=a.cronjob.testdata.kubebuilder.io,sideEffects=None,admissionReviewVersions=v1;v1beta1,path=/validate-testdata-kubebuilder-io-v1alpha1-cronjob
