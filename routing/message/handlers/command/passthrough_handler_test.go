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
	"encoding/json"
	"io"
	"log"
	"strings"
	"testing"

	"github.com/eclipse-kanto/azure-connector/config"
	connectormessage "github.com/eclipse-kanto/azure-connector/routing/message"
	routingmessage "github.com/eclipse-kanto/azure-connector/routing/message"
	handlers "github.com/eclipse-kanto/azure-connector/routing/message/handlers/common"
	"github.com/eclipse-kanto/suite-connector/connector"
	"github.com/eclipse-kanto/suite-connector/logger"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPassthroughMessageHandler(t *testing.T) {
	settings := &config.AzureSettings{
		ConnectionString:             "HostName=dummy-hub.azure-devices.net;DeviceId=dummy-device;SharedAccessKey=dGVzdGF6dXJlc2hhcmVkYWNjZXNza2V5",
		AllowedCloudMessageTypesList: "testVal,testCommand",
	}
	logger := logger.NewLogger(log.New(io.Discard, "", log.Ldate), logger.INFO)
	connSettings, err := config.CreateAzureConnectionSettings(settings, logger)
	require.NoError(t, err)
	messageHandler := &commandPassthroughMessageHandler{}

	require.NoError(t, messageHandler.Init(settings, connSettings))
	supportedCommandNames := strings.Split(settings.AllowedCloudMessageTypesList, ",")
	assert.Equal(t, messageHandler.settings, settings)
	assert.Equal(t, messageHandler.topics, supportedCommandNames)
	assert.Equal(t, commandPassthroughHandlerName, messageHandler.Name())
	assert.Equal(t, supportedCommandNames, messageHandler.Topics())
}

func TestValidC2DMessageWithSupportedCommandName(t *testing.T) {
	handler := createCommandCoreSubHandler(t)
	jsonPayload := `{
		"appId": "datapoints",
		"cId": "C2D-msg-correlation-id",
		"cmdName": "testCommand",
		"eVer": "1.0.0",
		"pVer": "1.0.0",
		"p": "CoEBCglzb21lLW5hbWUSBTAuMC4xGjYKEW15LWNvbnRhaW5lci1uYW1lEiFkb2NrZXIuaW8vbGlicmFyeS9pbmZsdXhkYjpsYXRlc3QaNQoQc2Vjb25kLWNvbnRhaW5lchIhZG9ja2VyLmlvL2xpYnJhcnkvaW5mbHV4ZGI6bGF0ZXN0"
	}`

	azureMessages, err := handler.HandleMessage(createWatermillMessageForC2D([]byte(jsonPayload)))
	require.NoError(t, err)

	azureMsg := azureMessages[0]
	azureMsgTopic, _ := connector.TopicFromCtx(azureMsg.Context())

	c2dMessage := &connectormessage.CloudMessage{}
	err = json.Unmarshal(azureMsg.Payload, c2dMessage)
	require.NoError(t, err)

	assert.Equal(t, "datapoints/testCommand", azureMsgTopic)
	assert.Equal(t, "testCommand", c2dMessage.CommandName)
	assert.Equal(t, "datapoints", c2dMessage.ApplicationID)
	assert.Equal(t, "C2D-msg-correlation-id", c2dMessage.CorrelationID)
	assert.Equal(t, "1.0.0", c2dMessage.EnvelopeVersion)
	assert.True(t, len(c2dMessage.Payload.(string)) > 0)
}

func TestValidC2DMessageWithNotSupportedCommandName(t *testing.T) {
	handler := createCommandCoreSubHandler(t)
	jsonPayload := `{
		"appId": "datapoints",
		"cId": "C2D-msg-correlation-id",
		"cmdName": "not-supported",
		"eVer": "1.0.0",
		"pVer": "1.0.0",
		"p": "CoEBCglzb21lLW5hbWUSBTAuMC4xGjYKEW15LWNvbnRhaW5lci1uYW1lEiFkb2NrZXIuaW8vbGlicmFyeS9pbmZsdXhkYjpsYXRlc3QaNQoQc2Vjb25kLWNvbnRhaW5lchIhZG9ja2VyLmlvL2xpYnJhcnkvaW5mbHV4ZGI6bGF0ZXN0"
	}`

	azureMessages, err := handler.HandleMessage(createWatermillMessageForC2D([]byte(jsonPayload)))
	require.Error(t, err)
	assert.Nil(t, azureMessages)
}

func TestInvalidC2DMessagePayload(t *testing.T) {
	handler := createCommandCoreSubHandler(t)
	jsonPayload := "invalid-payload"
	_, err := handler.HandleMessage(createWatermillMessageForC2D([]byte(jsonPayload)))
	require.Error(t, err)
}

func createCommandCoreSubHandler(t *testing.T) handlers.MessageHandler {
	settings := &config.AzureSettings{
		ConnectionString:             "HostName=dummy-hub.azure-devices.net;DeviceId=dummy-device;SharedAccessKey=dGVzdGF6dXJlc2hhcmVkYWNjZXNza2V5",
		AllowedCloudMessageTypesList: "testVal,testCommand",
	}
	messageHandler := &commandPassthroughMessageHandler{
		settings: settings,
	}
	require.NoError(t, messageHandler.Init(settings, &config.AzureConnectionSettings{}))
	return messageHandler
}

func createWatermillMessageForC2D(payload []byte) *message.Message {
	message := message.NewMessage(watermill.NewUUID(), payload)
	cloudMessage := &routingmessage.CloudMessage{}
	if err := json.Unmarshal(payload, cloudMessage); err == nil {
		message.SetContext(SetMessageToContext(message, cloudMessage))
	}
	return message
}
