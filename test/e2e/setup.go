package e2e

// Copyright (c) Microsoft Corporation.
// Licensed under the Apache License 2.0.

import (
	"fmt"
	"os"

	. "github.com/onsi/ginkgo"

	"github.com/Azure/go-autorest/autorest/azure/auth"
	//	"github.com/openshift/client-go/config/clientset/versioned"
	projectv1client "github.com/openshift/client-go/project/clientset/versioned/typed/project/v1"
	machineapiclient "github.com/openshift/machine-api-operator/pkg/generated/clientset/versioned"
	"github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/Azure/ARO-RP/pkg/util/azureclient/mgmt/redhatopenshift"
	"github.com/operator-framework/operator-lifecycle-manager/pkg/api/client/clientset/versioned"
)

type clientSet struct {
	OpenshiftClusters redhatopenshift.OpenShiftClustersClient
	Operations        redhatopenshift.OperationsClient
	Kubernetes        kubernetes.Interface
	MachineAPI        machineapiclient.Interface
	Project           projectv1client.ProjectV1Interface
	OLMClient         versioned.Interface
}

var (
	log     *logrus.Entry
	clients *clientSet
)

func newClientSet() (*clientSet, error) {
	authorizer, err := auth.NewAuthorizerFromEnvironment()
	if err != nil {
		return nil, err
	}

	kubeconfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		clientcmd.NewDefaultClientConfigLoadingRules(),
		&clientcmd.ConfigOverrides{},
	)

	restConfig, err := kubeconfig.ClientConfig()
	if err != nil {
		return nil, err
	}

	cli := kubernetes.NewForConfigOrDie(restConfig)
	machineapicli := machineapiclient.NewForConfigOrDie(restConfig)

	return &clientSet{
		OpenshiftClusters: redhatopenshift.NewOpenShiftClustersClient(os.Getenv("AZURE_SUBSCRIPTION_ID"), authorizer),
		Operations:        redhatopenshift.NewOperationsClient(os.Getenv("AZURE_SUBSCRIPTION_ID"), authorizer),
		Kubernetes:        cli,
		MachineAPI:        machineapicli,
		Project:           projectv1client.NewForConfigOrDie(restConfig),
		OLMClient:         versioned.NewForConfigOrDie(restConfig),
	}, nil
}

var _ = BeforeSuite(func() {
	log.Info("BeforeSuite")
	for _, key := range []string{
		"AZURE_SUBSCRIPTION_ID",
		"CLUSTER",
		"RESOURCEGROUP",
	} {
		if _, found := os.LookupEnv(key); !found {
			panic(fmt.Sprintf("environment variable %q unset", key))
		}
	}

	var err error
	clients, err = newClientSet()
	if err != nil {
		panic(err)
	}
})
