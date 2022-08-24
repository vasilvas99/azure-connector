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

package config_test

import (
	"encoding/base64"
	"io"
	"log"
	"testing"

	"go.uber.org/goleak"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/eclipse-kanto/azure-connector/config"
	"github.com/eclipse-kanto/suite-connector/logger"
)

func TestCreateAzureClientNoCacert(t *testing.T) {
	defer goleak.VerifyNone(t)

	decodedAccessKey, _ := base64.StdEncoding.DecodeString("x7HrdC+URzEneFam9ZKa0Ke7nvsDwiuJptzFkgs8JWA=")
	accessKey := &config.SharedAccessKey{
		SharedAccessKeyDecoded: decodedAccessKey,
	}
	settings := &config.AzureSettings{}
	connSettings := &config.AzureConnectionSettings{
		DeviceID:        "dummy-device",
		HostName:        "dummy-hub.azure-devices.net",
		HubName:         "dummy-hub",
		SharedAccessKey: accessKey,
	}

	logger := logger.NewLogger(log.New(io.Discard, "", log.Ldate), logger.INFO)
	azureClient, err := config.CreateAzureHubConnection(settings, connSettings, logger)
	require.Error(t, err)
	assert.Nil(t, azureClient)
}
