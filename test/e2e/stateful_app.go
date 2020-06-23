package e2e

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/operator-framework/operator-lifecycle-manager/pkg/api/apis/operators/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/Azure/ARO-RP/test/util/project"
)

var _ = Describe("Stateful app", func() {
	FSpecify("should create and validate test apps", func() {

		const (
			packageName   = "busybox"
			channelName   = "alpha"
			subName       = "test-subscription"
			secretName    = "mysecret"
			configmapName = "special-config"
			testNamespace = "test-e2e"
		)

		const (
			sourceName = "test-catalog"
			imageName  = "quay.io/olmtest/single-bundle-index:objects"
		)

		err := project.CreateProject(projectV1, testNamespace)
		// Expect(err).NotTo(HaveOccurred(), "failed to create a test namespace")

		// create catalog source
		source := &v1alpha1.CatalogSource{
			TypeMeta: metav1.TypeMeta{
				Kind:       v1alpha1.CatalogSourceKind,
				APIVersion: v1alpha1.CatalogSourceCRDAPIVersion,
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      sourceName,
				Namespace: testNamespace,
				Labels:    map[string]string{"olm.catalogSource": sourceName},
			},
			Spec: v1alpha1.CatalogSourceSpec{
				SourceType: v1alpha1.SourceTypeGrpc,
				Image:      imageName,
			},
		}

		source, err = operatorClient.OperatorsV1alpha1().CatalogSources(source.GetNamespace()).Create(source)
		Expect(err).ToNot(HaveOccurred(), "could not create catalog source")

		// Create a Subscription for package
		// _ = createSubscriptionForCatalog(operatorClient, source.GetNamespace(), subName, source.GetName(), packageName, channelName, "", v1alpha1.ApprovalAutomatic)

		// Wait for the Subscription to succeed
		// Eventually(func() error {
		// 	_, err = fetchSubscription(operatorClient, testNamespace, subName, subscriptionStateAtLatestChecker)
		// 	return err
		// }).Should(BeNil())

		// confirm extra bundle objects (secret and configmap) are installed
		Eventually(func() error {
			_, err := clients.Kubernetes.CoreV1().Secrets(testNamespace).Get(secretName, metav1.GetOptions{}) //.GetSecret(testNamespace, secretName)
			return err
		}).Should(Succeed(), "expected no error getting secret object associated with CSV")

		// Eventually(func() error {
		// 	_, err := kubeClient.GetConfigMap(testNamespace, configmapName)
		// 	return err
		// }).Should(Succeed(), "expected no error getting configmap object associated with CSV")

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
