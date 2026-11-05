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

// Package rawmessage has the Plugin API. Its fields use named slice and map
// types, such as json.RawMessage, a CA bundle and a list of ports. Each field
// must get the type of the schema generated for its named type. A named byte
// slice such as json.RawMessage is a base64-encoded string, like []byte, and
// must not be generated as an array.
//
// +groupName=testdata.kubebuilder.io
// +versionName=v1
package rawmessage

import (
	"encoding/json"
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"testdata.kubebuilder.io/cronjob/rawmessage/common"
)

// Plugin is an add-on that runs next to an application.
// +kubebuilder:object:root=true
type Plugin struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec PluginSpec `json:"spec"`
}

// +kubebuilder:object:root=true
type PluginList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Plugin `json:"items"`
}

// SpecDescriptor describes a field of the plugin spec, like the spec
// descriptors of an Operator Lifecycle Manager ClusterServiceVersion.
type SpecDescriptor struct {
	// Path is the path of the field, for example "spec.replicas".
	Path string `json:"path"`

	// Value is the default value of the field as raw JSON.
	// +optional
	Value json.RawMessage `json:"value,omitempty"`
}

// PEMBundle is a PEM-encoded certificate bundle. It is a named byte slice,
// so it is serialized as a base64-encoded string.
type PEMBundle []byte

// Certificate is a type alias to a PEM bundle.
type Certificate = PEMBundle

// Values holds plugin values, like Helm chart values. It is a named byte
// slice, but its type markers make it a JSON object.
// +kubebuilder:validation:Type=object
// +kubebuilder:pruning:PreserveUnknownFields
type Values []byte

// Tags is a list of tags. It has no type markers, so it is an array.
type Tags []string

// Keywords is a type alias to a list of tags.
type Keywords = Tags

// Topics is a type alias to Keywords, so an alias of an alias of a list of tags.
type Topics = Keywords

// Labels is a map of labels. It has no type markers, so it is an object.
type Labels map[string]string

// HostList is a list of hosts that is serialized as one comma-separated
// string, for example "example.com,example.org". Its type marker makes it a
// string.
// +kubebuilder:validation:Type=string
type HostList []string

func (h HostList) MarshalJSON() ([]byte, error) {
	return json.Marshal(strings.Join(h, ","))
}

// DeniedHosts is a list of hosts declared from HostList, so its schema is the
// string schema of HostList. Go gives it none of the methods of HostList, so it
// has its own MarshalJSON.
type DeniedHosts HostList

func (h DeniedHosts) MarshalJSON() ([]byte, error) {
	return json.Marshal(strings.Join(h, ","))
}

// Scopes is a list of OAuth scopes with custom JSON marshalling. It has no
// type marker, so it is generated from its Go type, as an array.
type Scopes []string

func (s Scopes) MarshalJSON() ([]byte, error) {
	return json.Marshal([]string(s))
}

// Headers is a map of HTTP headers with custom JSON marshalling. It has no
// type marker, so it is generated from its Go type, as an object.
type Headers map[string]string

func (h Headers) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]string(h))
}

// VersionList is a list of versions that is serialized as text, for example
// "v1,v2". Text marshalling makes it a string.
type VersionList []string

func (v VersionList) MarshalText() ([]byte, error) {
	return []byte(strings.Join(v, ",")), nil
}

// Port is a network port exposed by a container.
type Port struct {
	// Name is the name of the port, for example "http".
	Name string `json:"name"`

	// Protocol is the protocol of the port, for example "TCP".
	Protocol string `json:"protocol"`

	// Port is the port number.
	Port int32 `json:"port"`
}

// Ports is a list of ports, merged by port name.
// +listType=map
// +listMapKey=name
type Ports []Port

// ContainerPorts is a list of ports declared from another named list.
type ContainerPorts Ports

// DNSNames is a list of DNS names declared from another named list.
type DNSNames Tags

// Args is a list of command arguments with an explicit type marker.
// +kubebuilder:validation:Type=array
type Args []string

// Annotations is a map of annotations with an explicit type marker.
// +kubebuilder:validation:Type=object
type Annotations map[string]string

// Nameservers lists up to three DNS servers, like resolv.conf. It is a named
// fixed-size array.
type Nameservers [3]string

