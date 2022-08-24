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

package flags_test

import (
	"flag"
	"os"
	"testing"

	"github.com/eclipse-kanto/azure-connector/config"
	"github.com/eclipse-kanto/azure-connector/flags"
	"github.com/eclipse-kanto/suite-connector/logger"
	"github.com/eclipse-kanto/suite-connector/testutil"

	"github.com/imdario/mergo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFlagsMappings(t *testing.T) {
	l := testutil.NewLogger("flags", logger.INFO, t)

	f := flag.NewFlagSet("testing", flag.ContinueOnError)

	cmd := new(config.AzureSettings)
	flags.Add(f, cmd)
	configFile := flags.AddGlobal(f)

	args := []string{
		"-configFile=config.json",
		"-tenantId=tenant7172",
		"-connectionString=HostName=dummy-hub.azure-devices.net;DeviceId=dummy-device;SharedAccessKey=ZHVtbXk=",
		"-messageMapperConfig=messageMapperConfig.json",
		"-allowedLocalTopicsList=local",
		"-allowedCloudMessageTypesList=cloud",
		"-sasTokenValidity=2h",
		"-idScope=dummyIdScope",
		"-localAddress=tcp://localhost:1883",
		"-localUsername=user",
		"-localPassword=pass",
		"-cacert=cacert.crt",
		"-cert=cert.crt",
		"-key=key.crt",
		"-logFile=log.txt",
		"-logLevel=TRACE",
		"-logFileSize=10",
		"-logFileCount=100",
		"-logFileMaxAge=1000",
	}

	require.NoError(t, flags.Parse(f, args, "0.0.0", os.Exit))
	assert.Equal(t, "config.json", *configFile)
	assert.Equal(t, "tenant7172", cmd.TenantID)
	assert.Equal(t, "HostName=dummy-hub.azure-devices.net;DeviceId=dummy-device;SharedAccessKey=ZHVtbXk=", cmd.ConnectionString)
	assert.Equal(t, "messageMapperConfig.json", cmd.MessageMapperConfig)
	assert.Equal(t, "local", cmd.AllowedLocalTopicsList)
	assert.Equal(t, "cloud", cmd.AllowedCloudMessageTypesList)
	assert.Equal(t, "2h", cmd.SASTokenValidity)
	assert.Equal(t, "dummyIdScope", cmd.IDScope)
	assert.Equal(t, "tcp://localhost:1883", cmd.LocalAddress)
	assert.Equal(t, "user", cmd.LocalUsername)
	assert.Equal(t, "pass", cmd.LocalPassword)
	assert.Equal(t, "cacert.crt", cmd.CACert)
	assert.Equal(t, "cert.crt", cmd.Cert)
	assert.Equal(t, "key.crt", cmd.Key)
	assert.EqualValues(t, "log.txt", cmd.LogFile)
	assert.EqualValues(t, logger.TRACE, cmd.LogLevel)
	assert.EqualValues(t, 10, cmd.LogFileSize)
	assert.EqualValues(t, 100, cmd.LogFileCount)
	assert.EqualValues(t, 1000, cmd.LogFileMaxAge)

	flags.ConfigCheck(l, *configFile)

	m := flags.Copy(f)

	cp := config.DefaultSettings()
	if err := mergo.Map(cp, m, mergo.WithOverwriteWithEmptyValue); err != nil {
		require.NoError(t, err)
	}

	assert.Equal(t, cmd, cp)
}
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
		"messageMapperConfig",
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
}

func assertFlagExists(t *testing.T, flagName string, f *flag.FlagSet) {
	flg := f.Lookup(flagName)
	assert.NotNil(t, flg)
}
