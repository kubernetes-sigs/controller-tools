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

// Validating webhook with namespaceSelector via patch marker (matchExpressions). Uses backticks so no escapes are needed.
// +kubebuilder:webhook:verbs=create;update,path=/validate-testdata-kubebuilder-io-v1-cronjob,mutating=false,failurePolicy=fail,groups=testdata.kubebuilder.io,resources=cronjobs,versions=v1,name=validation.cronjob.testdata.kubebuilder.io,sideEffects=None,admissionReviewVersions=v1,patch=`{"namespaceSelector":{"matchExpressions":[{"key":"environment","operator":"In","values":["dev","staging","prod"]}]}}`
// Mutating webhook with objectSelector via patch marker (matchLabels and matchExpressions). Uses backticks so no escapes are needed.
// +kubebuilder:webhook:verbs=create;update,path=/mutate-testdata-kubebuilder-io-v1-cronjob,mutating=true,failurePolicy=fail,groups=testdata.kubebuilder.io,resources=cronjobs,versions=v1,name=default.cronjob.testdata.kubebuilder.io,sideEffects=None,admissionReviewVersions=v1,patch=`{"objectSelector":{"matchLabels":{"managed-by":"controller"},"matchExpressions":[{"key":"tier","operator":"In","values":["frontend","backend"]}]}}`
