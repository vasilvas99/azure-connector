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

//go:build !things
// +build !things

package flags_test

import (
	"flag"
	"testing"

	"github.com/eclipse-kanto/azure-connector/config"
	"github.com/eclipse-kanto/azure-connector/flags"

	"github.com/stretchr/testify/assert"
)

func TestFlagsSet(t *testing.T) {
	f := flag.NewFlagSet("testing", flag.ContinueOnError)
	cmd := new(config.AzureSettings)
	flags.Add(f, cmd)
	flags.AddGlobal(f)

	flagNames := []string{
		"configFile",
		"tenantId",
		"connectionString",
		"allowedLocalTopicsList",
		"allowedCloudMessageTypesList",
		"sasTokenValidity",
		"idScope",
		"localAddress",
		"localUsername",
		"localPassword",
		"cacert",
		"cert",
		"key",
		"logFile",
		"logLevel",
		"logFileSize",
		"logFileCount",
		"logFileMaxAge",
	}
	for _, flagName := range flagNames {
		assertFlagExists(t, flagName, f)
	}
	assertFlagNotExists(t, "messageMapperConfig", f)
}

func assertFlagExists(t *testing.T, flagName string, f *flag.FlagSet) {
	flg := f.Lookup(flagName)
	assert.NotNil(t, flg)
}

func assertFlagNotExists(t *testing.T, flagName string, f *flag.FlagSet) {
	flg := f.Lookup(flagName)
	assert.Nil(t, flg)
}
