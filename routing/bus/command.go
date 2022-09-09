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
	"encoding/json"
	"fmt"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/eclipse-kanto/suite-connector/connector"
	"github.com/pkg/errors"

	"github.com/eclipse-kanto/azure-connector/config"
	"github.com/eclipse-kanto/azure-connector/routing"
	routingmessage "github.com/eclipse-kanto/azure-connector/routing/message"
	"github.com/eclipse-kanto/azure-connector/routing/message/handlers/command"
	handlers "github.com/eclipse-kanto/azure-connector/routing/message/handlers/common"
	"github.com/eclipse-kanto/azure-connector/util"
)

const (
	commandHandlerName = "command_handler"
)

type commandBusHandler struct {
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
	commandBusHandler := &commandBusHandler{}
	for _, commandHandler := range commandHandlers {
		if err := commandHandler.Init(settings, connSettings); err != nil {
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
	cloudMessage := &routingmessage.CloudMessage{}
	if err := json.Unmarshal(msg.Payload, cloudMessage); err != nil {
		return nil, errors.Wrap(err, "cannot deserialize cloud message")
	}
	msg.SetContext(command.SetMessageToContext(msg, cloudMessage))
	for _, commandHandler := range h.commandHandlers {
		if util.ContainsString(commandHandler.Topics(), cloudMessage.CommandName) {
			return commandHandler.HandleMessage(msg)
		}
	}
	return nil, fmt.Errorf("no message handler for command name '%s'", cloudMessage.CommandName)
}
