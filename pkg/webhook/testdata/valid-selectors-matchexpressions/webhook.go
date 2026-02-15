/*

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

package cronjob

import (
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

func (c *CronJob) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(c).
		Complete()
}

// Test with matchExpressions for namespaceSelector
// +kubebuilder:webhook:verbs=create;update,path=/validate-testdata-kubebuilder-io-v1-cronjob,mutating=false,failurePolicy=fail,groups=testdata.kubebuilder.io,resources=cronjobs,versions=v1,name=validation.cronjob.testdata.kubebuilder.io,sideEffects=None,admissionReviewVersions=v1,namespaceSelector=matchExpressions~key=environment.operator=In.values=dev|staging|prod
// Test with combined matchLabels and matchExpressions for objectSelector
// +kubebuilder:webhook:verbs=create;update,path=/mutate-testdata-kubebuilder-io-v1-cronjob,mutating=true,failurePolicy=fail,groups=testdata.kubebuilder.io,resources=cronjobs,versions=v1,name=default.cronjob.testdata.kubebuilder.io,sideEffects=None,admissionReviewVersions=v1,objectSelector=matchLabels~managed-by=controller&matchExpressions~key=tier.operator=In.values=frontend|backend

var _ webhook.Defaulter = &CronJob{}
var _ webhook.Validator = &CronJob{}

func (c *CronJob) Default() {
}

func (c *CronJob) ValidateCreate() error {
	return nil
}

func (c *CronJob) ValidateUpdate(_ runtime.Object) error {
	return nil
}

func (c *CronJob) ValidateDelete() error {
	return nil
}
