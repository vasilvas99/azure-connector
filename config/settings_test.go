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

package config

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/eclipse-kanto/suite-connector/config"
	"github.com/eclipse-kanto/suite-connector/logger"
	"github.com/eclipse-kanto/suite-connector/util"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testConfig = "testdata/config.json"
)

func TestConfigEmpty(t *testing.T) {
	configPath := "configEmpty.json"

	err := ioutil.WriteFile(configPath, []byte{}, os.ModePerm)
	require.NoError(t, err)
	defer func() {
		os.Remove(configPath)
	}()
	require.True(t, util.FileExists(configPath))

	cmd := new(AzureSettings)
	assert.NoError(t, config.ReadConfig(configPath, cmd))
}

func TestConfigInvalid(t *testing.T) {
	settings := DefaultSettings()
	assert.Error(t, config.ReadConfig("settings_test.go", settings))
	assert.Error(t, config.ReadConfig("settings_test.go", nil))

	settings.LogFileCount = 0
	assert.Error(t, settings.Validate())

	settings.LogFileCount = 1
	settings.LocalAddress = ""
	assert.Error(t, settings.Validate())
}

func TestConfig(t *testing.T) {
	expSettings := DefaultSettings()
	expSettings.TenantID = "tenant7172"
	expSettings.PassthroughTelemetryTopics = "from-device-to-cloud"
	expSettings.PassthroughCommandTopic = "from-cloud-to-device"
	expSettings.LocalUsername = "localUsername_config"
	expSettings.LocalPassword = "localPassword_config"
	expSettings.CACert = ""
	expSettings.LogFile = "logFile_config"
	expSettings.LogLevel = logger.DEBUG

	settings := DefaultSettings()
	require.NoError(t, config.ReadConfig(testConfig, settings))
	assert.Equal(t, expSettings, settings)

	settings.TenantID = "tenantId_config"
	assert.NoError(t, settings.Validate())
}

func TestDefaults(t *testing.T) {
	settings := DefaultSettings()
	assert.Error(t, settings.Validate())

	assert.Equal(t, "defaultTenant", settings.TenantID)
	assert.Empty(t, settings.ConnectionString)
	assert.Equal(t, "1h", settings.SASTokenValidity)
	assert.Equal(t, "device-to-cloud", settings.PassthroughTelemetryTopics)
	assert.Equal(t, "cloud-to-device", settings.PassthroughCommandTopic)
	assert.Empty(t, settings.IDScope)

	defConnectorSettings := config.DefaultSettings()
	assert.Equal(t, defConnectorSettings.LocalConnectionSettings, settings.LocalConnectionSettings)

	defTLSSettings := config.TLSSettings{
		CACert: defConnectorSettings.CACert,
	}
	assert.Equal(t, defTLSSettings, settings.TLSSettings)

	defLogSettings := defConnectorSettings.LogSettings
	defLogSettings.LogFile = "logs/azure-connector.log"
	defLogSettings.LogFileMaxAge = 28
	assert.Equal(t, defLogSettings, settings.LogSettings)
}
