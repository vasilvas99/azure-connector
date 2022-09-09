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
	"io"
	"log"
	"strconv"
	"testing"

	"github.com/eclipse-kanto/azure-connector/config"
	handlers "github.com/eclipse-kanto/azure-connector/routing/message/handlers/common"
	"github.com/eclipse-kanto/suite-connector/logger"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPassthroughMessageHandler(t *testing.T) {
	settings := &config.AzureSettings{
		ConnectionString:       "HostName=dummy-hub.azure-devices.net;DeviceId=dummy-device;SharedAccessKey=dGVzdGF6dXJlc2hhcmVkYWNjZXNza2V5",
		AllowedLocalTopicsList: "localTopic1,localTopic2,localTopic3",
	}
	logger := logger.NewLogger(log.New(io.Discard, "", log.Ldate), logger.INFO)
	connSettings, err := config.CreateAzureConnectionSettings(settings, logger)
	require.NoError(t, err)
	messageHandler := &passthroughMessageHandler{}

	require.NoError(t, messageHandler.Init(settings, connSettings))
	assert.Equal(t, messageHandler.connSettings, connSettings)
	assert.Equal(t, messageHandler.settings, settings)
	for i, localTopic := range messageHandler.localTopics {
		assert.Equal(t, "localTopic"+strconv.Itoa(i+1), localTopic)
	}
	assert.Equal(t, passthroughHandlerName, messageHandler.Name())
	assert.Equal(t, []string{"localTopic1", "localTopic2", "localTopic3"}, messageHandler.Topics())
}

func TestPassthroughMessage(t *testing.T) {
	handler := createPassthroughMessageHandler(t)
	payload := "dummy_message"
	outgoingMessages, err := handler.HandleMessage(&message.Message{Payload: []byte(payload)})
	require.NoError(t, err)
	assert.NotNil(t, outgoingMessages)
	assert.Equal(t, 1, len(outgoingMessages))
	assert.Equal(t, payload, string(outgoingMessages[0].Payload))
}

func createPassthroughMessageHandler(t *testing.T) handlers.MessageHandler {
	settings := &config.AzureSettings{
		ConnectionString: "HostName=dummy-hub.azure-devices.net;DeviceId=dummy-device;SharedAccessKey=dGVzdGF6dXJlc2hhcmVkYWNjZXNza2V5",
	}
	logger := logger.NewLogger(log.New(io.Discard, "", log.Ldate), logger.INFO)
	connSettings, err := config.CreateAzureConnectionSettings(settings, logger)
	require.NoError(t, err)
	messageHandler := &passthroughMessageHandler{}
	require.NoError(t, messageHandler.Init(settings, connSettings))
	return messageHandler
}
