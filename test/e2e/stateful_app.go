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
			"argocd-operator",
			"argocd-catalog",
			"argocd-operator",
			"alpha",
			"quay.io/jmckind/argocd-operator-registry@sha256:45f9b0ad3722cd45125347e4e9d149728200221341583a998158d1b5b4761756",
		)

		// Deploy deploys the app via operator mechanism
		// it validates every step is finished successfully and panics if an error occurs
		testApp.Deploy()
	})
})
