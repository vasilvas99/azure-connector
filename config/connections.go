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
	"fmt"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/pkg/errors"

	"github.com/ThreeDotsLabs/watermill"

	"github.com/eclipse-kanto/azure-connector/util"
	"github.com/eclipse-kanto/suite-connector/config"
	"github.com/eclipse-kanto/suite-connector/connector"
)

//CreateAzureHubConnection creates the MQTT connection to the remote Azure Iot Hub MQTT broker.
func CreateAzureHubConnection(
	settings *AzureSettings, connSettings *AzureConnectionSettings, logger watermill.LoggerAdapter,
) (*connector.MQTTConnection, error) {
	configuration, connErr := createMQTTConfiguration(settings, connSettings, logger)
	if connErr != nil {
		return nil, errors.Wrap(connErr, "cannot establish MQTT connection to Azure IoT Hub")
	}
	provider := func() (string, string) {
		var pass string
		username := connSettings.HostName + "/" + connSettings.DeviceID + "/api-version=2020-09-30"
		if connSettings.SharedAccessKey != nil {
			sasToken := GenerateSASToken(connSettings)
			logger.Debug(
				fmt.Sprintf("Generated SAS token with validity %s. Expires after %s.",
					connSettings.TokenValidity,
					sasToken.Se,
				),
				nil)
			pass = sasTokenToString(sasToken)
		}
		return username, pass
	}
	return connector.NewMQTTConnectionCredentialsProvider(configuration, connSettings.DeviceID, logger, provider)
}

func createMQTTConfiguration(
	settings *AzureSettings, connSettings *AzureConnectionSettings, logger watermill.LoggerAdapter,
) (*connector.Configuration, error) {
	brokerURL := url.URL{
		Scheme: "tls",
		Host:   fmt.Sprintf("%s:%d", connSettings.HostName, 8883),
	}

	configuration, err := connector.NewMQTTClientConfig(brokerURL.String())
	if err != nil {
		return nil, errors.Wrap(err, "cannot create MQTT client configuration")
	}

	if connSettings.SharedAccessKey == nil &&
		!util.DeviceCertificatesArePresent(connSettings.DeviceCert, connSettings.DeviceKey) {
		return nil, errors.New("missing the PEM encoded certificate file and private key file for device authentication")
	}

	tlsConfig, _, err := config.NewHubTLSConfig(&settings.TLSSettings, logger)
	if err != nil {
		return nil, errors.Wrap(err, "cannot create TLS configuration")
	}

	configuration.TLSConfig = tlsConfig
	configuration.ConnectRetryInterval = 0

	configuration.BackoffMultiplier = 2
	configuration.MinReconnectInterval = time.Minute
	configuration.MaxReconnectInterval = time.Minute * 4

	if min, err := strconv.ParseInt(os.Getenv("HUB_CONNECT_INIT"), 0, 64); err == nil {
		configuration.MinReconnectInterval = time.Duration(min) * time.Second
	}

	if max, err := strconv.ParseInt(os.Getenv("HUB_CONNECT_MAX"), 0, 64); err == nil {
		configuration.MaxReconnectInterval = time.Duration(max) * time.Second
	}

	if mul, err := strconv.ParseFloat(os.Getenv("HUB_CONNECT_MUL"), 32); err == nil {
		configuration.BackoffMultiplier = mul
	}
	return configuration, nil
}
