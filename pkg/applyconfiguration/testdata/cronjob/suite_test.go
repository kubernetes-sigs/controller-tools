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

package cronjob

import (
	"path/filepath"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"

	apiv1 "sigs.k8s.io/controller-tools/pkg/applyconfiguration/testdata/cronjob/api/v1"
)

var (
	testEnv   *envtest.Environment
	cfg       *rest.Config
	k8sClient client.Client
)

func TestAppyconfigurations(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "ApplyConfiguration Suite")
}

var _ = BeforeSuite(func(ctx SpecContext) {
	Expect(apiv1.AddToScheme(scheme.Scheme)).To(Succeed())

	testEnv = &envtest.Environment{
		CRDDirectoryPaths: []string{
			filepath.Join("api", "v1"),
		},
		Scheme: scheme.Scheme,
	}

	var err error
	cfg, err = testEnv.Start()
	Expect(err).NotTo(HaveOccurred())
	Expect(cfg).NotTo(BeNil())

	k8sClient, err = client.New(cfg, client.Options{})
	Expect(err).NotTo(HaveOccurred())
	Expect(k8sClient).NotTo(BeNil())
})

var _ = AfterSuite(func() {
	Expect(testEnv.Stop()).To(Succeed())
})
