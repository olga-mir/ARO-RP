package project

import (

	// authorizationv1 "k8s.io/api/authorization/v1"

	"time"

	projectv1 "github.com/openshift/api/project/v1"
	projectv1client "github.com/openshift/client-go/project/clientset/versioned/typed/project/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
)

// func (cli *Client) CreateProject(namespace string) error {
// 	_, err := cli.ProjectV1.ProjectRequests().Create(&projectv1.ProjectRequest{
// 		ObjectMeta: metav1.ObjectMeta{
// 			Name: namespace,
// 		},
// 	})

// CreateProject creates a new project and validates it
func CreateProject(projectV1Client projectv1client.ProjectV1Interface, name string) error {
	_, err := projectV1Client.Projects().Create(&projectv1.Project{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	})
	if err != nil {
		return err
	}

	// err = wait.PollImmediate(time.Second, time.Minute, func() (bool, error) {
	// 	res, err := cli.AuthorizationV1.SelfSubjectAccessReviews().Create(
	// 		&authorizationv1.SelfSubjectAccessReview{
	// 			Spec: authorizationv1.SelfSubjectAccessReviewSpec{
	// 				ResourceAttributes: &authorizationv1.ResourceAttributes{
	// 					Namespace: namespace,
	// 					Verb:      "create",
	// 					Resource:  "pods",
	// 				},
	// 			},
	// 		},
	// 	)
	// 	if err != nil {
	// 		return false, err
	// 	}
	// 	return res.Status.Allowed, nil
	// })
	// if err != nil {
	// 	return fmt.Errorf("failed to wait for self-sar: %v", err)
	// }

	// err = wait.PollImmediate(time.Second, time.Minute, func() (bool, error) {
	// 	sa, err := cli.CoreV1.ServiceAccounts(namespace).Get("default", metav1.GetOptions{})
	// 	if errors.IsNotFound(err) {
	// 		return false, nil
	// 	}
	// 	if err != nil {
	// 		return false, err
	// 	}
	// 	return len(sa.Secrets) > 0, nil
	// })
	// if err != nil {
	// 	return fmt.Errorf("failed to wait for default service account: %v", err)
	// }

	// err = wait.PollImmediate(time.Second, time.Minute, func() (bool, error) {
	// 	project, err := cli.ProjectV1.Projects().Get(namespace, metav1.GetOptions{})
	// 	if err != nil {
	// 		return false, err
	// 	}
	// 	_, found := project.Annotations["openshift.io/sa.scc.uid-range"]
	// 	return found, nil
	// })
	// if err != nil {
	// 	return fmt.Errorf("failed to wait for scc: %v", err)
	// }

	return nil
}

// CleanupProject deletes a project and validates its deletion
func CleanupProject(projectV1Client projectv1client.ProjectV1Interface, name string) error {
	err := projectV1Client.Projects().Delete(name, &metav1.DeleteOptions{})
	if err != nil {
		return err
	}

	return wait.PollImmediate(2*time.Second, 10*time.Minute, func() (bool, error) {
		_, err := projectV1Client.Projects().Get(name, metav1.GetOptions{})
		if errors.IsNotFound(err) || errors.IsForbidden(err) {
			return true, nil
		}
		return false, err
	})
}
