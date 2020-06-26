package e2e

// Copyright (c) Microsoft Corporation.
// Licensed under the Apache License 2.0.

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/Azure/ARO-RP/test/util/project"
)

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
		testApp := NewTestApp(
			testNamespace,
			"etcd",
			"community-operators",
			"etcd",
			"singlenamespace-alpha",
			"quay.io/coreos/etcd-operator@sha256:66a37fd61a06a43969854ee6d3e21087a98b93838e284a6086b13917f96b0d9b",
		)

		// Deploy deploys the app via operator mechanism
		// it validates every step is finished successfully and panics if an error occurs
		testApp.Deploy()
	})
})
