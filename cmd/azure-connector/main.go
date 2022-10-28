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
	"log"
	"os"

	"github.com/imdario/mergo"
	"github.com/pkg/errors"

	"github.com/eclipse-kanto/suite-connector/config"
	"github.com/eclipse-kanto/suite-connector/logger"

	"github.com/eclipse-kanto/azure-connector/cmd/azure-connector/app"
	azurecfg "github.com/eclipse-kanto/azure-connector/config"
	"github.com/eclipse-kanto/azure-connector/flags"
	"github.com/eclipse-kanto/azure-connector/routing/message/handlers"
	"github.com/eclipse-kanto/azure-connector/routing/message/handlers/passthrough"
)

var (
	version = "development"
)

func main() {
	f := flag.NewFlagSet("azure-connector", flag.ContinueOnError)

	cmd := new(azurecfg.AzureSettings)
	flags.Add(f, cmd)
	fConfigFile := flags.AddGlobal(f)

	if err := flags.Parse(f, os.Args[1:], version, os.Exit); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			os.Exit(0)
		} else {
			os.Exit(2)
		}
	}

	settings := azurecfg.DefaultSettings()
	if err := config.ReadConfig(*fConfigFile, settings); err != nil {
		log.Fatal(errors.Wrap(err, "cannot parse config"))
	}

	cli := flags.Copy(f)
	if err := mergo.Map(settings, cli, mergo.WithOverwriteWithEmptyValue); err != nil {
		log.Fatal(errors.Wrap(err, "cannot process settings"))
	}

	if err := settings.Validate(); err != nil {
		log.Fatal(errors.Wrap(err, "settings validation error"))
	}

	loggerOut, logger := logger.Setup("azure-connector", &settings.LogSettings)
	defer loggerOut.Close()

	logger.Infof("Starting azure connector %s", version)
	flags.ConfigCheck(logger, *fConfigFile)

	telemetryHandlers := telemetryHandlers(settings)
	commandHandlers := commandHandlers(settings)

	if err := app.MainLoop(settings, logger, nil, telemetryHandlers, commandHandlers); err != nil {
		logger.Error("Init failure", err, nil)

		loggerOut.Close()

		os.Exit(1)
	}
}

func telemetryHandlers(settings *azurecfg.AzureSettings) []handlers.TelemetryHandler {
	passthroughHandler := passthrough.CreateDefaultTelemetryHandler()
	return []handlers.TelemetryHandler{passthroughHandler}
}

func commandHandlers(settings *azurecfg.AzureSettings) []handlers.CommandHandler {
	passthroughHandler := passthrough.CreateDefaultCommandHandler()
	return []handlers.CommandHandler{passthroughHandler}
}
