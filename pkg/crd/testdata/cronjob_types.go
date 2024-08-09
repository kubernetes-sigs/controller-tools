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

// TODO(directxman12): test this across both versions (right now we're just
// trusting k/k conversion, which is probably fine though)

//go:generate ../../../.run-controller-gen.sh crd:ignoreUnexportedFields=true,allowDangerousTypes=true paths=./;./deprecated;./unserved;./job/... output:dir=.

// +groupName=testdata.kubebuilder.io
// +versionName=v1
package cronjob

import (
	"encoding/json"
	"fmt"
	"net/url"
	"time"

	batchv1beta1 "k8s.io/api/batch/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

const DefaultRefValue = "defaultRefValue"

// CronJobSpec defines the desired state of CronJob
// +kubebuilder:validation:XValidation:rule="has(oldSelf.forbiddenInt) || !has(self.forbiddenInt)",message="forbiddenInt is not allowed",fieldPath=".forbiddenInt",reason="FieldValueForbidden"
type CronJobSpec struct {
	// The schedule in Cron format, see https://en.wikipedia.org/wiki/Cron.
	Schedule string `json:"schedule"`

	// Optional deadline in seconds for starting the job if it misses scheduled
	// time for any reason.  Missed jobs executions will be counted as failed ones.
	// +optional
	StartingDeadlineSeconds *int64 `json:"startingDeadlineSeconds,omitempty"`

	// Specifies how to treat concurrent executions of a Job.
	// Valid values are:
	// - "Allow" (default): allows CronJobs to run concurrently;
	// - "Forbid": forbids concurrent runs, skipping next run if previous run hasn't finished yet;
	// - "Replace": cancels currently running job and replaces it with a new one
	// +optional
	ConcurrencyPolicy ConcurrencyPolicy `json:"concurrencyPolicy,omitempty"`

	// This flag tells the controller to suspend subsequent executions, it does
	// not apply to already started executions.  Defaults to false.
	// +optional
	Suspend *bool `json:"suspend,omitempty"`

	// This tests that non-serialized fields aren't included in the schema.
	InternalData string `json:"-"`

	// This flag is like suspend, but for when you really mean it.
	// It helps test the +kubebuilder:validation:Type marker.
	// +optional
	NoReallySuspend *TotallyABool `json:"noReallySuspend,omitempty"`

	// This tests byte slice schema generation.
	BinaryName []byte `json:"binaryName"`

	// This tests that nullable works correctly
	// +nullable
	CanBeNull string `json:"canBeNull"`

	// Specifies the job that will be created when executing a CronJob.
	JobTemplate batchv1beta1.JobTemplateSpec `json:"jobTemplate"`

	// The number of successful finished jobs to retain.
	// This is a pointer to distinguish between explicit zero and not specified.
	// +optional
	SuccessfulJobsHistoryLimit *int32 `json:"successfulJobsHistoryLimit,omitempty"`

	// The number of failed finished jobs to retain.
	// This is a pointer to distinguish between explicit zero and not specified.
	// +optional
	FailedJobsHistoryLimit *int32 `json:"failedJobsHistoryLimit,omitempty"`

	// This tests byte slices are allowed as map values.
	ByteSliceData map[string][]byte `json:"byteSliceData,omitempty"`

	// This tests string slices are allowed as map values.
	StringSliceData map[string][]string `json:"stringSliceData,omitempty"`

	// This tests pointers are allowed as map values.
	PtrData map[string]*string `json:"ptrData,omitempty"`

	// This tests that markers that are allowed on both fields and types are applied to fields
	// +kubebuilder:validation:MinLength=4
	TwoOfAKindPart0 string `json:"twoOfAKindPart0"`

	// This tests that markers that are allowed on both fields and types are applied to types
	TwoOfAKindPart1 LongerString `json:"twoOfAKindPart1"`

	// This tests that primitive defaulting can be performed.
	// +kubebuilder:default=forty-two
	// +kubebuilder:example=forty-two
	DefaultedString string `json:"defaultedString"`

	// This tests that slice defaulting can be performed.
	// +kubebuilder:default={a,b}
	// +kubebuilder:example={a,b}
	DefaultedSlice []string `json:"defaultedSlice"`

	// This tests that slice and object defaulting can be performed.
	// +kubebuilder:default={{nested: {foo: "baz", bar: true}},{nested: {foo: "qux", bar: false}}}
	// +kubebuilder:example={{nested: {foo: "baz", bar: true}},{nested: {foo: "qux", bar: false}}}
	DefaultedObject []RootObject `json:"defaultedObject"`

	// This tests that empty slice defaulting can be performed.
	// +kubebuilder:default={}
	DefaultedEmptySlice []string `json:"defaultedEmptySlice"`

	// This tests that an empty object defaulting can be performed on a map.
	// +kubebuilder:default={}
	DefaultedEmptyMap map[string]string `json:"defaultedEmptyMap"`

	// This tests that an empty object defaulting can be performed on an object.
	// +kubebuilder:default={}
	DefaultedEmptyObject EmpiableObject `json:"defaultedEmptyObject"`

	// This tests that kubebuilder defaulting takes precedence.
	// +kubebuilder:default="kubebuilder-default"
	// +default="kubernetes-default"
	DoubleDefaultedString string `json:"doubleDefaultedString"`

	// This tests that primitive defaulting can be performed.
	// +default="forty-two"
	KubernetesDefaultedString string `json:"kubernetesDefaultedString"`

	// This tests that slice defaulting can be performed.
	// +default=["a","b"]
	KubernetesDefaultedSlice []string `json:"kubernetesDefaultedSlice"`

	// This tests that slice and object defaulting can be performed.
	// +default=[{"nested": {"foo": "baz", "bar": true}},{"nested": {"foo": "qux", "bar": false}}]
	KubernetesDefaultedObject []RootObject `json:"kubernetesDefaultedObject"`

	// This tests that empty slice defaulting can be performed.
	// +default=[]
	KubernetesDefaultedEmptySlice []string `json:"kubernetesDefaultedEmptySlice"`

	// This tests that an empty object defaulting can be performed on a map.
	// +default={}
	KubernetesDefaultedEmptyMap map[string]string `json:"kubernetesDefaultedEmptyMap"`

	// This tests that an empty object defaulting can be performed on an object.
	// +default={}
	KubernetesDefaultedEmptyObject EmpiableObject `json:"kubernetesDefaultedEmptyObject"`

	// This tests that use of +default=ref(...) doesn't break generation
	// +default=ref(DefaultRefValue)
	KubernetesDefaultedRef string `json:"kubernetesDefaultedRef,omitempty"`

	// This tests that pattern validator is properly applied.
	// +kubebuilder:validation:Pattern=`^$|^((https):\/\/?)[^\s()<>]+(?:\([\w\d]+\)|([^[:punct:]\s]|\/?))$`
	PatternObject string `json:"patternObject"`

	// +kubebuilder:validation:EmbeddedResource
	// +kubebuilder:validation:nullable
	EmbeddedResource runtime.RawExtension `json:"embeddedResource"`

	// +kubebuilder:validation:nullable
	// +kubebuilder:pruning:PreserveUnknownFields
	UnprunedJSON NestedObject `json:"unprunedJSON"`

	// +kubebuilder:pruning:PreserveUnknownFields
	// +kubebuilder:validation:EmbeddedResource
	// +kubebuilder:validation:nullable
	UnprunedEmbeddedResource runtime.RawExtension `json:"unprunedEmbeddedResource"`

	// This tests that a type-level pruning marker works.
	UnprunedFromType Preserved `json:"unprunedFomType"`

	// This tests that a type-level pruning marker combined with a field-level pruning marker works.
	// +kubebuilder:pruning:PreserveUnknownFields
	UnprunedFromTypeAndField Preserved `json:"unprunedFomTypeAndField"`

	// This tests that associative lists work.
	// +listType=map
	// +listMapKey=name
	// +listMapKey=secondary
	AssociativeList []AssociativeType `json:"associativeList"`

	// This tests that associative lists work via a nested type.
	NestedAssociativeList NestedAssociativeList `json:"nestedassociativeList"`

	// A map that allows different actors to manage different fields
	// +mapType=granular
	MapOfInfo map[string][]byte `json:"mapOfInfo"`

	// A map that allows different actors to manage different fields via a nested type.
	NestedMapOfInfo NestedMapOfInfo `json:"nestedMapOfInfo"`

	// A struct that can only be entirely replaced
	// +structType=atomic
	StructWithSeveralFields NestedObject `json:"structWithSeveralFields"`

	// A struct that can only be entirely replaced via a nested type.
	NestedStructWithSeveralFields NestedStructWithSeveralFields `json:"nestedStructWithSeveralFields"`

	// A struct that can only be entirely replaced via a nested type and
	// field markers.
	// +structType=atomic
	NestedStructWithSeveralFieldsDoubleMarked NestedStructWithSeveralFields `json:"nestedStructWithSeveralFieldsDoubleMarked"`

	// This tests that type references are properly flattened
	// +kubebuilder:validation:optional
	JustNestedObject *JustNestedObject `json:"justNestedObject,omitempty"`

	// This tests explicitly optional kubebuilder fields
	// +kubebuilder:validation:Optional
	ExplicitlyOptionalKubebuilder string `json:"explicitlyOptionalKubebuilder"`

	// This tests explicitly optional kubernetes fields
	// +optional
	ExplicitlyOptionalKubernetes string `json:"explicitlyOptionalKubernetes"`

	// This tests explicitly required kubebuilder fields
	// +kubebuilder:validation:Required
	ExplicitlyRequiredKubebuilder string `json:"explicitlyRequiredKubebuilder,omitempty"`

	// This tests explicitly required kubernetes fields
	// +required
	ExplicitlyRequiredKubernetes string `json:"explicitlyRequiredKubernetes,omitempty"`

	// This tests that min/max properties work
	MinMaxProperties MinMaxObject `json:"minMaxProperties,omitempty"`

	// This tests that the schemaless marker works
	// +kubebuilder:validation:Schemaless
	Schemaless []byte `json:"schemaless,omitempty"`

	// This tests that an IntOrString can also have a pattern attached
	// to it.
	// This can be useful if you want to limit the string to a percentage or integer.
	// The XIntOrString marker is a requirement for having a pattern on this type.
	// +kubebuilder:validation:XIntOrString
	// +kubebuilder:validation:Pattern="^((100|[0-9]{1,2})%|[0-9]+)$"
	IntOrStringWithAPattern *intstr.IntOrString `json:"intOrStringWithAPattern,omitempty"`

	// Checks that nested maps work
	NestedMap map[string]map[string]string `json:"nestedMap,omitempty"`

	// Checks that multiply-nested maps work
	NestedNestedMap map[string]map[string]map[string]string `json:"nestedNestedMap,omitempty"`

	// Checks that maps containing types that contain maps work
	ContainsNestedMapMap map[string]ContainsNestedMap `json:"nestedMapInStruct,omitempty"`

	// Maps of arrays of things-that-aren’t-strings are permitted
	MapOfArraysOfFloats map[string][]bool `json:"mapOfArraysOfFloats,omitempty"`

	// +kubebuilder:validation:Minimum=-0.5
	// +kubebuilder:validation:Maximum=1.5
	// +kubebuilder:validation:MultipleOf=0.5
	FloatWithValidations float64 `json:"floatWithValidations"`

	// +kubebuilder:validation:Minimum=-0.5
	// +kubebuilder:validation:Maximum=1.5
	// +kubebuilder:validation:MultipleOf=0.5
	Float64WithValidations float64 `json:"float64WithValidations"`

	// +kubebuilder:validation:Minimum=-2
	// +kubebuilder:validation:Maximum=2
	// +kubebuilder:validation:MultipleOf=2
	IntWithValidations int `json:"intWithValidations"`

	// +kubebuilder:validation:Minimum=-2
	// +kubebuilder:validation:Maximum=2
	// +kubebuilder:validation:MultipleOf=2
	Int32WithValidations int32 `json:"int32WithValidations"`

	// This tests that unexported fields are skipped in the schema generation
	unexportedField string

	// This tests that both unexported and exported inline fields are not skipped in the schema generation
	unexportedStruct `json:",inline"`
	ExportedStruct   `json:",inline"`

	// Test of the expression-based validation rule marker, with optional message.
	// +kubebuilder:validation:XValidation:rule="self.size() % 2 == 0",message="must have even length"
	// +kubebuilder:validation:XValidation:rule="true"
	StringWithEvenLength string `json:"stringWithEvenLength,omitempty"`

	// Test of the expression-based validation with messageExpression marker.
	// +kubebuilder:validation:XValidation:rule="self.size() % 2 == 0",messageExpression="'Length has to be even but is ' + len(self.stringWithEvenLengthAndMessageExpression) + ' instead'"
	StringWithEvenLengthAndMessageExpression string `json:"stringWithEvenLengthAndMessageExpression,omitempty"`

	// Test of the expression-based validation on both field and type.
	// +kubebuilder:validation:XValidation:rule="self.startsWith('good-')",message="must have good prefix"
	StringWithEvenLengthAndGoodPrefix StringEvenType `json:"stringWithEvenLengthAndGoodPrefix,omitempty"`

	// Test that we can add a forbidden field using XValidation Reason and FieldPath.
	// The validation is applied to the spec struct itself and not the field.
	ForbiddenInt int `json:"forbiddenInt,omitempty"`

	// Checks that fixed-length arrays work
	Array [3]int `json:"array,omitempty"`

	// Checks that arrays work when the type contains a composite literal
	ArrayUsingCompositeLiteral [len(struct{ X [3]int }{}.X)]string `json:"arrayUsingCompositeLiteral,omitempty"`

	// This tests string slice item validation.
	// +kubebuilder:validation:MinItems=1
	// +kubebuilder:validation:items:MinLength=1
	// +kubebuilder:validation:items:MaxLength=255
	// +kubebuilder:validation:items:Pattern=^[a-z0-9]([-a-z0-9]*[a-z0-9])?([.][a-z0-9]([-a-z0-9]*[a-z0-9])?)*$
	// +listType=set
	Hosts []string `json:"hosts,omitempty"`

	HostsAlias Hosts `json:"hostsAlias,omitempty"`

	// This tests string slice validation.
	// +kubebuilder:validation:MinItems=2
	// +kubebuilder:validation:MaxItems=2
	StringPair []string `json:"stringPair"`

	// This tests string alias slice item validation.
	// +kubebuilder:validation:MinItems=3
	LongerStringArray []LongerString `json:"longerStringArray,omitempty"`

	// This tests that a slice of IntOrString can also have a pattern attached to it.
	// This can be useful if you want to limit the string to a percentage or integer.
	// The XIntOrString marker is a requirement for having a pattern on this type.
	// +kubebuilder:validation:items:XIntOrString
	// +kubebuilder:validation:items:Pattern="^((100|[0-9]{1,2})%|[0-9]+)$"
	IntOrStringArrayWithAPattern []*intstr.IntOrString `json:"intOrStringArrayWithAPattern,omitempty"`
}

type ContainsNestedMap struct {
	InnerMap map[string]string `json:"innerMap,omitempty"`
}

// +kubebuilder:validation:Type=object
// +kubebuilder:pruning:PreserveUnknownFields
type Preserved struct {
	ConcreteField string                 `json:"concreteField"`
	Rest          map[string]interface{} `json:"-"`
}

func (p *Preserved) UnmarshalJSON(data []byte) error {
	if err := json.Unmarshal(data, &p.Rest); err != nil {
		return err
	}
	conc, found := p.Rest["concreteField"]
	if !found {
		return nil
	}
	concStr, isStr := conc.(string)
	if !isStr {
		return fmt.Errorf("concreteField was not string")
	}
	delete(p.Rest, "concreteField")
	p.ConcreteField = concStr
	return nil
}

func (p *Preserved) MarshalJSON() ([]byte, error) {
	full := make(map[string]interface{}, len(p.Rest)+1)
	for k, v := range p.Rest {
		full[k] = v
	}
	full["concreteField"] = p.ConcreteField
	return json.Marshal(full)
}

type NestedObject struct {
	Foo string `json:"foo"`
	Bar bool   `json:"bar"`
}

// +structType=atomic
type NestedStructWithSeveralFields NestedObject

type JustNestedObject NestedObject

// +kubebuilder:validation:MinProperties=1
// +kubebuilder:validation:MaxProperties=2
type MinMaxObject struct {
	Foo string `json:"foo,omitempty"`
	Bar string `json:"bar,omitempty"`
	Baz string `json:"baz,omitempty"`
}

type EmpiableObject struct {
	// +kubebuilder:default=forty-two
	Foo string `json:"foo,omitempty"`
	Bar string `json:"bar,omitempty"`
}

type unexportedStruct struct {
	// This tests that exported fields are not skipped in the schema generation
	Foo string `json:"foo"`

	// This tests that unexported fields are skipped in the schema generation
	bar string
}

type ExportedStruct struct {
	// This tests that exported fields are not skipped in the schema generation
	Baz string `json:"baz"`

	// This tests that unexported fields are skipped in the schema generation
	qux string
}

type RootObject struct {
	Nested NestedObject `json:"nested"`
}

type AssociativeType struct {
	Name      string `json:"name"`
	Secondary int    `json:"secondary"`
	Foo       string `json:"foo"`
}

// +listType=map
// +listMapKey=name
// +listMapKey=secondary
type NestedAssociativeList []AssociativeType

// +mapType=granular
type NestedMapOfInfo map[string][]byte

// +kubebuilder:validation:MinLength=4
// This tests that markers that are allowed on both fields and types are applied to types
type LongerString string

// use an explicit type marker to verify that apply-first markers generate properly

// +kubebuilder:validation:Type=string
// TotallyABool is a bool that serializes as a string.
type TotallyABool bool

// This tests string slice item validation.
// +kubebuilder:validation:MinItems=1
// +kubebuilder:validation:items:MinLength=1
// +kubebuilder:validation:items:MaxLength=255
// +kubebuilder:validation:items:Pattern=^[a-z0-9]([-a-z0-9]*[a-z0-9])?([.][a-z0-9]([-a-z0-9]*[a-z0-9])?)*$
// +listType=set
type Hosts []string

func (t TotallyABool) MarshalJSON() ([]byte, error) {
	if t {
		return []byte(`"true"`), nil
	} else {
		return []byte(`"false"`), nil
	}
}

func (t *TotallyABool) UnmarshalJSON(in []byte) error {
	switch string(in) {
	case `"true"`:
		*t = true
		return nil
	case `"false"`:
		*t = false
	default:
		return fmt.Errorf("bad TotallyABool value %q", string(in))
	}
	return nil
}

// +kubebuilder:validation:Type=string
// URL wraps url.URL.
// It has custom json marshal methods that enable it to be used in K8s CRDs
// such that the CRD resource will have the URL but operator code can can work with url.URL struct
type URL struct {
	url.URL
}

func (u *URL) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("%q", u.String())), nil
}

