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

package routing_test

import (
	"testing"

	"github.com/ThreeDotsLabs/watermill/pubsub/gochannel"

	"github.com/eclipse-kanto/suite-connector/logger"
	"github.com/eclipse-kanto/suite-connector/routing"
	"github.com/eclipse-kanto/suite-connector/testutil"

	azurerouting "github.com/eclipse-kanto/azure-connector/routing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testDeviceID = "dummy-device"
	testTenantID = "dummy-tenant"
	testHubName  = "dummy-hub"
)

func TestAzureGwParamsCustomTenantId(t *testing.T) {
	gwParams := azurerouting.NewAzureGwParams(testDeviceID, testTenantID, testHubName)
	assert.Equal(t, "azure.edge:dummy-hub:dummy-device", gwParams.DeviceID)
	assert.Equal(t, "dummy-tenant", gwParams.TenantID)
	assert.Equal(t, "", gwParams.PolicyID)
}

func TestAzureGwParamsDefaultTenantId(t *testing.T) {
	gwParams := azurerouting.NewAzureGwParams(testDeviceID, "", testHubName)
	assert.Equal(t, "azure.edge:dummy-hub:dummy-device", gwParams.DeviceID)
	assert.Equal(t, "defaultTenant", gwParams.TenantID)
	assert.Equal(t, "", gwParams.PolicyID)
}

func TestPublishError(t *testing.T) {
	logger := testutil.NewLogger("routing", logger.ERROR, t)

	sink := gochannel.NewGoChannel(
		gochannel.Config{
			Persistent:          true,
			OutputChannelBuffer: int64(1),
		},
		logger,
	)
	require.NoError(t, sink.Close())

	routing.SendStatus(routing.StatusConfigError, sink, logger)

	params := azurerouting.NewAzureGwParams("deviceId", "tenantId", "hubName")
	routing.SendGwParams(params, true, sink, logger)
}
