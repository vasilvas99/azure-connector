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

	conn "github.com/eclipse-kanto/suite-connector/config"
	"github.com/eclipse-kanto/suite-connector/flags"
	"github.com/eclipse-kanto/suite-connector/logger"

	"github.com/eclipse-kanto/azure-connector/config"
)

const (
	flagCACert           = "cacert"
	flagTenantID         = "tenantId"
	flagIDScope          = "idScope"
	flagSASTokenValidity = "sasTokenValidity"
)

// AddGlobal adds the Cloud Agent global flags.
func AddGlobal(f *flag.FlagSet) (configFile *string) {
	return flags.AddGlobal(f)
}

// ConfigCheck checks for config file existence.
func ConfigCheck(logger logger.Logger, configFile string) {
	flags.ConfigCheck(logger, configFile)
}

// Add adds the Cloud Agent flags and uses the provided settings to collect the provided values.
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
		"The validity period for the generated SAS token for device authentication",
	)
	f.StringVar(&settings.AllowedLocalTopicsList,
		"allowedLocalTopicsList", def.AllowedLocalTopicsList,
		"List of allowed Local topics",
	)
	f.StringVar(&settings.AllowedCloudMessageTypesList,
		"allowedCloudMessageTypesList", def.AllowedCloudMessageTypesList,
		"List of allowed cloud message types",
	)
	f.StringVar(&settings.IDScope, flagIDScope, def.IDScope,
		"ID Scope from Azure Device Provisioning Service",
	)

	AddHub(f, settings, def)
	flags.AddLocalBroker(f, &settings.LocalConnectionSettings, &def.LocalConnectionSettings)
	flags.AddLog(f, &settings.LogSettings, &def.LogSettings)

	AddTLS(f, &settings.TLSSettings, &def.TLSSettings)
}

// AddTLS adds cacert, cert and key flags only as minimized TLS connection settings.
func AddTLS(f *flag.FlagSet, settings, def *conn.TLSSettings) {

	f.StringVar(&settings.CACert,
		flagCACert, def.CACert,
		"A PEM encoded CA certificates `file`",
	)
	f.StringVar(&settings.Cert,
		"cert", def.Cert,
		"A PEM encoded certificate `file` for cloud access",
	)
	f.StringVar(&settings.Key,
		"key", def.Key,
		"A PEM encoded unencrypted private key `file` for cloud access",
	)
}

// Copy configured all set flag values to map.
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
