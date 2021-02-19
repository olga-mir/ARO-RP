package credentialsmanager

// Copyright (c) Microsoft Corporation.
// Licensed under the Apache License 2.0.

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"fmt"
	"time"

	"github.com/Azure/ARO-RP/pkg/util/keyvault"
)

type credentialsmanager struct {
	keyvault    keyvault.Manager
	secretName  string
	certificate *x509.Certificate
	privateKey  *rsa.PrivateKey

	lastFetched     time.Time
	refreshInterval time.Duration

	// if the credentials are older than the maxAge, consider them invalid
	maxAge time.Duration
}

type Manager interface {
	Get() (*x509.Certificate, *rsa.PrivateKey, error)
}

// NewManager returns new Manager
func NewManager(kv keyvault.Manager, secretName string, refreshInterval time.Duration, maxAge time.Duration) (Manager, error) {
	manager := credentialsmanager{
		keyvault:        kv,
		secretName:      secretName,
		lastFetched:     time.Time{},
		refreshInterval: refreshInterval,
		maxAge:          maxAge,
	}

	err := fetch(&manager)
	if err != nil {
		return nil, err
	}
	manager.lastFetched = time.Now()
	ticker := time.NewTicker(manager.refreshInterval)
	stop := make(chan struct{})
	go func() {
		for {
			select {
			case <-ticker.C:
				err := fetch(&manager)
				if err != nil {
					return
				}
				manager.lastFetched = time.Now()
			case <-stop:
				ticker.Stop()
				return
			}
		}
	}()

	return manager, nil
}

func fetch(m *credentialsmanager) error {
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	privateKey, certs, err := m.keyvault.GetCertificateSecret(ctx, m.secretName)
	if err != nil {
		return err
	}

	m.privateKey = privateKey
	m.certificate = certs[0]

	return nil
}

func (m credentialsmanager) Get() (*x509.Certificate, *rsa.PrivateKey, error) {
	if m.certificate == nil || m.privateKey == nil {
		return nil, nil, fmt.Errorf("Certificate or PrivateKey does not exist")
	}
	now := time.Now()
	if now.Sub(m.lastFetched) > m.maxAge {
		// returning error here means that instantiating FPAuthorizer will fail and consequently
		// the task will fail, it might be picked up later by another instance of RP.
		// This is grey failure which we may not notice, which is not great.
		// In the longet term we probably want to implement a metric and alert when the credentials are stale.
		return nil, nil, fmt.Errorf("Certificate or PrivateKey failed to fetch in last 24 hours")
	}
	if false {
		return nil, nil, fmt.Errorf("No valid credentials available")
	}

	return m.certificate, m.privateKey, nil
}
