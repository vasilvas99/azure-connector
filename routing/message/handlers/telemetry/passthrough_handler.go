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

package telemetry

import (
	"strings"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/eclipse-kanto/azure-connector/config"
	"github.com/eclipse-kanto/azure-connector/routing"
	"github.com/eclipse-kanto/suite-connector/connector"
)

const (
	passthroughHandlerName = "passthrough_telemetry_handler"
)

type passthroughMessageHandler struct {
	connSettings               *config.AzureConnectionSettings
	passthroughTelemetryTopics []string
}

func (h *passthroughMessageHandler) Init(settings *config.AzureSettings, connSettings *config.AzureConnectionSettings) error {
	h.connSettings = connSettings
	h.passthroughTelemetryTopics = strings.Split(settings.PassthroughTelemetryTopics, ",")
	return nil
}

func (h *passthroughMessageHandler) HandleMessage(msg *message.Message) ([]*message.Message, error) {
	msgID := watermill.NewUUID()
	outgoingMessage := message.NewMessage(msgID, msg.Payload)
	outgoingTopic := routing.CreateTelemetryTopic(h.connSettings.DeviceID, msgID)
	outgoingMessage.SetContext(connector.SetTopicToCtx(outgoingMessage.Context(), outgoingTopic))
	return []*message.Message{outgoingMessage}, nil
}

func (h *passthroughMessageHandler) Name() string {
	return passthroughHandlerName
}

func (h *passthroughMessageHandler) Topics() []string {
	return h.passthroughTelemetryTopics
}

func init() {
	registerMessageHandler(&passthroughMessageHandler{})
}
