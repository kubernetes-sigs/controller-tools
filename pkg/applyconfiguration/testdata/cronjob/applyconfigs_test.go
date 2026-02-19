package cronjob

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"sigs.k8s.io/controller-runtime/pkg/client"
	cronjobsv1 "sigs.k8s.io/controller-tools/pkg/applyconfiguration/testdata/cronjob/api/v1"
	cronjobsv1acs "sigs.k8s.io/controller-tools/pkg/applyconfiguration/testdata/cronjob/api/v1/applyconfiguration/api/v1"
)

var _ = Describe("ApplyConfigurations", func() {
	It("should only extract finalizers belonging to the current fieldOwner", func(ctx SpecContext) {
		const namespace, name = "default", "test"
		first := cronjobsv1acs.CronJob(name, namespace).WithFinalizers("foo.bar")
		Expect(k8sClient.Apply(ctx, first, client.FieldOwner("first"))).To(Succeed())

		second := cronjobsv1acs.CronJob(name, namespace).WithFinalizers("foo.baz")
		Expect(k8sClient.Apply(ctx, second, client.FieldOwner("second"))).To(Succeed())

		cronJob := cronjobsv1.CronJob{}
		Expect(k8sClient.Get(ctx, client.ObjectKey{Namespace: namespace, Name: name}, &cronJob)).To(Succeed())
		Expect(cronJob.Finalizers).To(Equal([]string{"foo.bar", "foo.baz"}))

		first, err := cronjobsv1acs.ExtractCronJob(&cronJob, "first")
		Expect(err).ToNot(HaveOccurred())
		Expect(first.Finalizers).To(Equal([]string{"foo.bar"}))
	})
})
