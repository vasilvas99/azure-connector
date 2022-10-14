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
	"github.com/eclipse-kanto/azure-connector/routing"
	"github.com/eclipse-kanto/suite-connector/connector"

	"github.com/eclipse-kanto/azure-connector/routing/message/handlers"
)

const (
	telemetryHandlerName = "passthrough_telemetry_handler"
)

// telemetryHandler forwards incoming device messages published on certain local topics to the Azure Iot Hub.
type telemetryHandler struct {
	deviceID string
	topics   string
}

// CreateTelemetryHandler instantiates a new passthrough telemetry handler that forward messages received from local message broker on the given topics as device-to-cloud messages to Azure IoT Hub.
func CreateTelemetryHandler(topics string) handlers.TelemetryHandler {
	return &telemetryHandler{
		topics: topics,
	}
}

// Init gets the device ID that is needed for the message forwarding towards Azure IoT Hub.
func (h *telemetryHandler) Init(connInfo *config.RemoteConnectionInfo) error {
	h.deviceID = connInfo.DeviceID
	return nil
}

// HandleMessage creates a new message with the same payload as the incoming message and sets the correct topic so that the message can be forwarded to Azure Iot Hub
func (h *telemetryHandler) HandleMessage(msg *message.Message) ([]*message.Message, error) {
	msgID := watermill.NewUUID()
	outgoingMessage := message.NewMessage(msgID, msg.Payload)
	outgoingTopic := routing.CreateTelemetryTopic(h.deviceID, msgID)
	outgoingMessage.SetContext(connector.SetTopicToCtx(outgoingMessage.Context(), outgoingTopic))
	return []*message.Message{outgoingMessage}, nil
}

// Name returns the message handler name.
func (h *telemetryHandler) Name() string {
	return telemetryHandlerName
}

// Topics returns the configurable list of topics that are used for subscription on the local message broker.
func (h *telemetryHandler) Topics() string {
	return h.topics
}
