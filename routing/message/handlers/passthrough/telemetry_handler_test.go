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
	"testing"

	"github.com/eclipse-kanto/azure-connector/config"
	"github.com/eclipse-kanto/azure-connector/routing/message/handlers"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateTelemetryHandler(t *testing.T) {
	topic := "localTopic1,localTopic2,localTopic3"
	deviceID := "dummy-device"
	messageHandler := createTelemetryHandler(t, topic, deviceID)
	assert.Equal(t, telemetryHandlerName, messageHandler.Name())
	assert.Equal(t, []string{"localTopic1", "localTopic2", "localTopic3"}, messageHandler.Topics())
}

func TestHandleTelemetryMessage(t *testing.T) {
	topic := "test-telemetry"
	deviceID := "dummy-device"
	handler := createTelemetryHandler(t, topic, deviceID)
	payload := "dummy_message"
	outgoingMessages, err := handler.HandleMessage(&message.Message{Payload: []byte(payload)})
	require.NoError(t, err)
	assert.NotNil(t, outgoingMessages)
	assert.Equal(t, 1, len(outgoingMessages))
	assert.Equal(t, payload, string(outgoingMessages[0].Payload))
}

func createTelemetryHandler(t *testing.T, topic string, deviceID string) handlers.MessageHandler {
	settings := &config.AzureSettings{
		PassthroughTelemetryTopics: topic,
	}
	connSettings := &config.AzureConnectionSettings{
		DeviceID: deviceID,
	}
	messageHandler := CreateTelemetryHandler()
	require.NoError(t, messageHandler.Init(settings, connSettings))
	return messageHandler
}