func (u *URL) UnmarshalJSON(b []byte) error {
	var ref string
	if err := json.Unmarshal(b, &ref); err != nil {
		return err
	}
	if ref == "" {
		*u = URL{}
		return nil
	}

	r, err := url.Parse(ref)
	if err != nil {
		return err
	} else if r != nil {
		*u = URL{*r}
	} else {
		*u = URL{}
	}
	return nil
}

func (u *URL) String() string {
	if u == nil {
		return ""
	}
	return u.URL.String()
}

// +kubebuilder:validation:Type=string
// URL2 is an alias of url.URL.
// It has custom json marshal methods that enable it to be used in K8s CRDs
// such that the CRD resource will have the URL but operator code can can work with url.URL struct
type URL2 url.URL

func (u *URL2) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("%q", u.String())), nil
}

func (u *URL2) UnmarshalJSON(b []byte) error {
	var ref string
	if err := json.Unmarshal(b, &ref); err != nil {
		return err
	}
	if ref == "" {
		*u = URL2{}
		return nil
	}

	r, err := url.Parse(ref)
	if err != nil {
		return err
	} else if r != nil {
		*u = *(*URL2)(r)
	} else {
		*u = URL2{}
	}
	return nil
}

func (u *URL2) String() string {
	if u == nil {
		return ""
	}
	return (*url.URL)(u).String()
}

