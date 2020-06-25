package e2e

import (
	"time"

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
	// err := project.CreateProject(clients.Project, testNamespace)
	// Expect(err).NotTo(HaveOccurred(), "failed to create test namespace")
})

var _ = AfterEach(func() {
	// err := project.CleanupProject(clients.Project, testNamespace)
	// Expect(err).NotTo(HaveOccurred(), "failed to delete test namespace")
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

		// Install CatalogSource and validate it is created successfully
		catalogSource, err := createCatalogSource(catalogSourceName, testNamespace, image)
		Expect(err).ToNot(HaveOccurred(), "failed to create catalog source")

		Eventually(func() error {
			_, err = fetchCatalogSource(testNamespace, catalogSourceName)
			return err
		}).Should(BeNil())

		// Install OperatorGroup and validate it is created successfully
		_, err = createOperatorGroup(catalogSource, operatorName)
		Expect(err).ToNot(HaveOccurred(), "Failed to create operator group")

		Eventually(func() error {
			_, err = fetchOperatorGroup(testNamespace, operatorName)
			return err
		}).Should(BeNil())

		// Install Subscription and validate it is created successfully
		sub, err := createSubscriptionForCatalog(catalogSource, subscriptionName, channelName)
		Expect(err).ToNot(HaveOccurred(), "Failed to create subscription")

		Eventually(func() error {
			_, err = fetchSubscription(testNamespace, sub.GetName())
			return err
		}).Should(BeNil())

		// Validate InstallPlan has been created
		Eventually(func() error {
			_, err := fetchInstallPlan(testNamespace, sub.GetName())
			// _, err := fetchInstallPlan(testNamespace, sub.GetName())
			return err
		}).Should(BeNil(), "InstallPlan not found")

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
			// Package:                "argocd-operator",
			Channel:             channel,
			InstallPlanApproval: v1alpha1.ApprovalAutomatic,
		},
	}

	return clients.OLMClient.OperatorsV1alpha1().Subscriptions(ns).Create(subscription)
}

func fetchCatalogSource(namespace, name string) (fetchedCatalogSource *v1alpha1.CatalogSource, err error) {
	err = wait.Poll(pollInterval, pollDuration, func() (bool, error) {
		fetchedCatalogSource, err = clients.OLMClient.OperatorsV1alpha1().CatalogSources(namespace).Get(name, metav1.GetOptions{})
		if err != nil || fetchedCatalogSource == nil {
			return false, err
		}
		return true, nil
	})
	return fetchedCatalogSource, nil
}

func fetchOperatorGroup(namespace, name string) (fetchedOperatorGroup *v1.OperatorGroup, err error) {
	err = wait.Poll(pollInterval, pollDuration, func() (bool, error) {
		fetchedOperatorGroup, err = clients.OLMClient.OperatorsV1().OperatorGroups(namespace).Get(name, metav1.GetOptions{})
		if err != nil || fetchedOperatorGroup == nil {
			return false, err
		}
		return true, nil
	})
	return fetchedOperatorGroup, err
}

func fetchSubscription(namespace, name string) (fetchedSubscription *v1alpha1.Subscription, err error) {
	err = wait.Poll(pollInterval, pollDuration, func() (bool, error) {
		fetchedSubscription, err = clients.OLMClient.OperatorsV1alpha1().Subscriptions(namespace).Get(name, metav1.GetOptions{})
		if err != nil || fetchedSubscription == nil {
			return false, err
		}
		return true, nil
	})
	return fetchedSubscription, err
}

func fetchInstallPlan(namespace, subscriptionName string) (fetchedInstallPlan *v1alpha1.InstallPlan, err error) {
	err = wait.Poll(pollInterval, pollDuration, func() (bool, error) {
		list, err := clients.OLMClient.OperatorsV1alpha1().InstallPlans(namespace).List(metav1.ListOptions{})
		if err != nil || list == nil || len(list.Items) == 0 {
			return false, err
		}
		for _, i := range list.Items {
			for _, r := range i.ObjectMeta.OwnerReferences {
				if r.Kind == "Subscription" && r.Name == subscriptionName {
					fetchedInstallPlan = &i
					return true, nil
				}
			}
		}
		return true, nil
	})
	return fetchedInstallPlan, err
}
