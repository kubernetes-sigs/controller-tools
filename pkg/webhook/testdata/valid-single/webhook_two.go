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

func (c *CronJobTwo) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(c).
		Complete()
}

// +kubebuilder:webhook:webhookVersions=v1,verbs=create;update,path=/validate-testdata-kubebuilder-io-v1-cronjobtwo,mutating=false,failurePolicy=fail,matchPolicy=Equivalent,groups=testdata.kubebuiler.io,resources=cronjobtwos,versions=v1,name=validation.cronjobtwo.testdata.kubebuilder.io,sideEffects=None,timeoutSeconds=10,admissionReviewVersions=v1;v1beta1
// +kubebuilder:webhook:verbs=create;update,path=/validate-testdata-kubebuilder-io-v1-cronjobtwo,mutating=false,failurePolicy=fail,matchPolicy=Equivalent,groups=testdata.kubebuiler.io,resources=cronjobtwos,versions=v1,name=validation.cronjobtwo.testdata.kubebuilder.io,sideEffects=NoneOnDryRun,timeoutSeconds=10,admissionReviewVersions=v1;v1beta1
// +kubebuilder:webhook:webhookVersions=v1,verbs=create;update,path=/mutate-testdata-kubebuilder-io-v1-cronjobtwo,mutating=true,failurePolicy=fail,matchPolicy=Equivalent,groups=testdata.kubebuiler.io,resources=cronjobtwos,versions=v1,name=default.cronjobtwo.testdata.kubebuilder.io,sideEffects=None,timeoutSeconds=10,admissionReviewVersions=v1;v1beta1,reinvocationPolicy=IfNeeded

var _ webhook.Defaulter = &CronJobTwo{}
var _ webhook.Validator = &CronJobTwo{}

func (c *CronJobTwo) Default() {
}

func (c *CronJobTwo) ValidateCreate() error {
	return nil
}

func (c *CronJobTwo) ValidateUpdate(_ runtime.Object) error {
	return nil
}

func (c *CronJobTwo) ValidateDelete() error {
	return nil
}
