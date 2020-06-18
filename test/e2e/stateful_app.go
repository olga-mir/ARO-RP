package e2e

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/Azure/ARO-RP/test/util/project"
)

var _ = Describe("Stateful app", func() {
	FSpecify("should create and validate test apps", func() {

		project.CreateProject("e2e-test")
		// list, err := clients.Kubernetes.CoreV1().Nodes().List(metav1.ListOptions{})
		// Expect(err).NotTo(HaveOccurred())

		// ctx := context.Background()
		// By("creating test app")
		// namespace, errs := sanity.Checker.CreateTestApp(ctx)
		// Expect(errs).To(BeEmpty())
		// defer func() {
		// 	By("deleting test app")
		// 	_ = sanity.Checker.DeleteTestApp(ctx, namespace)
		// }()

		// By("validating test app")
		// errs = sanity.Checker.ValidateTestApp(ctx, namespace)
		// Expect(errs).To(BeEmpty())
	})
})
