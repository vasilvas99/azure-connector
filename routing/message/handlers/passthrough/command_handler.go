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

package passthrough

import (
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/eclipse-kanto/azure-connector/config"
	"github.com/eclipse-kanto/suite-connector/connector"

	"github.com/eclipse-kanto/azure-connector/routing/message/handlers"
)

const commandHandlerName = "passthrough_command_handler"

// A simple command passthrough handler that forwards all cloud-to-device messages from Azure IoT Hub to local MQTT broker on a preconfigured topic.
type commandHandler struct {
	topics string
}

// CreateCommandHandler instantiates a new command handler that forward cloud-to-device messages to the local message broker using the given topic.
func CreateCommandHandler() handlers.MessageHandler {
	return &commandHandler{}
}

// Init gets the local topic to publish the messages.
func (h *commandHandler) Init(settings *config.AzureSettings, connSettings *config.AzureConnectionSettings) error {
	h.topics = settings.PassthroughCommandTopic
	return nil
}

// HandleMessage creates a new message with the same payload as the incoming message and sets the configured local topic to publish it.
func (h *commandHandler) HandleMessage(msg *message.Message) ([]*message.Message, error) {
	msgID := watermill.NewUUID()
	outgoingMessage := message.NewMessage(msgID, msg.Payload)
	outgoingMessage.SetContext(connector.SetTopicToCtx(outgoingMessage.Context(), h.topics))
	return []*message.Message{outgoingMessage}, nil
}

// Name returns the message handler name.
func (h *commandHandler) Name() string {
	return commandHandlerName
}

// Topics returns the configurable topics, actually never used for the command handlers.
func (h *commandHandler) Topics() []string {
	return []string{h.topics}
}
