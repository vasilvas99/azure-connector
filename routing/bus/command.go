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
	"fmt"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/eclipse-kanto/suite-connector/connector"

	"github.com/eclipse-kanto/azure-connector/config"
	"github.com/eclipse-kanto/azure-connector/routing"
	"github.com/eclipse-kanto/azure-connector/routing/message/handlers"
)

const (
	commandHandlerName = "command_handler"
)

type commandBusHandler struct {
	logger          watermill.LoggerAdapter
	commandHandlers []handlers.MessageHandler
}

// CommandBus creates the cloud message bus for processing the C2D messages from the Azure IoT Hub device.
func CommandBus(router *message.Router,
	mosquittoPub message.Publisher,
	azureSub message.Subscriber,
	settings *config.AzureSettings,
	connSettings *config.AzureConnectionSettings,
	commandHandlers []handlers.MessageHandler,
) {
	//Azure IoT Hub -> Message bus -> Mosquitto Broker -> Gateway
	initCommandHandlers := []handlers.MessageHandler{}
	commandBusHandler := &commandBusHandler{
		logger: router.Logger(),
	}
	for _, commandHandler := range commandHandlers {
		if err := commandHandler.Init(settings, connSettings); err != nil {
			logFields := watermill.LogFields{"handler_name": commandHandler.Name()}
			router.Logger().Error("skipping command handler that cannot be initialized", err, logFields)
			continue
		}
		initCommandHandlers = append(initCommandHandlers, commandHandler)
	}
	commandBusHandler.commandHandlers = initCommandHandlers
	router.AddHandler(commandHandlerName,
		routing.CreateRemoteCloudTopic(connSettings.DeviceID),
		azureSub,
		connector.TopicEmpty,
		mosquittoPub,
		commandBusHandler.HandleMessage,
	)
}

func (h *commandBusHandler) HandleMessage(msg *message.Message) ([]*message.Message, error) {
	for _, commandHandler := range h.commandHandlers {
		msg, err := commandHandler.HandleMessage(msg)
		if err == nil {
			return msg, nil
		}
		logFields := watermill.LogFields{"handler_name": commandHandler.Name()}
		h.logger.Error("error handling command message", err, logFields)
	}
	return nil, fmt.Errorf("cannot handle command message '%v'", string(msg.Payload))

}
