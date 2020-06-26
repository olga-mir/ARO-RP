package e2e

// Copyright (c) Microsoft Corporation.
// Licensed under the Apache License 2.0.

import (
	"time"

	// . "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	v1 "github.com/operator-framework/operator-lifecycle-manager/pkg/api/apis/operators/v1"
	"github.com/operator-framework/operator-lifecycle-manager/pkg/api/apis/operators/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
)

const pollInterval = 1 * time.Second
const pollDuration = 2 * time.Minute

// TestApp provides all necessary information to install a namespaced app from
// an operator and preconfigured clients to communicate with the cluster
type TestApp struct {
	Namespace         string
	OperatorName      string
	CatalogSourceName string
	SubscriptionName  string
	ChannelName       string
	Image             string
}

// NewTestApp constructs a new TestApp object
func NewTestApp(ns, operatorName, catalogSourceName, subscriptionName, channelName, image string) TestApp {
	return TestApp{
		Namespace:         ns,
		OperatorName:      operatorName,
		CatalogSourceName: catalogSourceName,
		SubscriptionName:  subscriptionName,
		ChannelName:       channelName,
		Image:             image,
	}
}

// Deploy deploys an app using operator.
// It blocks and validates that every step is finished successfully
// This function exists if it encounters an error
func (app *TestApp) Deploy() {
	// Deploy CatalogSource and validate it is created successfully
	catalogSource, err := app.createCatalogSource()
	Expect(err).ToNot(HaveOccurred(), "failed to create catalog source")

	Eventually(func() error {
		_, err = app.fetchCatalogSource()
		return err
	}).Should(BeNil(), "CatalogSource not found")

	// Deploy OperatorGroup and validate it is created successfully
	_, err = app.createOperatorGroup(catalogSource)
	Expect(err).ToNot(HaveOccurred(), "Failed to create operator group")

	Eventually(func() error {
		_, err = app.fetchOperatorGroup()
		return err
	}).Should(BeNil(), "OperatorGroup not found")

	// Deploy Subscription and validate it is created successfully
	_, err = app.createSubscriptionForCatalog(catalogSource)
	Expect(err).ToNot(HaveOccurred(), "Failed to create subscription")

	Eventually(func() error {
		_, err = app.fetchSubscription()
		return err
	}).Should(BeNil(), "Subscription not found")

	// When Subscription is created, InstallPlan should be created automatically
	Eventually(func() error {
		_, err := app.fetchInstallPlan()
		return err
	}).Should(BeNil(), "InstallPlan not found")
}

func (app *TestApp) createCatalogSource() (*v1alpha1.CatalogSource, error) {
	catalogSource := &v1alpha1.CatalogSource{
		TypeMeta: metav1.TypeMeta{
			Kind:       v1alpha1.CatalogSourceKind,
			APIVersion: v1alpha1.CatalogSourceCRDAPIVersion,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      app.CatalogSourceName,
			Namespace: app.Namespace,
			Labels:    map[string]string{"olm.catalogSource": app.CatalogSourceName},
		},
		Spec: v1alpha1.CatalogSourceSpec{
			SourceType:  v1alpha1.SourceTypeGrpc,
			Image:       app.Image,
			DisplayName: "Argo CD Operators",
			Publisher:   "Argo CD Community",
		},
	}
	return clients.OLMClient.OperatorsV1alpha1().CatalogSources(app.Namespace).Create(catalogSource)
}

func (app *TestApp) createOperatorGroup(catalogSource *v1alpha1.CatalogSource) (*v1.OperatorGroup, error) {
	ns := catalogSource.GetNamespace()
	operatorGroup := &v1.OperatorGroup{
		TypeMeta: metav1.TypeMeta{
			Kind: v1.OperatorGroupKind,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: app.OperatorName,
		},
		Spec: v1.OperatorGroupSpec{
			TargetNamespaces: []string{
				ns,
			},
		},
	}

	return clients.OLMClient.OperatorsV1().OperatorGroups(ns).Create(operatorGroup)
}

func (app *TestApp) createSubscriptionForCatalog(catalogSource *v1alpha1.CatalogSource) (*v1alpha1.Subscription, error) {
	ns := catalogSource.GetNamespace()
	subscription := &v1alpha1.Subscription{
		TypeMeta: metav1.TypeMeta{
			Kind:       v1alpha1.SubscriptionKind,
			APIVersion: v1alpha1.SubscriptionCRDAPIVersion,
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: ns,
			Name:      app.SubscriptionName,
		},
		Spec: &v1alpha1.SubscriptionSpec{
			CatalogSource:          catalogSource.GetName(),
			CatalogSourceNamespace: ns,
			Package:                app.SubscriptionName, // TODO need package for installplan, but can be diff name?
			Channel:                app.ChannelName,
			InstallPlanApproval:    v1alpha1.ApprovalAutomatic,
		},
	}

	return clients.OLMClient.OperatorsV1alpha1().Subscriptions(ns).Create(subscription)
}

func (app *TestApp) fetchCatalogSource() (fetchedCatalogSource *v1alpha1.CatalogSource, err error) {
	err = wait.Poll(pollInterval, pollDuration, func() (bool, error) {
		fetchedCatalogSource, err = clients.OLMClient.OperatorsV1alpha1().CatalogSources(app.Namespace).Get(app.CatalogSourceName, metav1.GetOptions{})
		if err != nil || fetchedCatalogSource == nil {
			return false, err
		}
		return true, nil
	})
	return fetchedCatalogSource, nil
}

func (app *TestApp) fetchOperatorGroup() (fetchedOperatorGroup *v1.OperatorGroup, err error) {
	err = wait.Poll(pollInterval, pollDuration, func() (bool, error) {
		fetchedOperatorGroup, err = clients.OLMClient.OperatorsV1().OperatorGroups(app.Namespace).Get(app.OperatorName, metav1.GetOptions{})
		if err != nil || fetchedOperatorGroup == nil {
			return false, err
		}
		return true, nil
	})
	return fetchedOperatorGroup, err
}

func (app *TestApp) fetchSubscription() (fetchedSubscription *v1alpha1.Subscription, err error) {
	err = wait.Poll(pollInterval, pollDuration, func() (bool, error) {
		fetchedSubscription, err = clients.OLMClient.OperatorsV1alpha1().Subscriptions(app.Namespace).Get(app.SubscriptionName, metav1.GetOptions{})
		if err != nil || fetchedSubscription == nil {
			return false, err
		}
		return true, nil
	})
	return fetchedSubscription, err
}

// fetches list of InstallPlans in the namespace and returns a first one that
// matches provided subscription. InstallPlan has autogenerated name
// therefore we have to fetch list and find it by the properties.
func (app *TestApp) fetchInstallPlan() (fetchedInstallPlan *v1alpha1.InstallPlan, err error) {
	err = wait.Poll(pollInterval, pollDuration, func() (bool, error) {
		list, err := clients.OLMClient.OperatorsV1alpha1().InstallPlans(app.Namespace).List(metav1.ListOptions{})
		if err != nil || list == nil || len(list.Items) == 0 {
			return false, err
		}
		for _, i := range list.Items {
			for _, r := range i.ObjectMeta.OwnerReferences {
				if r.Kind == "Subscription" && r.Name == app.SubscriptionName {
					fetchedInstallPlan = &i
					return true, nil
				}
			}
		}
		return true, nil
	})
	return fetchedInstallPlan, err
}
