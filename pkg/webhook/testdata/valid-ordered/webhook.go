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
	"context"
	"net/http"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

func (c *CronJob) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(c).
		Complete()
}

// +kubebuilder:webhook:webhookVersions=v1,verbs=create;update,path=/validate-testdata-kubebuilder-io-v1-cronjob,mutating=false,failurePolicy=fail,matchPolicy=Equivalent,groups=testdata.kubebuiler.io,resources=cronjobs,versions=v1,name=cronjob.testdata.kubebuilder.io,sideEffects=None,admissionReviewVersions=v1;v1beta1,reinvocationPolicy=Never

type cronjobValidator struct {
	client  client.Client
	decoder *admission.Decoder
}

func NewCronjobValidator(c client.Client, dec *admission.Decoder) http.Handler {
	return &webhook.Admission{Handler: &cronjobValidator{c, dec}}
}

func (v *cronjobValidator) Handle(ctx context.Context, req admission.Request) admission.Response {
	return admission.Allowed("ok")
}

// +kubebuilder:webhook:webhookVersions=v1,verbs=create;update,path=/validate-testdata-kubebuilder-io-v1-cronjoblist,mutating=false,failurePolicy=fail,matchPolicy=Equivalent,groups=testdata.kubebuiler.io,resources=cronjoblist,versions=v1,name=cronjoblist.testdata.kubebuilder.io,sideEffects=None,admissionReviewVersions=v1;v1beta1,reinvocationPolicy=Never

type cronjobListValidator struct {
	client  client.Client
	decoder *admission.Decoder
}

func NewCronjobListValidator(c client.Client, dec *admission.Decoder) http.Handler {
	return &webhook.Admission{Handler: &cronjobListValidator{c, dec}}
}

func (v *cronjobListValidator) Handle(ctx context.Context, req admission.Request) admission.Response {
	return admission.Allowed("ok")
}

// +kubebuilder:webhook:webhookVersions=v1,verbs=create;update,path=/validate-testdata-kubebuilder-io-v1-deployments,mutating=false,failurePolicy=fail,matchPolicy=Equivalent,groups=testdata.kubebuiler.io,resources=deployments,versions=v1,name=deployment.testdata.kubebuilder.io,sideEffects=None,admissionReviewVersions=v1;v1beta1,reinvocationPolicy=Never

type deploymentValidator struct {
	client  client.Client
	decoder *admission.Decoder
}

func NewDeploymentValidator(c client.Client, dec *admission.Decoder) http.Handler {
	return &webhook.Admission{Handler: &deploymentValidator{c, dec}}
}

func (v *deploymentValidator) Handle(ctx context.Context, req admission.Request) admission.Response {
	return admission.Allowed("ok")
}
