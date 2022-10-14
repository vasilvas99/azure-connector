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

package app

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/pkg/errors"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/eclipse-kanto/suite-connector/config"
	"github.com/eclipse-kanto/suite-connector/connector"
	"github.com/eclipse-kanto/suite-connector/logger"
	"github.com/eclipse-kanto/suite-connector/routing"

	azurecfg "github.com/eclipse-kanto/azure-connector/config"
	azurerouting "github.com/eclipse-kanto/azure-connector/routing"
	routingbus "github.com/eclipse-kanto/azure-connector/routing/bus"
	"github.com/eclipse-kanto/azure-connector/routing/message/handlers"
)

func startRouter(
	localClient *connector.MQTTConnection,
	settings *azurecfg.AzureSettings,
	connSettings *azurecfg.AzureConnectionSettings,
	statusPub message.Publisher,
	telemetryHandlers []handlers.TelemetryHandler,
	commandHandlers []handlers.CommandHandler,
	done chan bool,
	logger logger.Logger,
) (*message.Router, error) {
	cloudClient, err := config.CreateCloudConnection(&settings.LocalConnectionSettings, false, logger)
	if err != nil {
		return nil, errors.Wrap(err, "cannot create mosquitto connection")
	}

	azureClient, err := azurecfg.CreateAzureHubConnection(settings, connSettings, logger)
	if err != nil {
		routing.SendStatus(routing.StatusConnectionError, statusPub, logger)
		return nil, errors.Wrap(err, "cannot create Hub connection")
	}

	logger.Info("Starting messages router...", nil)
	router, err := message.NewRouter(message.RouterConfig{}, logger)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create router")
	}

	paramsPub := connector.NewPublisher(localClient, connector.QosAtMostOnce, logger, nil)
	paramsSub := connector.NewSubscriber(cloudClient, connector.QosAtMostOnce, false, logger, nil)

	gwParams := azurerouting.NewAzureGwParams(connSettings.DeviceID, settings.TenantID, connSettings.HubName)
	routing.ParamsBus(router, gwParams, paramsPub, paramsSub, logger)
	routing.SendGwParams(gwParams, false, paramsPub, logger)

	azurePub := connector.NewPublisher(azureClient, connector.QosAtLeastOnce, logger, nil)
	azureSub := connector.NewSubscriber(azureClient, connector.QosAtMostOnce, false, logger, nil)
	mosquittoSub := connector.NewSubscriber(cloudClient, connector.QosAtLeastOnce, false, router.Logger(), nil)

	routingbus.TelemetryBus(router, azurePub, mosquittoSub, &connSettings.RemoteConnectionInfo, telemetryHandlers)

	cloudPub := connector.NewPublisher(cloudClient, connector.QosAtLeastOnce, router.Logger(), nil)
	routingbus.CommandBus(router, cloudPub, azureSub, &connSettings.RemoteConnectionInfo, commandHandlers)

	go func() {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		go func() {
			defer func() {
				done <- true
			}()

			<-router.Running()

			statusHandler := &routing.ConnectionStatusHandler{
				Pub:    statusPub,
				Logger: logger,
			}
			cloudClient.AddConnectionListener(statusHandler)

			reconnectHandler := &routing.ReconnectHandler{
				Pub:    paramsPub,
				Params: gwParams,
				Logger: logger,
			}
			cloudClient.AddConnectionListener(reconnectHandler)

			connHandler := &routing.CloudConnectionHandler{
				CloudClient: cloudClient,
				Logger:      logger,
			}
			azureClient.AddConnectionListener(connHandler)

			errorsHandler := &routing.ErrorsHandler{
				StatusPub: statusPub,
				Logger:    logger,
			}
			azureClient.AddConnectionListener(errorsHandler)

			if err := config.HonoConnect(nil, statusPub, azureClient, logger); err != nil {
				router.Close()
				return
			}

			if connSettings.SharedAccessKey != nil {
				tokenRefreshPeriod := int64(connSettings.TokenValidity.Seconds() * azurecfg.SASTokenValidityFactor)
				go func() {
					for {
						select {
						case <-time.After(time.Duration(tokenRefreshPeriod) * time.Second):
							logger.Debug("SAS token validity period is about to expire.", nil)
							azureClient.Disconnect()
							routing.SendStatus(azurerouting.StatusConnectionTokenExpired, statusPub, logger)

							if err := config.HonoConnect(nil, statusPub, azureClient, logger); err != nil {
								router.Close()
								return
							}
						case <-ctx.Done():
							return
						}
					}
				}()
			}

			<-ctx.Done()

			azureClient.RemoveConnectionListener(errorsHandler)
			azureClient.RemoveConnectionListener(connHandler)
			cloudClient.RemoveConnectionListener(reconnectHandler)
			cloudClient.RemoveConnectionListener(statusHandler)

			defer routing.SendStatus(routing.StatusConnectionClosed, statusPub, logger)

			defer azureClient.Disconnect()

			cloudClient.Disconnect()
		}()

		if err := router.Run(context.Background()); err != nil {
			logger.Error("Failed to create cloud router", err, nil)
		}

		logger.Info("Messages router stopped", nil)
	}()

	return router, nil
}

// MainLoop is the main loop of the application
func MainLoop(settings *azurecfg.AzureSettings, log logger.Logger, idScopeProvider azurecfg.IDScopeProvider, telemetryHandlers []handlers.TelemetryHandler, commandHandlers []handlers.CommandHandler) error {
	localClient, err := config.CreateLocalConnection(&settings.LocalConnectionSettings, log)
	if err != nil {
		return errors.Wrap(err, "cannot create mosquitto connection")
	}
	if err := config.LocalConnect(context.Background(), localClient, log); err != nil {
		return errors.Wrap(err, "cannot connect to mosquitto")
	}
	defer localClient.Disconnect()

	statusPub := connector.NewPublisher(localClient, connector.QosAtLeastOnce, log, nil)
	defer statusPub.Close()

	connSettings, err := azurecfg.PrepareAzureConnectionSettings(settings, idScopeProvider, log)
	if err != nil {
		return errors.Wrap(err, "cannot create Azure IoT Hub device connection settings")
	}

	done := make(chan bool, 1)
	azureRouter, err := startRouter(localClient, settings, connSettings, statusPub, telemetryHandlers, commandHandlers, done, log)
	if err != nil {
		log.Error("Failed to create message bus", err, nil)
	}

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	<-sigs

	stopRouter(azureRouter, done)

	return nil
}

func stopRouter(router *message.Router, done <-chan bool) {
	if router != nil {
		router.Close()
		<-done
	}
}
