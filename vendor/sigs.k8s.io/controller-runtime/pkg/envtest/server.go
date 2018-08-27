/*
Copyright 2016 The Kubernetes Authors.

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

package envtest

import (
	"os"

	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/testing_frameworks/integration"
)

// Default binary path for test framework
const (
	envKubeAPIServerBin     = "TEST_ASSET_KUBE_APISERVER"
	envEtcdBin              = "TEST_ASSET_ETCD"
	defaultKubeAPIServerBin = "/usr/local/kubebuilder/bin/kube-apiserver"
	defaultEtcdBin          = "/usr/local/kubebuilder/bin/etcd"
	StartTimeout            = 60
	StopTimeout             = 60
)

// Environment creates a Kubernetes test environment that will start / stop the Kubernetes control plane and
// install extension APIs
type Environment struct {
	// ControlPlane is the ControlPlane including the apiserver and etcd
	ControlPlane integration.ControlPlane

	// Config can be used to talk to the apiserver
	Config *rest.Config

	// CRDs is a list of CRDs to install
	CRDs []*apiextensionsv1beta1.CustomResourceDefinition

	// CRDDirectoryPaths is a list of paths containing CRD yaml or json configs.
	CRDDirectoryPaths []string
}

// Stop stops a running server
func (te *Environment) Stop() error {
	return te.ControlPlane.Stop()
}

// Start starts a local Kubernetes server and updates te.ApiserverPort with the port it is listening on
func (te *Environment) Start() (*rest.Config, error) {
	te.ControlPlane = integration.ControlPlane{}
	if os.Getenv(envKubeAPIServerBin) == "" {
		te.ControlPlane.APIServer = &integration.APIServer{Path: defaultKubeAPIServerBin}
	}
	if os.Getenv(envEtcdBin) == "" {
		te.ControlPlane.Etcd = &integration.Etcd{Path: defaultEtcdBin}
	}

	// Start the control plane - retry if it fails
	if err := te.ControlPlane.Start(); err != nil {
		return nil, err
	}

	// Create the *rest.Config for creating new clients
	te.Config = &rest.Config{
		Host: te.ControlPlane.APIURL().Host,
	}

	_, err := InstallCRDs(te.Config, CRDInstallOptions{
		Paths: te.CRDDirectoryPaths,
		CRDs:  te.CRDs,
	})
	return te.Config, err
}
