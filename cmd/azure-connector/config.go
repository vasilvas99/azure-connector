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

	azurecfg "github.com/eclipse-kanto/azure-connector/config"
)

// AzureSettingsExt extends the general configurable data of the azure connector with configuration for the message handlers.
type AzureSettingsExt struct {
	*azurecfg.AzureSettings
	PassthroughCommandTopic    string `json:"passthroughCommandTopic"`
	PassthroughTelemetryTopics string `json:"passthroughTelemetryTopics"`
}

func defaultAzureSettingsExt() *AzureSettingsExt {
	azureSettings := azurecfg.DefaultSettings()
	return &AzureSettingsExt{
		PassthroughCommandTopic:    "cloud-to-device",
		PassthroughTelemetryTopics: "device-to-cloud",
		AzureSettings:              azureSettings,
	}
}

func addMessageHandlerFlags(f *flag.FlagSet, settings *AzureSettingsExt) {
	def := defaultAzureSettingsExt()

	f.StringVar(&settings.PassthroughTelemetryTopics,
		"passthroughTelemetryTopics", def.PassthroughTelemetryTopics,
		"The comma-separated list of passthrough telemetry MQTT topics the azure connector listens to on the local broker",
	)
	f.StringVar(&settings.PassthroughCommandTopic,
		"passthroughCommandTopic", def.PassthroughCommandTopic,
		"The passthrough command MQTT topic where all messages from the cloud are forwarded to on the local broker",
	)
}
