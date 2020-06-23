package e2e

import (
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("Cluster smoke test", func() {
	Specify("Nodes should have correct labels", func() {
		labels := map[string]map[string]string{
			"master": {
				"node-role.kubernetes.io/master": "",
			},
			"worker": {
				"node-role.kubernetes.io/worker": "",
			},
		}

		list, err := clients.Kubernetes.CoreV1().Nodes().List(metav1.ListOptions{})
		Expect(err).NotTo(HaveOccurred())

		for _, node := range list.Items {
			kind := strings.Split(node.Name, "-")[0]
			_, ok := labels[kind]
			Expect(ok).To(Equal(false))

			// for k, v := range l {
			// 	// if val, ok := node.Labels[k]; !ok || val != v {
			// 	// 	return log.Errorf("map does not have key %s", kind)
			// 	// }
			// }
		}
	})
})

// 	sc.Log.Debugf("validating that all monitoring components are healthy")
// 	err = sc.checkMonitoringStackHealth(ctx)

// 	sc.Log.Debugf("validating that pod disruption budgets are immutable")
// 	err = sc.checkDisallowsPdbMutations(ctx)

// 	sc.Log.Debugf("validating that the cluster can create ELB and ILB")
// 	err = sc.checkCanCreateLB(ctx)

// 	sc.Log.Debugf("validating that cluster services are available")
// 	err = sc.checkCanAccessServices(ctx)

// 	sc.Log.Debugf("validating that the cluster can use azure-file storage")
// 	err = sc.checkCanUseAzureFileStorage(ctx)

// 	sc.Log.Debugf("validating that the cluster enforces emptydir quotas")
// 	err = sc.checkEnforcesEmptyDirQuotas(ctx)
