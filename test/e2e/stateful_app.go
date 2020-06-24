package e2e

import (
	"time"

	"github.com/Azure/ARO-RP/test/util/project"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	v1 "github.com/operator-framework/operator-lifecycle-manager/pkg/api/apis/operators/v1"
	"github.com/operator-framework/operator-lifecycle-manager/pkg/api/apis/operators/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
)

const pollInterval = 1 * time.Second
const pollDuration = 5 * time.Minute
const testNamespace = "test-e2e"

var _ = BeforeEach(func() {
	err := project.CreateProject(clients.Project, testNamespace)
	Expect(err).NotTo(HaveOccurred(), "failed to create test namespace")
})

var _ = AfterEach(func() {
	err := project.CleanupProject(clients.Project, testNamespace)
	Expect(err).NotTo(HaveOccurred(), "failed to delete test namespace")
})

var _ = Describe("Stateful app", func() {
	FSpecify("should create and validate test apps", func() {

		const (
			channelName       = "alpha"
			subscriptionName  = "argocd-operator"
			operatorName      = "argocd-operator"
			catalogSourceName = "argocd-catalog"
			image             = "quay.io/jmckind/argocd-operator-registry@sha256:45f9b0ad3722cd45125347e4e9d149728200221341583a998158d1b5b4761756"
		)

		catalogSource, err := createCatalogSource(catalogSourceName, testNamespace, image)
		Expect(err).ToNot(HaveOccurred(), "failed to create catalog source")
		time.Sleep(30 * time.Second)

		_, err = createOperatorGroup(catalogSource, operatorName)
		Expect(err).ToNot(HaveOccurred(), "Failed to create operator group")
		time.Sleep(30 * time.Second)

		sub, err := createSubscriptionForCatalog(catalogSource, subscriptionName, channelName)
		Expect(err).ToNot(HaveOccurred(), "Failed to create subscription")

		// Wait for the Subscription to succeed
		Eventually(func() error {
			_, err = fetchSubscription(testNamespace, sub.GetName())
			return err
		}).Should(BeNil())

		// confirm extra bundle objects (secret and configmap) are installed

		// Eventually(func() error {
		// 	_, err := kubeClient.GetConfigMap(testNamespace, configmapName)
		// 	return err
		// }).Should(Succeed(), "expected no error getting configmap object associated with CSV")

	})
})

func createCatalogSource(catalogSourceName, namespace, image string) (*v1alpha1.CatalogSource, error) {
	catalogSource := &v1alpha1.CatalogSource{
		TypeMeta: metav1.TypeMeta{
			Kind:       v1alpha1.CatalogSourceKind,
			APIVersion: v1alpha1.CatalogSourceCRDAPIVersion,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      catalogSourceName,
			Namespace: namespace,
			Labels:    map[string]string{"olm.catalogSource": catalogSourceName},
		},
		Spec: v1alpha1.CatalogSourceSpec{
			SourceType:  v1alpha1.SourceTypeGrpc,
			Image:       image,
			DisplayName: "Argo CD Operators",
			Publisher:   "Argo CD Community",
		},
	}
	return clients.OLMClient.OperatorsV1alpha1().CatalogSources(namespace).Create(catalogSource)
}

func createOperatorGroup(catalogSource *v1alpha1.CatalogSource, operatorName string) (*v1.OperatorGroup, error) {
	ns := catalogSource.GetNamespace()
	operatorGroup := &v1.OperatorGroup{
		TypeMeta: metav1.TypeMeta{
			Kind: v1.OperatorGroupKind,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: operatorName,
		},
		Spec: v1.OperatorGroupSpec{
			TargetNamespaces: []string{
				ns,
			},
		},
	}

	return clients.OLMClient.OperatorsV1().OperatorGroups(ns).Create(operatorGroup)
}

func createSubscriptionForCatalog(catalogSource *v1alpha1.CatalogSource, name, channel string) (*v1alpha1.Subscription, error) {
	ns := catalogSource.GetNamespace()
	subscription := &v1alpha1.Subscription{
		TypeMeta: metav1.TypeMeta{
			Kind:       v1alpha1.SubscriptionKind,
			APIVersion: v1alpha1.SubscriptionCRDAPIVersion,
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: ns,
			Name:      name,
		},
		Spec: &v1alpha1.SubscriptionSpec{
			CatalogSource:          catalogSource.GetName(),
			CatalogSourceNamespace: ns,
			Package:                "argocd-operator",
			Channel:                channel,
			InstallPlanApproval:    v1alpha1.ApprovalAutomatic,
		},
	}

	return clients.OLMClient.OperatorsV1alpha1().Subscriptions(ns).Create(subscription)
}

func fetchSubscription(namespace, name string) (*v1alpha1.Subscription, error) {
	var fetchedSubscription *v1alpha1.Subscription
	var err error

	err = wait.Poll(pollInterval, pollDuration, func() (bool, error) {
		fetchedSubscription, err = clients.OLMClient.OperatorsV1alpha1().Subscriptions(namespace).Get(name, metav1.GetOptions{})
		if err != nil || fetchedSubscription == nil {
			return false, err
		}
		return true, nil
	})
	return fetchedSubscription, err
}
