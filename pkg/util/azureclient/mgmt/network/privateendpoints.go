package network

// Copyright (c) Microsoft Corporation.
// Licensed under the Apache License 2.0.

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/services/network/mgmt/2019-07-01/network"
	"github.com/Azure/go-autorest/autorest"
)

// PrivateEndpointsClient is a minimal interface for azure PrivateEndpointsClient
type PrivateEndpointsClient interface {
	Get(ctx context.Context, resourceGroupName string, privateEndpointName string, expand string) (result network.PrivateEndpoint, err error)
	PrivateEndpointsClientAddons
}

type privateEndpointsClient struct {
	network.PrivateEndpointsClient
}

var _ PrivateEndpointsClient = &privateEndpointsClient{}

// NewPrivateEndpointsClient creates a new PrivateEndpointsClient
func NewPrivateEndpointsClient(subscriptionID string, authorizer autorest.Authorizer) PrivateEndpointsClient {
	client := network.NewPrivateEndpointsClient(subscriptionID)
	client.Authorizer = authorizer

	return &privateEndpointsClient{
		PrivateEndpointsClient: client,
	}
}
