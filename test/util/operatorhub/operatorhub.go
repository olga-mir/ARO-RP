package operatorhub

// Copyright (c) Microsoft Corporation.
// Licensed under the Apache License 2.0.

import (
	"errors"
	"time"

	. "github.com/onsi/gomega"

	v1 "github.com/operator-framework/operator-lifecycle-manager/pkg/api/apis/operators/v1"
	"github.com/operator-framework/operator-lifecycle-manager/pkg/api/apis/operators/v1alpha1"
	"github.com/operator-framework/operator-lifecycle-manager/pkg/api/client/clientset/versioned"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
)

const pollInterval = 10 * time.Second
const pollDuration = 1 * time.Minute

// TestApp provides all necessary information to install a namespaced app from
// an operator and preconfigured clients to communicate with the cluster
type TestApp struct {
	olmclient              versioned.Interface
	name                   string
	namespace              string
	catalogSourceName      string
	catalogSourceNamespace string
	channelName            string
	image                  string
}

// NewTestApp constructs a new TestApp object
func NewTestApp(olmclient versioned.Interface, name, namespace, catalogSourceName, catalogSourceNamespace, channelName, image string) TestApp {
	return TestApp{
		olmclient:              olmclient,
		name:                   name,
		namespace:              namespace,
		catalogSourceName:      catalogSourceName,
		catalogSourceNamespace: catalogSourceNamespace,
		channelName:            channelName,
		image:                  image,
	}
}

// Deploy deploys an app using operator.
// It blocks and validates that every step is finished successfully
// This function exists if it encounters an error
func (app *TestApp) Deploy() {
	// Deploy OperatorGroup and validate it is created successfully
	_, err := app.createOperatorGroup()
	Expect(err).ToNot(HaveOccurred(), "Failed to create operator group")

	Eventually(func() error {
		_, err = app.fetchOperatorGroup()
		return err
	}).Should(BeNil(), "OperatorGroup not found")

	// Deploy Subscription and validate it is created successfully
	_, err = app.createSubscriptionForCatalog()
	Expect(err).ToNot(HaveOccurred(), "Failed to create subscription")

	Eventually(func() error {
		_, err = app.fetchSubscription()
		return err
	}).Should(BeNil(), "Subscription not found")

	// When Subscription is created, InstallPlan should be created automatically
	Eventually(func() error {
		return app.ensureInstallPlan()
	}).Should(BeNil(), "InstallPlan not found")
}

func (app *TestApp) createOperatorGroup() (*v1.OperatorGroup, error) {
	operatorGroup := &v1.OperatorGroup{
		TypeMeta: metav1.TypeMeta{
			Kind: v1.OperatorGroupKind,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: app.name,
		},
		Spec: v1.OperatorGroupSpec{
			TargetNamespaces: []string{
				app.namespace,
			},
		},
	}
	return app.olmclient.OperatorsV1().OperatorGroups(app.namespace).Create(operatorGroup)
}

func (app *TestApp) createSubscriptionForCatalog() (*v1alpha1.Subscription, error) {
	subscription := &v1alpha1.Subscription{
		TypeMeta: metav1.TypeMeta{
			Kind:       v1alpha1.SubscriptionKind,
			APIVersion: v1alpha1.SubscriptionCRDAPIVersion,
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: app.namespace,
			Name:      app.name,
		},
		Spec: &v1alpha1.SubscriptionSpec{
			CatalogSource:          app.catalogSourceName,
			CatalogSourceNamespace: app.catalogSourceNamespace,
			Package:                app.name,
			Channel:                app.channelName,
			InstallPlanApproval:    v1alpha1.ApprovalAutomatic,
		},
	}

	return app.olmclient.OperatorsV1alpha1().Subscriptions(app.namespace).Create(subscription)
}

func (app *TestApp) fetchOperatorGroup() (fetchedOperatorGroup *v1.OperatorGroup, err error) {
	err = wait.Poll(pollInterval, pollDuration, func() (bool, error) {
		fetchedOperatorGroup, err = app.olmclient.OperatorsV1().OperatorGroups(app.namespace).Get(app.name, metav1.GetOptions{})
		if err != nil || fetchedOperatorGroup == nil {
			return false, err
		}
		return true, nil
	})
	return fetchedOperatorGroup, err
}

func (app *TestApp) fetchSubscription() (fetchedSubscription *v1alpha1.Subscription, err error) {
	err = wait.Poll(pollInterval, pollDuration, func() (bool, error) {
		fetchedSubscription, err = app.olmclient.OperatorsV1alpha1().Subscriptions(app.namespace).Get(app.name, metav1.GetOptions{})
		if err != nil || fetchedSubscription == nil {
			return false, err
		}
		return true, nil
	})
	return fetchedSubscription, err
}

func (app *TestApp) ensureInstallPlan() (err error) {
	list, err := app.olmclient.OperatorsV1alpha1().InstallPlans(app.namespace).List(metav1.ListOptions{})
	if err != nil || list == nil || len(list.Items) == 0 {
		return err
	}
	for _, i := range list.Items {
		for _, r := range i.ObjectMeta.OwnerReferences {
			if r.Kind == "Subscription" && r.Name == app.name {
				return nil
			}
		}
	}
	return errors.New("InstallPlan is not found")
}
