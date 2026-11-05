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

// Package common has named types that the Plugin API uses from another
// package.
package common

// Token is a bearer token. It is a named byte slice, so it is serialized as
// a base64-encoded string.
type Token []byte

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