type PluginSpec struct {
	// Config is the plugin configuration as raw JSON.
	// +optional
	Config json.RawMessage `json:"config,omitempty"`

	// SpecDescriptors describe the fields of the plugin spec. Each one has a
	// default value as raw JSON.
	// +optional
	SpecDescriptors []SpecDescriptor `json:"specDescriptors,omitempty"`

	// CABundle is the PEM-encoded CA bundle used to verify the plugin server.
	// +optional
	CABundle PEMBundle `json:"caBundle,omitempty"`

	// ClientCertificate is the PEM-encoded client certificate. Its type is a
	// type alias to a named byte slice.
	// +optional
	ClientCertificate Certificate `json:"clientCertificate,omitempty"`

	// AuthToken is the token used to call the plugin server. Its type is a
	// named byte slice from another package.
	// +optional
	AuthToken common.Token `json:"authToken,omitempty"`

	// Values are the plugin values. Its type is a byte slice typed as an object
	// by its type markers.
	// +optional
	Values Values `json:"values,omitempty"`

	// Tags is a named list with a field-level list marker.
	// +listType=set
	// +optional
	Tags Tags `json:"tags,omitempty"`

	// Keywords is a type alias to a named list, with a field-level list marker.
	// +listType=set
	// +optional
	Keywords Keywords `json:"keywords,omitempty"`

	// Topics is an alias of an alias of a named list, with a field-level list marker.
	// +listType=set
	// +optional
	Topics Topics `json:"topics,omitempty"`

	// Labels is a named map with a field-level map marker.
	// +mapType=granular
	// +optional
	Labels Labels `json:"labels,omitempty"`

	// AllowedHosts is a named list with custom JSON marshalling and a type marker.
	// +optional
	AllowedHosts HostList `json:"allowedHosts,omitempty"`

	// DeniedHosts is declared from a named list with a type marker, and has its
	// own custom JSON marshalling.
	// +optional
	DeniedHosts DeniedHosts `json:"deniedHosts,omitempty"`

	// Scopes is a named list with custom JSON marshalling and a field-level list marker.
	// +listType=set
	// +optional
	Scopes Scopes `json:"scopes,omitempty"`

	// Headers is a named map with custom JSON marshalling and a field-level map marker.
	// +mapType=granular
	// +optional
	Headers Headers `json:"headers,omitempty"`

	// SupportedVersions is a named list with custom text marshalling.
	// +optional
	SupportedVersions VersionList `json:"supportedVersions,omitempty"`

	// Ports uses the list type and keys of its named list: a map list keyed
	// by name.
	// +optional
	Ports Ports `json:"ports,omitempty"`

	// ExposedPorts is replaced as a whole on update, so it can have two ports
	// with the same name.
	// +listType=atomic
	// +optional
	ExposedPorts Ports `json:"exposedPorts,omitempty"`

	// ProtocolPorts has at most one port per protocol. It sets its own key on
	// the named list.
	// +listType=map
	// +listMapKey=protocol
	// +optional
	ProtocolPorts Ports `json:"protocolPorts,omitempty"`

	// ServicePorts must have between 1 and 10 ports.
	// +kubebuilder:validation:MinItems=1
	// +kubebuilder:validation:MaxItems=10
	// +optional
	ServicePorts Ports `json:"servicePorts,omitempty"`

	// MetricsPorts is replaced as a whole on update. It uses the k8s list
	// marker.
	// +k8s:listType=atomic
	// +optional
	MetricsPorts Ports `json:"metricsPorts,omitempty"`

	// HealthPorts has at most one port per protocol. It uses the k8s list
	// markers.
	// +k8s:listType=map
	// +k8s:listMapKey=protocol
	// +optional
	HealthPorts Ports `json:"healthPorts,omitempty"`

	// ContainerPorts is a named list declared from another named list.
	// +listType=map
	// +listMapKey=name
	// +optional
	ContainerPorts ContainerPorts `json:"containerPorts,omitempty"`

	// SidecarPorts is a named list from another package. It is replaced as a
	// whole on update.
	// +listType=atomic
	// +optional
	SidecarPorts common.Ports `json:"sidecarPorts,omitempty"`

	// DNSNames is a named list declared from another named list.
	// +listType=set
	// +optional
	DNSNames DNSNames `json:"dnsNames,omitempty"`

	// ExcludedTags is a pointer to a named list.
	// +listType=atomic
	// +nullable
	// +optional
	ExcludedTags *Tags `json:"excludedTags,omitempty"`

	// Args is a named list with an explicit type marker.
	// +listType=atomic
	// +optional
	Args Args `json:"args,omitempty"`

	// Annotations is a named map with an explicit type marker.
	// +mapType=atomic
	// +optional
	Annotations Annotations `json:"annotations,omitempty"`

	// Nameservers is a named fixed-size array with a field-level list marker.
	// +listType=atomic
	// +optional
	Nameservers Nameservers `json:"nameservers,omitempty"`

	// Selector is a named map with an atomic field-level map marker.
	// +mapType=atomic
	// +optional
	Selector Labels `json:"selector,omitempty"`

	// NodeSelector is a named map with no field-level marker.
	// +optional
	NodeSelector Labels `json:"nodeSelector,omitempty"`

	// Requests is a named map from another module, with a field-level map marker.
	// +mapType=atomic
	// +optional
	Requests corev1.ResourceList `json:"requests,omitempty"`
}
