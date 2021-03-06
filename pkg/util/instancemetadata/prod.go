package instancemetadata

// Copyright (c) Microsoft Corporation.
// Licensed under the Apache License 2.0.

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/Azure/go-autorest/autorest/adal"
	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/dgrijalva/jwt-go"
)

type azureClaim struct {
	TenantID string `json:"tid,omitempty"`
}

func (*azureClaim) Valid() error {
	return fmt.Errorf("unimplemented")
}

func NewProd() (InstanceMetadata, error) {
	im := &instanceMetadata{}

	err := im.populateTenantIDFromMSI()
	if err != nil {
		return nil, err
	}

	err = im.populateInstanceMetadata()
	if err != nil {
		return nil, err
	}

	return im, nil
}

func (im *instanceMetadata) populateTenantIDFromMSI() error {
	msiEndpoint, err := adal.GetMSIVMEndpoint()
	if err != nil {
		return err
	}

	token, err := adal.NewServicePrincipalTokenFromMSI(msiEndpoint, azure.PublicCloud.ResourceManagerEndpoint)
	if err != nil {
		return err
	}

	err = token.EnsureFresh()
	if err != nil {
		return err
	}

	p := &jwt.Parser{}
	c := &azureClaim{}
	_, _, err = p.ParseUnverified(token.OAuthToken(), c)
	if err != nil {
		return err
	}

	im.tenantID = c.TenantID

	return nil
}

func (im *instanceMetadata) populateInstanceMetadata() error {
	req, err := http.NewRequest(http.MethodGet, "http://169.254.169.254/metadata/instance/compute?api-version=2019-03-11", nil)
	if err != nil {
		return err
	}
	req.Header.Set("Metadata", "true")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code %q", resp.StatusCode)
	}

	if strings.SplitN(resp.Header.Get("Content-Type"), ";", 2)[0] != "application/json" {
		return fmt.Errorf("unexpected content type %q", resp.Header.Get("Content-Type"))
	}

	var m *struct {
		Location          string `json:"location,omitempty"`
		ResourceGroupName string `json:"resourceGroupName,omitempty"`
		SubscriptionID    string `json:"subscriptionId,omitempty"`
	}

	err = json.NewDecoder(resp.Body).Decode(&m)
	if err != nil {
		return err
	}

	im.subscriptionID = m.SubscriptionID
	im.location = m.Location
	im.resourceGroup = m.ResourceGroupName

	return nil
}
