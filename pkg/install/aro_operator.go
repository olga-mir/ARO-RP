package install

// Copyright (c) Microsoft Corporation.
// Licensed under the Apache License 2.0.

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (i *Installer) installAroOperator(ctx context.Context) error {
	i.log.Print("Installing ARO operator resources")

	secret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "GGGG",
			Namespace: "GGGG",
		},
		Type: v1.SecretTypeDockerConfigJson,
	}
	secretData = map[string][]byte{}

	pullSecret, err := pullsecret.SetRegistryProfiles(string(ps.Data[v1.DockerConfigJsonKey]), i.doc.OpenShiftCluster.Properties.RegistryProfiles...)
	if err != nil {
		return err
	}
	_, err = i.kubernetescli.CoreV1().Secrets("GGGG").Create(secret)
	if err != nil {
		return err
	}

	return nil
}
