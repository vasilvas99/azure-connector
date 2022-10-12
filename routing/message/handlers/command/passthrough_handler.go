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

package command

import (
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/eclipse-kanto/azure-connector/config"
	"github.com/eclipse-kanto/suite-connector/connector"
)

const commandPassthroughHandlerName = "command_passthrough_handler"

// A simple command passthrough handler that forwards all cloud-to-device messages from Azure IoT Hub to local MQTT broker on a preconfigured topic.
type commandPassthroughMessageHandler struct {
	passthroughCommandTopic string
}

func (h *commandPassthroughMessageHandler) Init(settings *config.AzureSettings, connSettings *config.AzureConnectionSettings) error {
	h.passthroughCommandTopic = settings.PassthroughCommandTopic
	return nil
}

func (h *commandPassthroughMessageHandler) HandleMessage(msg *message.Message) ([]*message.Message, error) {
	msgID := watermill.NewUUID()
	outgoingMessage := message.NewMessage(msgID, msg.Payload)
	outgoingMessage.SetContext(connector.SetTopicToCtx(outgoingMessage.Context(), h.passthroughCommandTopic))
	return []*message.Message{outgoingMessage}, nil
}

func (h *commandPassthroughMessageHandler) Name() string {
	return commandPassthroughHandlerName
}

func (h *commandPassthroughMessageHandler) Topics() []string {
	return []string{h.passthroughCommandTopic}
}

func init() {
	registerMessageHandler(&commandPassthroughMessageHandler{})
}
