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
	"strings"
	"testing"

	"github.com/ThreeDotsLabs/watermill/message"

	"github.com/eclipse-kanto/azure-connector/config"

	"github.com/eclipse-kanto/suite-connector/connector"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateDefaultTelemetryHandler(t *testing.T) {
	messageHandler := CreateDefaultTelemetryHandler()
	assert.Equal(t, telemetryHandlerName, messageHandler.Name())
	assert.Equal(t, topicsEvent, messageHandler.Topics())
}

func TestCreateTelemetryHandler(t *testing.T) {
	topic := "localTopic1,localTopic2,localTopic3"

	messageHandler := CreateTelemetryHandler(topic)
	assert.Equal(t, telemetryHandlerName, messageHandler.Name())
	assert.Equal(t, topic, messageHandler.Topics())
}

func TestHandleTelemetryMessage(t *testing.T) {
	handler := CreateTelemetryHandler("telemetry_topic")
	require.NoError(t, handler.Init(&config.RemoteConnectionInfo{DeviceID: "dummy_device"}))

	payload := "dummy_message"
	outgoingMessages, err := handler.HandleMessage(&message.Message{Payload: []byte(payload)})
	require.NoError(t, err)
	assert.NotNil(t, outgoingMessages)
	assert.Equal(t, 1, len(outgoingMessages))

	message := outgoingMessages[0]
	messageTopic, _ := connector.TopicFromCtx(message.Context())
	assert.True(t, strings.HasPrefix(messageTopic, "devices/dummy_device/messages/events/"))
	assert.Equal(t, payload, string(message.Payload))
}