// Duration has a custom Marshaler but no markers.
// We want the CRD generation to infer type information
// from the go types and ignore the presense of the Marshaler.
type Duration struct {
	Value time.Duration `json:"value"`
}

func (d Duration) MarshalJSON() ([]byte, error) {
	type durationWithoutUnmarshaler Duration
	return json.Marshal(durationWithoutUnmarshaler(d))
}

var _ json.Marshaler = Duration{}

// ConcurrencyPolicy describes how the job will be handled.
// Only one of the following concurrent policies may be specified.
// If none of the following policies is specified, the default one
// is AllowConcurrent.
// +kubebuilder:validation:Enum=Allow;Forbid;Replace
type ConcurrencyPolicy string

const (
	// AllowConcurrent allows CronJobs to run concurrently.
	AllowConcurrent ConcurrencyPolicy = "Allow"

	// ForbidConcurrent forbids concurrent runs, skipping next run if previous
	// hasn't finished yet.
	ForbidConcurrent ConcurrencyPolicy = "Forbid"

	// ReplaceConcurrent cancels currently running job and replaces it with a new one.
	ReplaceConcurrent ConcurrencyPolicy = "Replace"
)

// StringEvenType is a type that includes an expression-based validation.
// +kubebuilder:validation:XValidation:rule="self.size() % 2 == 0",message="must have even length"
type StringEvenType string

