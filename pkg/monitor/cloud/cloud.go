package cloud

// Copyright (c) Microsoft Corporation.
// Licensed under the Apache License 2.0.

import (
	"context"

	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/sirupsen/logrus"

	"github.com/Azure/ARO-RP/pkg/api"
	"github.com/Azure/ARO-RP/pkg/api/validate"
	"github.com/Azure/ARO-RP/pkg/env"
	"github.com/Azure/ARO-RP/pkg/metrics"
	"github.com/Azure/ARO-RP/pkg/util/azureclient/mgmt/authorization"
	"github.com/Azure/ARO-RP/pkg/util/subnet"
)

type Monitor struct {
	env       env.Interface
	log       *logrus.Entry
	hourlyRun bool

	oc   *api.OpenShiftCluster
	dims map[string]string

	m  metrics.Interface
	dv validate.OpenShiftClusterDynamicValidator
}

// NewMonitor creates new Monitor
func NewMonitor(ctx context.Context, env env.Interface, log *logrus.Entry, oc *api.OpenShiftCluster, subscriptionDoc *api.SubscriptionDocument, m metrics.Interface, hourlyRun bool) (*Monitor, error) {
	r, err := azure.ParseResourceID(oc.ID)
	if err != nil {
		return nil, err
	}
	dv, err := validate.NewOpenShiftClusterDynamicValidator(log, env, oc, subscriptionDoc)
	if err != nil {
		return nil, err
	}

	dims := map[string]string{
		"resourceId":     oc.ID,
		"subscriptionId": r.SubscriptionID,
		"resourceGroup":  r.ResourceGroup,
		"resourceName":   r.ResourceName,
	}

	return &Monitor{
		env:       env,
		log:       log,
		hourlyRun: hourlyRun,

		oc:   oc,
		dims: dims,

		m:  m,
		dv: dv,
	}, nil
}

//       api.CloudErrorCodeInvalidLinkedRouteTable:            "invalid_route_table",
//       api.CloudErrorCodeInvalidLinkedVNet:                  "invalid_vnet",
//       api.CloudErrorCodeInvalidResourceProviderPermissions: "invalid_rp_permissions",
//       api.CloudErrorCodeInvalidServicePrincipalCredentials: "invalid_sp_credentials",
//       api.CloudErrorCodeInvalidServicePrincipalPermissions: "invalid_sp_permissions",
//       api.CloudErrorResourceProviderNotRegistered:          "rp_not_registered",

func (mon *Monitor) v1(ctx context.Context) error {
	r, err := azure.ParseResourceID(mon.oc.ID)
	if err != nil {
		return err
	}

	spAuthorizer, err := mon.dv.ValidateServicePrincipalProfile(ctx)
	if err != nil {
		return err
	}

	spPermissions := authorization.NewPermissionsClient(r.SubscriptionID, spAuthorizer)
	// spProviders := features.NewProvidersClient(r.SubscriptionID, spAuthorizer)
	// spUsage := compute.NewUsageClient(r.SubscriptionID, spAuthorizer)
	// spVirtualNetworks := network.NewVirtualNetworksClient(r.SubscriptionID, spAuthorizer)

	vnetID, _, err := subnet.Split(mon.oc.Properties.MasterProfile.SubnetID)
	if err != nil {
		return err
	}

	vnetr, err := azure.ParseResourceID(vnetID)
	if err != nil {
		return err
	}

	return mon.dv.ValidateVnetPermissions(ctx, spAuthorizer, spPermissions, vnetID, &vnetr, api.CloudErrorCodeInvalidServicePrincipalPermissions, "provided service principal")
}

// err = dv.ValidateVnetPermissions(ctx, dv.fpAuthorizer, dv.fpPermissions, vnetID, &vnetr, api.CloudErrorCodeInvalidResourceProviderPermissions, "resource provider")
// func (mon *Monitor) v2(ctx context.Context) error {
// 	return mon.dv.ValidateRouteTablePermissions(ctx, spAuthorizer, dv.spPermissions, &vnet, api.CloudErrorCodeInvalidServicePrincipalPermissions, "provided service principal")
// }

//        err = dv.ValidateRouteTablePermissions(ctx, dv.fpAuthorizer, dv.fpPermissions, &vnet, api.CloudErrorCodeInvalidResourceProviderPermissions, "resource provider")
//        err = dv.ValidateVnet(ctx, &vnet)
//        err = dv.ValidateProviders(ctx)

// Monitor checks the API server health of a cluster
func (mon *Monitor) Monitor(ctx context.Context) {
	mon.log.Debug("monitoring")

	for _, f := range []func(context.Context) error{
		mon.v1,
		// mon.v2,
	} {
		err := f(ctx)
		if err != nil {
			if err, ok := err.(*api.CloudError); ok {
				mon.log.Printf("Found cloud config error: %s", err)
				mon.emitGauge("monitor.clouderrors", 1, map[string]string{"monitor": "bu"})
			}
		}
	}
}

func (mon *Monitor) emitGauge(m string, value int64, dims map[string]string) {
	if dims == nil {
		dims = map[string]string{}
	}
	for k, v := range mon.dims {
		dims[k] = v
	}
	mon.m.EmitGauge(m, value, dims)
}
