package e2e

// Copyright (c) Microsoft Corporation.
// Licensed under the Apache License 2.0.

import (
	"context"
	"math/rand"
	"os"
	"strconv"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Update clusters", func() {
	Specify("should be possible to set the tag on a cluster via PUT", func() {
		// az resource tag --tags key=value --ids /subscriptions/xxx/resourceGroups/xxx/providers/Microsoft.RedHatOpenShift/openShiftClusters/xxx

		ctx := context.Background()
		value := strconv.Itoa(rand.Int())

		oc, err := clients.OpenshiftClusters.Get(ctx, os.Getenv("RESOURCEGROUP"), os.Getenv("CLUSTER"))
		Expect(err).NotTo(HaveOccurred())
		Expect(oc.Tags).NotTo(HaveKeyWithValue("key", &value))

		if oc.Tags == nil {
			oc.Tags = map[string]*string{}
		}
		oc.Tags["key"] = &value

		err = clients.OpenshiftClusters.CreateOrUpdateAndWait(ctx, os.Getenv("RESOURCEGROUP"), os.Getenv("CLUSTER"), oc)
		Expect(err).NotTo(HaveOccurred())

		oc, err = clients.OpenshiftClusters.Get(ctx, os.Getenv("RESOURCEGROUP"), os.Getenv("CLUSTER"))
		Expect(err).NotTo(HaveOccurred())
		Expect(oc.Tags).To(HaveKeyWithValue("key", &value))
	})
})
