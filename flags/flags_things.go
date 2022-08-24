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

//go:build things
// +build things

package flags

import (
	"flag"

	"github.com/eclipse-kanto/azure-connector/config"
)

// AddHub adds the Hub connection related flags.
func AddHub(f *flag.FlagSet, settings, def *config.AzureSettings) {
	f.StringVar(&settings.MessageMapperConfig,
		"messageMapperConfig", def.MessageMapperConfig,
		"The path to the configuration file for the message mappings",
	)
}