// CronJobStatus defines the observed state of CronJob
type CronJobStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// A list of pointers to currently running jobs.
	// +optional
	Active []corev1.ObjectReference `json:"active,omitempty"`

	// Information when was the last time the job was successfully scheduled.
	// +optional
	LastScheduleTime *metav1.Time `json:"lastScheduleTime,omitempty"`

	// Information about the last time the job was successfully scheduled,
	// with microsecond precision.
	// +optional
	LastScheduleMicroTime *metav1.MicroTime `json:"lastScheduleMicroTime,omitempty"`

	// LastActiveLogURL specifies the logging url for the last started job
	// +optional
	LastActiveLogURL *URL `json:"lastActiveLogURL,omitempty"`

	// LastActiveLogURL2 specifies the logging url for the last started job
	// +optional
	LastActiveLogURL2 *URL2 `json:"lastActiveLogURL2,omitempty"`

	Runtime *Duration `json:"duration,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:singular=mycronjob
// +kubebuilder:storageversion
// +kubebuilder:metadata:annotations="api-approved.kubernetes.io=https://github.com/kubernetes-sigs/controller-tools";"cert-manager.io/inject-ca-from-secret=cert-manager/cert-manager-webhook-ca"

// CronJob is the Schema for the cronjobs API
type CronJob struct {
	/*
	 */
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CronJobSpec   `json:"spec,omitempty"`
	Status CronJobStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// CronJobList contains a list of CronJob
type CronJobList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CronJob `json:"items"`
}
