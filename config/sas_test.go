// Copyright (c) 2022 Contributors to the Eclipse Foundation
//
// See the NOTICE file(s) distributed with this work for additional
// information regarding copyright ownership.
//
// This program and the accompanying materials are made available under the
// terms of the Eclipse Public License 2.0 which is available at
// https://www.eclipse.org/legal/epl-2.0, or the Apache License, Version 2.0
// which is available at https://www.apache.org/licenses/LICENSE-2.0.
//
// SPDX-License-Identifier: EPL-2.0 OR Apache-2.0

package config_test

import (
	"io"
	"log"
	"testing"
	"time"

	"github.com/eclipse-kanto/azure-connector/config"
	"github.com/eclipse-kanto/suite-connector/logger"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateSASToken(t *testing.T) {
	config.Now = func() time.Time {
		return time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)
	}
	connectionString := "HostName=dummy-hub.azure-devices.net;DeviceId=dummy-device;SharedAccessKey=cGFzc3dvcmQ="

	settings := &config.AzureSettings{ConnectionString: connectionString}
	logger := logger.NewLogger(log.New(io.Discard, "", log.Ldate), logger.INFO)

	connSettings, err := config.PrepareAzureConnectionSettings(settings, nil, logger)
	require.NoError(t, err)

	sasToken := config.GenerateSASToken(connSettings)

	assert.Equal(t, "dummy-hub.azure-devices.net", sasToken.Sr)
	assert.Equal(t, "2021-01-01 01:00:00 +0000 UTC", sasToken.Se.String())
	assert.Equal(t, "ifZm2I0YKRkwc8Pc49e0qKSsu3l3FbxoWZRqGBtXtng=", sasToken.Sig)
}

func TestParseSASTokenValidity(t *testing.T) {
	testData := []struct {
		testName              string
		tokenValidity         string
		expectedTokenValidity time.Duration
	}{
		{
			"empty_token_validity",
			"",
			-1,
		},
		{
			"zero_period",
			"0h",
			-1,
		},
		{
			"missing_period_type",
			"1",
			-1,
		},
		{
			"missing_period",
			"h",
			-1,
		},
		{
			"negative_period",
			"-1h",
			-1,
		},
		{
			"unsupported_period_type",
			"1w",
			-1,
		},
		{
			"invalid_period",
			"1oh",
			-1,
		},
		{
			"minute_period",
			"1m",
			time.Minute,
		},
		{
			"minutes_period",
			"30m",
			time.Duration(30) * time.Minute,
		},
		{
			"hour_period",
			"1h",
			time.Hour,
		},
		{
			"hours_period",
			"2h",
			time.Duration(2) * time.Hour,
		},
		{
			"day_period",
			"1d",
			time.Duration(24) * time.Hour,
		},
		{
			"days_period",
			"7d",
			7 * time.Duration(24) * time.Hour,
		},
	}
	for _, testValues := range testData {
		t.Run(testValues.testName, func(t *testing.T) {
			tokenValidity, err := config.ParseSASTokenValidity(testValues.tokenValidity)
			if testValues.expectedTokenValidity == -1 {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, testValues.expectedTokenValidity, tokenValidity)
			}
		})
	}
}
