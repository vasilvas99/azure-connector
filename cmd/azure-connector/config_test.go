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

package main

import (
	"flag"
	"testing"

	azurecfg "github.com/eclipse-kanto/azure-connector/config"
	"github.com/eclipse-kanto/azure-connector/flags"
	"github.com/eclipse-kanto/suite-connector/config"

	"github.com/stretchr/testify/assert"
)

const (
	testConfig = "../../config/testdata/config.json"
)

func TestConfigFile(t *testing.T) {
	settings := defaultAzureSettingsExt()
	assert.NoError(t, config.ReadConfig(testConfig, settings))
	assert.Equal(t, "from-device-to-cloud", settings.PassthroughTelemetryTopics)
	assert.Equal(t, "from-cloud-to-device", settings.PassthroughCommandTopic)
}

func TestDefaults(t *testing.T) {
	settings := defaultAzureSettingsExt()
	assert.NoError(t, settings.Validate())
	assert.Equal(t, "device-to-cloud", settings.PassthroughTelemetryTopics)
	assert.Equal(t, "cloud-to-device", settings.PassthroughCommandTopic)
	assert.Equal(t, azurecfg.DefaultSettings(), settings.AzureSettings)
}

func TestFlagsSet(t *testing.T) {
	f := flag.NewFlagSet("testing", flag.ContinueOnError)
	cmd := defaultAzureSettingsExt()
	flags.Add(f, cmd.AzureSettings)
	flags.AddGlobal(f)
	addMessageHandlerFlags(f, cmd)

	assert.NotNil(t, f.Lookup("passthroughTelemetryTopics"))
	assert.NotNil(t, f.Lookup("passthroughCommandTopic"))
}
