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

package bus

import (
	"strings"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/eclipse-kanto/suite-connector/connector"

	"github.com/eclipse-kanto/azure-connector/config"
	"github.com/eclipse-kanto/azure-connector/routing/message/handlers"
)

// TelemetryBus creates the telemetry message bus for processing & forwarding the telemetry messages from the local MQTT broker to the Azure IoT Hub.
func TelemetryBus(
	router *message.Router,
	azurePub message.Publisher,
	mosquittoSub message.Subscriber,
	settings *config.AzureSettings,
	connSettings *config.AzureConnectionSettings,
	telemetryHandlers []handlers.MessageHandler,
) {
	//Gateway -> Mosquitto Broker -> Message bus -> Azure IoT Hub
	for _, telemetryHandler := range telemetryHandlers {
		if err := telemetryHandler.Init(settings, connSettings); err != nil {
			logFields := watermill.LogFields{"handler_name": telemetryHandler.Name()}
			router.Logger().Error("skipping telemetry handler that cannot be initialized", err, logFields)
			continue
		}
		handlerName := telemetryHandler.Name()
		handlerTopics := telemetryHandler.Topics()
		if len(handlerTopics) == 0 {
			logFields := watermill.LogFields{"handler_name": telemetryHandler.Name()}
			router.Logger().Error("skipping telemetry handler without any topics", nil, logFields)
			continue
		}
		router.AddHandler(handlerName,
			strings.Join(handlerTopics, ","),
			mosquittoSub,
			connector.TopicEmpty,
			azurePub,
			telemetryHandler.HandleMessage,
		)
	}
}
