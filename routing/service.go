// Copyright (c) 2022 Contributors to the Eclipse Foundation
//
// See the NOTICE file(s) distributed with this work for additional
// information regarding copyright ownership.
//
// This program and the accompanying materials are made available under the
// terms of the Eclipse Public License 2.0 which is available at
// http://www.eclipse.org/legal/epl-2.0
//
// SPDX-License-Identifier: EPL-2.0

package routing

import (
	"fmt"

	"github.com/eclipse-kanto/suite-connector/routing"
)

const (
	// StatusConnectionNotAuthorized defines a not authorized connection status.
	StatusConnectionNotAuthorized = "CONNECTION_NOT_AUTHORIZED"
	// StatusConnectionTokenExpired defines a token expired connection status.
	StatusConnectionTokenExpired = "CONNECTION_TOKEN_EXPIRED"
)

const (
	defaultTenantID = "defaultTenant"
	dittoNamespace  = "azure.edge"
)

// NewAzureGwParams creates the gateway parameters
func NewAzureGwParams(deviceID, tenantID, hubName string) *routing.GwParams {
	if len(tenantID) == 0 {
		tenantID = defaultTenantID
	}

	return &routing.GwParams{
		DeviceID: fmt.Sprintf("%s:%s:%s", dittoNamespace, hubName, deviceID),
		TenantID: tenantID,
		PolicyID: "",
	}
}
