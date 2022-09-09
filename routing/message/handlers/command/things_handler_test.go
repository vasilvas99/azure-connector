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

//go:build things
// +build things

package command

import (
	"encoding/json"
	"io"
	"log"
	"testing"

	"github.com/eclipse-kanto/azure-connector/config"
	handlers "github.com/eclipse-kanto/azure-connector/routing/message/handlers/common"
	"github.com/eclipse-kanto/azure-connector/util"
	"github.com/eclipse-kanto/suite-connector/connector"
	"github.com/eclipse-kanto/suite-connector/logger"
	"github.com/eclipse/ditto-clients-golang/protocol"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestThingsMessageHandler(t *testing.T) {
	settings := &config.AzureSettings{
		ConnectionString:             "HostName=dummy-hub.azure-devices.net;DeviceId=dummy-device;SharedAccessKey=dGVzdGF6dXJlc2hhcmVkYWNjZXNza2V5",
		MessageMapperConfig:          "../internal/testdata/handlers-mapper-config.json",
		AllowedCloudMessageTypesList: "testVal,testCommand",
	}
	logger := logger.NewLogger(log.New(io.Discard, "", log.Ldate), logger.INFO)
	connSettings, err := config.CreateAzureConnectionSettings(settings, logger)
	require.NoError(t, err)
	messageHandler := &commandThingsMessageHandler{}

	messageHandler.Init(settings, connSettings)
	assert.Equal(t, commandThingsHandlerName, messageHandler.Name())
	supportedCommandNames := []string{"container.manifest", "simple.message", "unsupported.field", "message.without.thing", "message.only.path"}
	assert.Equal(t, len(supportedCommandNames), len(messageHandler.Topics()))
	for _, commandName := range messageHandler.Topics() {
		assert.Equal(t, true, util.ContainsString(supportedCommandNames, commandName))
	}
}

func TestHandleC2DMessageCorrectly(t *testing.T) {
	handler := createCommandThingsHandler(t)
	var testData = []struct {
		jsonPayload   string
		msgTopic      string
		topic         string
		path          string
		correlationId string
	}{
		{
			`{
				"appId": "app1",
				"cmdName": "container.manifest",
				"cId": "C2D-msg-correlation-id",
				"eVer": "1.0.0",
				"pVer": "1.0.0",
				"p": "Cq4BCglzb21lLW5hbWUSBTAuMC4xGmAKEW15LWNvbnRhaW5lci1uYW1lEiFkb2NrZXIuaW8vbGlicmFyeS9pbmZsdXhkYjpsYXRlc3QaHQoLdGVzdC1kb21haW4SBlZBUjE9MhIGVkFSPXR6KgkKB1JVTk5JTkcaOAoQc2Vjb25kLWNvbnRhaW5lchIkZG9ja2VyLmlvL2xpYnJhcnkvaGVsbG8td29ybGQ6bGF0ZXN0"
			}`,
			"command//azure.edge:dummy-hub:dummy-device:edge:containers/req/C2D-msg-correlation-id/apply",
			"azure.edge/dummy-hub:dummy-device:edge:containers/things/live/messages/apply",
			"/features/ContainerOrchestrator/inbox/messages/apply",
			"C2D-msg-correlation-id",
		},
		{
			`{
				"appId": "app1",
				"cmdName": "container.manifest",
				"cId": "C2D-msg-correlation-id2",
				"eVer": "1.0.0",
				"pVer": "1.0.0",
				"p": "Cq4BCglzb21lLW5hbWUSBTAuMC4xGmAKEW15LWNvbnRhaW5lci1uYW1lEiFkb2NrZXIuaW8vbGlicmFyeS9pbmZsdXhkYjpsYXRlc3QaHQoLdGVzdC1kb21haW4SBlZBUjE9MhIGVkFSPXR6KgkKB1JVTk5JTkcaOAoQc2Vjb25kLWNvbnRhaW5lchIkZG9ja2VyLmlvL2xpYnJhcnkvaGVsbG8td29ybGQ6bGF0ZXN0"
			}`,
			"command//azure.edge:dummy-hub:dummy-device:edge:containers/req/C2D-msg-correlation-id2/apply",
			"azure.edge/dummy-hub:dummy-device:edge:containers/things/live/messages/apply",
			"/features/ContainerOrchestrator/inbox/messages/apply",
			"C2D-msg-correlation-id2",
		},
	}
	for _, testValues := range testData {
		t.Run(testValues.msgTopic, func(t *testing.T) {
			azureMessages, err := handler.HandleMessage(createWatermillMessageForC2D([]byte(testValues.jsonPayload)))
			require.NoError(t, err)

			azureMsg := azureMessages[0]
			azureMsgTopic, _ := connector.TopicFromCtx(azureMsg.Context())

			c2dMessage := &protocol.Envelope{}
			err = json.Unmarshal(azureMsg.Payload, c2dMessage)
			require.NoError(t, err)

			// TODO: needs change for ditto envelope API
			// c2dMessagePayload := make(map[string]interface{})
			// json.Unmarshal(c2dMessage.Value, &c2dMessagePayload)
			// assert.True(t, len(c2dMessagePayload) > 0)

			assert.Equal(t, testValues.msgTopic, azureMsgTopic)
			assert.Equal(t, testValues.topic, c2dMessage.Topic.String())
			assert.Equal(t, testValues.path, c2dMessage.Path)
			assert.Equal(t, "application/json", c2dMessage.Headers.ContentType())
			assert.Equal(t, testValues.correlationId, c2dMessage.Headers.CorrelationID())
		})
	}
}

func TestValidC2DMessageWithTypeSimpleMessage(t *testing.T) {
	handler := createCommandThingsHandler(t)
	var testSimpleData = []struct {
		jsonPayload    string
		msgTopic       string
		topic          string
		path           string
		correlationId  string
		payloadText    string
		payloadVersion string
	}{
		{
			`{
				"appId": "app1",
				"cmdName": "simple.message",
				"cId": "C2D-msg-correlation-id",
				"eVer": "1.0.0",
				"pVer": "1.0.0",
				"p": "ChRzb21lLW1lc3NhZ2UtaWQtMTIzNBIRc2ltcGxlIHRleHQgYWRkZWQaBTEuMC44"
			}`,
			"command//azure.edge:dummy-hub:dummy-device:edge:containers/req/C2D-msg-correlation-id/send",
			"azure.edge/dummy-hub:dummy-device:edge:containers/things/live/messages/send",
			"/features/ContainerOrchestrator/inbox/messages/send",
			"C2D-msg-correlation-id",
			"simple text added",
			"1.0.0",
		},
		{
			`{
				"appId": "app1",
				"cmdName": "simple.message",
				"eVer": "1.0.0",
				"pVer": "1.0.0",
				"p": "ChRzb21lLW1lc3NhZ2UtaWQtMTIzNBIRc2ltcGxlIHRleHQgYWRkZWQaBTEuMC44"
			}`,
			"command//azure.edge:dummy-hub:dummy-device:edge:containers/req//send",
			"azure.edge/dummy-hub:dummy-device:edge:containers/things/live/messages/send",
			"/features/ContainerOrchestrator/inbox/messages/send",
			"",
			"simple text added",
			"1.0.0",
		},
		{
			`{
				"appId": "app1",
				"cmdName": "message.without.thing",
				"cId": "C2D-msg-correlation-id",
				"eVer": "1.0.0",
				"pVer": "1.0.0",
				"p": "ChRzb21lLW1lc3NhZ2UtaWQtMTIzNBIRc2ltcGxlIHRleHQgYWRkZWQaBTEuMC44"
			}`,
			"command///req/C2D-msg-correlation-id/noThing",
			"azure.edge/dummy-hub:dummy-device/things/live/messages/noThing",
			"/features/ContainerOrchestrator/inbox/messages/noThing",
			"C2D-msg-correlation-id",
			"simple text added",
			"1.0.0",
		},
	}

	for _, testValues := range testSimpleData {
		t.Run(testValues.msgTopic, func(t *testing.T) {
			azureMessages, err := handler.HandleMessage(createWatermillMessageForC2D([]byte(testValues.jsonPayload)))
			require.NoError(t, err)

			azureMsg := azureMessages[0]
			azureMsgTopic, _ := connector.TopicFromCtx(azureMsg.Context())

			c2dMessage := &protocol.Envelope{}
			err = json.Unmarshal(azureMsg.Payload, c2dMessage)
			require.NoError(t, err)

			// TODO: needs change for ditto envelope API
			// c2dMessagePayload := make(map[string]interface{})
			// json.Unmarshal(c2dMessage.Value, &c2dMessagePayload)
			// assert.True(t, len(c2dMessagePayload) > 0)

			assert.Equal(t, testValues.msgTopic, azureMsgTopic)
			assert.Equal(t, testValues.topic, c2dMessage.Topic.String())
			assert.Equal(t, testValues.path, c2dMessage.Path)
			assert.Equal(t, "application/json", c2dMessage.Headers.ContentType())
			assert.Equal(t, testValues.correlationId, c2dMessage.Headers.CorrelationID())
		})
	}
}

func TestValidC2DMessageWithTypeUnsupportedProtobufFieldInJSON(t *testing.T) {
	handler := createCommandThingsHandler(t)
	jsonPayload := `{
		"appId": "app1",
		"cmdName": "unsupported.field",
		"cId": "C2D-msg-correlation-id",
		"eVer": "1.0.0",
		"pVer": "1.0.0",
		"p": "ChRzb21lLW1lc3NhZ2UtaWQtMTIzNBIRc2ltcGxlIHRleHQgYWRkZWQaBTEuMC44"
	}`
	_, err := handler.HandleMessage(createWatermillMessageForC2D([]byte(jsonPayload)))
	require.Error(t, err)
}

func TestUnsupportedC2DVSSMessageType(t *testing.T) {
	handler := createCommandThingsHandler(t)
	jsonPayload := `{
		"appId": "app1",
		"cmdName": "subscribeCommand",
		"cId": "C2D-msg-correlation-id",
		"eVer": "1.0.0",
		"pVer": "1.0.0",
		"p": {}
	}`
	_, err := handler.HandleMessage(createWatermillMessageForC2D([]byte(jsonPayload)))
	require.Error(t, err)
}

func TestUnsupportedC2DMessageType(t *testing.T) {
	handler := createCommandThingsHandler(t)
	jsonPayload := `{
		"appId": "app1",
		"cmdName": "container.non-existing",
		"cId": "C2D-msg-correlation-id",
		"eVer": "1.0.0",
		"pVer": "1.0.0",
		"p": {}
	}`
	_, err := handler.HandleMessage(createWatermillMessageForC2D([]byte(jsonPayload)))
	require.Error(t, err)
}

func TestInvalidPayloadC2DMessageType(t *testing.T) {
	handler := createCommandThingsHandler(t)
	jsonPayload := `{
		"appId": "app1",
		"cmdName": "container.manifest",
		"cId": "C2D-msg-correlation-id",
		"eVer": "1.0.0",
		"pVer": "1.0.0",
		"p": "Cq4BCglzb21lLW5hbWUSBTAuMdGVzdC1kb21haW4SBlZBUjE9MhIGVkFSPXR6KgkKB1JVTk5JTkcaOAoQc2Vjb25kLWNvbnRhaW5lchIkZG9ja2VyLmlvL2xpYnJhcnkvaGVsbG8td29ybGQ6bGF0ZXN0"
	}`
	_, err := handler.HandleMessage(createWatermillMessageForC2D([]byte(jsonPayload)))
	require.Error(t, err)
}

func TestNoPayloadC2DMessageType(t *testing.T) {
	handler := createCommandThingsHandler(t)
	jsonPayload := `{
		"appId": "app1",
		"cmdName": "container.manifest",
		"cId": "C2D-msg-correlation-id",
		"eVer": "1.0.0"
	}`
	messages, err := handler.HandleMessage(createWatermillMessageForC2D([]byte(jsonPayload)))
	require.NoError(t, err)
	dittoMessage := &protocol.Envelope{}
	err = json.Unmarshal(messages[0].Payload, dittoMessage)
	require.NoError(t, err)
	// TODO: needs change for ditto envelope API
	// dittoPayload := make(map[string]interface{})
	// json.Unmarshal(dittoMessage.Value, &dittoPayload)
	// assert.Equal(t, "{}", string(dittoMessage.Value))
}

func TestInvalidJSONC2DMessageType(t *testing.T) {
	handler := createCommandThingsHandler(t)
	jsonPayload := `{
		"appId": "app1",
		"cmdName": "container.manifest",
		"cId": "C2D-msg-correlation-id",
		"eVer": "1.0.0",
		"pVer": "1.0.0",
		"p": {}`
	_, err := handler.HandleMessage(createWatermillMessageForC2D([]byte(jsonPayload)))
	require.Error(t, err)
}

func createCommandThingsHandler(t *testing.T) handlers.MessageHandler {
	settings := &config.AzureSettings{
		ConnectionString:             "HostName=dummy-hub.azure-devices.net;DeviceId=dummy-device;SharedAccessKey=dGVzdGF6dXJlc2hhcmVkYWNjZXNza2V5",
		MessageMapperConfig:          "../internal/testdata/handlers-mapper-config.json",
		AllowedCloudMessageTypesList: "testVal,testCommand",
	}
	logger := logger.NewLogger(log.New(io.Discard, "", log.Ldate), logger.INFO)
	connSettings, err := config.CreateAzureConnectionSettings(settings, logger)
	require.NoError(t, err)
	messageHandler := &commandThingsMessageHandler{}
	messageHandler.Init(settings, connSettings)
	return messageHandler
}
