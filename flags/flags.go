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

package flags

import (
	"flag"

	"github.com/eclipse-kanto/suite-connector/flags"
	"github.com/eclipse-kanto/suite-connector/logger"

	"github.com/eclipse-kanto/azure-connector/config"
)

const (
	flagCACert           = "caCert"
	flagTenantID         = "tenantId"
	flagIDScope          = "idScope"
	flagSASTokenValidity = "sasTokenValidity"
)

// AddGlobal adds the azure connector global flags.
func AddGlobal(f *flag.FlagSet) (configFile *string) {
	return flags.AddGlobal(f)
}

// ConfigCheck checks for config file existence.
func ConfigCheck(logger logger.Logger, configFile string) {
	flags.ConfigCheck(logger, configFile)
}

// Add adds the azure connector flags and uses the given settings structure to store the values of the flags.
func Add(f *flag.FlagSet, settings *config.AzureSettings) {
	def := config.DefaultSettings()

	f.StringVar(&settings.TenantID,
		flagTenantID, def.TenantID,
		"Tenant ID",
	)
	f.StringVar(&settings.ConnectionString,
		"connectionString", def.ConnectionString,
		"The connection string for connectivity to Azure IoT Hub",
	)
	f.StringVar(&settings.SASTokenValidity,
		flagSASTokenValidity, def.SASTokenValidity,
		"The validity period for the generated SAS token for device authentication. Should be a positive integer number followed by a unit suffix, such as '300m', '1h', etc. Valid time units are 'm' (minutes), 'h' (hours), 'd' (days)",
	)
	f.StringVar(&settings.PassthroughTelemetryTopics,
		"passthroughTelemetryTopics", def.PassthroughTelemetryTopics,
		"The comma-separated list of passthrough telemetry MQTT topics the azure connector listens to on the local broker",
	)
	f.StringVar(&settings.PassthroughCommandTopic,
		"passthroughCommandTopic", def.PassthroughCommandTopic,
		"The passthrough command MQTT topic where all messages from the cloud are forwarded to on the local broker",
	)
	f.StringVar(&settings.IDScope, flagIDScope, def.IDScope,
		"ID scope for Azure Device Provisioning service",
	)

	flags.AddLocalBroker(f, &settings.LocalConnectionSettings, &def.LocalConnectionSettings)
	flags.AddLog(f, &settings.LogSettings, &def.LogSettings)

	flags.AddTLS(f, &settings.TLSSettings, &def.TLSSettings)
}

// Copy copies all configured flag values to map.
func Copy(f *flag.FlagSet) map[string]interface{} {
	m := make(map[string]interface{}, f.NFlag())

	f.Visit(func(f *flag.Flag) {
		name := f.Name
		getter := f.Value.(flag.Getter)

		if name == flagCACert {
			name = "CACert"
		} else if name == flagSASTokenValidity {
			name = "SASTokenValidity"
		} else if name == flagTenantID {
			name = "TenantID"
		} else if name == flagIDScope {
			name = "IDScope"
		}

		m[name] = getter.Get()
	})

	return m
}

// Parse invokes flagset parse and processes the version
func Parse(f *flag.FlagSet, args []string, version string, exit func(code int)) error {
	return flags.Parse(f, args, version, exit)
}
