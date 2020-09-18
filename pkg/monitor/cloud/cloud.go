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
	"github.com/Azure/ARO-RP/pkg/util/subnet"
)

type Monitor struct {
	env       env.Interface
	log       *logrus.Entry
	hourlyRun bool

	oc       *api.OpenShiftCluster
	dims     map[string]string
	resource azure.Resource

	m  metrics.Interface
	dv validate.OpenShiftClusterDynamicValidator

	// spAuthorizer refreshable.Authorizer
	// fpAuthorizer refreshable.Authorizer
	// fpPermissions     authorization.PermissionsClient
	// spPermissions     authorization.PermissionsClient
	// spProviders       features.ProvidersClient
	// spUsage           compute.UsageClient
	// spVirtualNetworks network.VirtualNetworksClient
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

		oc:       oc,
		dims:     dims,
		resource: r,

		m:  m,
		dv: dv,
	}, nil
}

var errorMap = map[string]string{
	api.CloudErrorCodeInvalidLinkedRouteTable:            "invalid_route_table",
	api.CloudErrorCodeInvalidLinkedVNet:                  "invalid_vnet",
	api.CloudErrorCodeInvalidResourceProviderPermissions: "invalid_rp_permissions",
	api.CloudErrorCodeInvalidServicePrincipalCredentials: "invalid_sp_credentials",
	api.CloudErrorCodeInvalidServicePrincipalPermissions: "invalid_sp_permissions",
	api.CloudErrorResourceProviderNotRegistered:          "rp_not_registered",
}

func (mon *Monitor) spValidateVnetPermissions(ctx context.Context) error {
	// TODO - vnetID and vnetr to setup as properties to avoid duplication
	vnetID, _, err := subnet.Split(mon.oc.Properties.MasterProfile.SubnetID)
	if err != nil {
		return err
	}

	vnetr, err := azure.ParseResourceID(vnetID)
	if err != nil {
		return err
	}

	return mon.dv.ValidateVnetPermissions(ctx, vnetID, &vnetr, "service principal")
}

func (mon *Monitor) rpValidateVnetPermissions(ctx context.Context) error {
	vnetID, _, err := subnet.Split(mon.oc.Properties.MasterProfile.SubnetID)
	if err != nil {
		return err
	}

	vnetr, err := azure.ParseResourceID(vnetID)
	if err != nil {
		return err
	}

	return mon.dv.ValidateVnetPermissions(ctx, vnetID, &vnetr, "resource provider")
}

func (mon *Monitor) spValidateRouteTablePermissions(ctx context.Context) error {
	return nil
	// vnet, err := dv.spVirtualNetworks.Get(ctx, vnetr.ResourceGroup, vnetr.ResourceName, "")
	// if err != nil {
	// 	return err
	// }
	// return mon.dv.ValidateRouteTablePermissions(ctx, &vnet, "service principal")
}

//        err = dv.ValidateRouteTablePermissions(ctx, dv.fpAuthorizer, dv.fpPermissions, &vnet, api.CloudErrorCodeInvalidResourceProviderPermissions, "resource provider")
//        err = dv.ValidateVnet(ctx, &vnet)
//        err = dv.ValidateProviders(ctx)

// Monitor checks various misconfigurations in cloud infrastructure
func (mon *Monitor) Monitor(ctx context.Context) {
	mon.log.Debug("monitoring")

	err := mon.dv.Setup(ctx)
	mon.reportError(err)
	// probably we need to split checks in rp and sp series. Abort the line if the permissions are missing?

	for _, f := range []func(context.Context) error{
		mon.spValidateVnetPermissions,
		mon.rpValidateVnetPermissions,
		mon.spValidateRouteTablePermissions,
	} {
		mon.reportError(f(ctx))
	}
}

func (mon *Monitor) reportError(err error) {
	if err != nil {
		if err, ok := err.(*api.CloudError); ok {
			mon.log.Printf("Found cloud config error: %s", err)
			mon.emitGauge("monitor.clouderrors", 1, map[string]string{"monitor": errorMap[err.CloudErrorBody.Code]})
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
