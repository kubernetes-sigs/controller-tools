/*
Copyright The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

// +kubebuilder:webhook:webhookVersions=v1,verbs=create;update,mutating=false,failurePolicy=fail,matchPolicy=Equivalent,groups=testdata.kubebuilder.io,resources=cronjobs,versions=v1alpha1,name=a.cronjob.testdata.kubebuilder.io,sideEffects=None,admissionReviewVersions=v1;v1beta1,path=/validate-testdata-kubebuilder-io-v1alpha1-cronjob
