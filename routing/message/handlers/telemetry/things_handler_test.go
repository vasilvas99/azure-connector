// Copyright (c) 2022 Contributors to the Eclipse Foundation
//
// See the NOTICE file(s) distributed with this work for additional
// information regarding copyright ownership.
//
// This program and the accompanying materials are made available under the
// terms of the Eclipse Public License 2.0 which is available at
// http://www.eclipse.org/legal/epl-2.0
//
// SPDX-License-Identifier: EPL-2.0

//go:build things
// +build things

package telemetry

import (
	"encoding/json"
	"io"
	"log"
	"testing"

	"github.com/eclipse-kanto/azure-connector/config"
	connectormessage "github.com/eclipse-kanto/azure-connector/routing/message"
	handlers "github.com/eclipse-kanto/azure-connector/routing/message/handlers/common"
	"github.com/eclipse-kanto/suite-connector/logger"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestThingsMessageHandler(t *testing.T) {
	settings := &config.AzureSettings{
		ConnectionString:    "HostName=dummy-hub.azure-devices.net;DeviceId=dummy-device;SharedAccessKey=dGVzdGF6dXJlc2hhcmVkYWNjZXNza2V5",
		MessageMapperConfig: "../internal/testdata/handlers-mapper-config.json",
	}
	logger := logger.NewLogger(log.New(io.Discard, "", log.Ldate), logger.INFO)
	connSettings, err := config.CreateAzureConnectionSettings(settings, logger)
	require.NoError(t, err)
	messageHandler := &telemetryThingsMessageHandler{}

	messageHandler.Init(settings, connSettings)
	supportedLocalTopics := []string{"event/#", "e/#", "telemetry/#", "t/#"}
	assert.Equal(t, telemetryHandlerName, messageHandler.Name())
	assert.Equal(t, supportedLocalTopics, messageHandler.Topics())
}

func TestHandleContainerMessageTypes(t *testing.T) {
	handler := createTelemetryMessageHandler(t)
	var testData = []struct {
		jsonPayload     string
		messageType     int
		messageSubType  string
		correlationId   string
		protobufPayload string
	}{
		{
			`{
				"topic": "tenant1/dummy-device:edge:containers/things/live/messages/created",
				"headers": {
				  "response-required": false
				},
				"path": "/features/ContainerOrchestator/outbox/messages/created",
				"value": {
				  "name": "influxdb",
				  "imageRef": "docker.io/library/influxdb:latest",
				  "config": {
					"domainName": "some-domain",
					"restartPolicy": {
					  "type": "UNLESS_STOPPED"
					},
					"log": {
					  "type": "JSON_FILE",
					  "maxFiles": 2,
					  "maxSize": "100M",
					  "mode": "BLOCKING"
					}
				  },
				  "createdAt": "2021-06-03T11:52:56.614763386Z"
				}
			}`,
			1,
			"container.created",
			"",
			"CghpbmZsdXhkYhIhZG9ja2VyLmlvL2xpYnJhcnkvaW5mbHV4ZGI6bGF0ZXN0Gj4KC3NvbWUtZG9tYWluMhAaDlVOTEVTU19TVE9QUEVEWh0KCUpTT05fRklMRRACGgQxMDBNKghCTE9DS0lORyIeMjAyMS0wNi0wM1QxMTo1Mjo1Ni42MTQ3NjMzODZa",
		},
		{
			`{
				"topic": "tenant1/dummy-device:edge:containers/things/live/messages/created",
				"headers": {
				  "response-required": false,
				  "correlation-id": "some-correlation-id"
				},
				"path": "/features/ContainerOrchestator/outbox/messages/created",
				"value": {
				  "name": "influxdb",
				  "imageRef": "docker.io/library/influxdb:latest",
				  "config": {
					"domainName": "some-domain",
					"restartPolicy": {
					  "type": "UNLESS_STOPPED"
					},
					"log": {
					  "type": "JSON_FILE",
					  "maxFiles": 2,
					  "maxSize": "100M",
					  "mode": "BLOCKING"
					}
				  },
				  "createdAt": "2021-06-03T11:52:56.614763386Z"
				}
			}`,
			1,
			"container.created",
			"some-correlation-id",
			"CghpbmZsdXhkYhIhZG9ja2VyLmlvL2xpYnJhcnkvaW5mbHV4ZGI6bGF0ZXN0Gj4KC3NvbWUtZG9tYWluMhAaDlVOTEVTU19TVE9QUEVEWh0KCUpTT05fRklMRRACGgQxMDBNKghCTE9DS0lORyIeMjAyMS0wNi0wM1QxMTo1Mjo1Ni42MTQ3NjMzODZa",
		},
		{
			`{
				"topic": "tenant1/dummy-device:edge:containers/things/live/messages/removed",
				"headers": {
				  "response-required": false
				},
				"path": "/features/ContainerOrchestator/outbox/messages/removed",
				"value": {
				  "name": "influxdb"
				}
			}`,
			1,
			"container.removed",
			"",
			"CghpbmZsdXhkYg==",
		},
		{
			`{
				"topic": "tenant1/dummy-device:edge:containers/things/live/messages/stateChanged",
				"headers": {
				  "response-required": false
				},
				"path": "/features/ContainerOrchestator/outbox/messages/stateChanged",
				"value": {
				  "name": "influxdb",
				  "state": {
					"status": "RUNNING",
					"pid": 4294967295,
					"startedAt": "2021-04-29T17:37:47.15018946Z"
				  }
				}
			}`,
			1,
			"container.stateChanged",
			"",
			"CghpbmZsdXhkYhIuCgdSVU5OSU5HEP////8PKh0yMDIxLTA0LTI5VDE3OjM3OjQ3LjE1MDE4OTQ2Wg==",
		},
		{
			`{
				"topic": "tenant1/dummy-device:edge:containers/things/twin/commands/modify",
				"headers": {
				  "response-required": false
				},
				"path": "/features/ContainerOrchestrator/properties/state/status",
				"value": {
				  "manifest": {
					"name": "some-name",
					"version": "0.0.1"
				  },
				  "state": "FINISHED_ERROR",
				  "error": {
					"code": 500,
					"message": "something went wrong :("
				  }
				}
			}`,
			1,
			"container.manifest",
			"",
			"ChIKCXNvbWUtbmFtZRIFMC4wLjESDkZJTklTSEVEX0VSUk9SGhwI9AMSF3NvbWV0aGluZyB3ZW50IHdyb25nIDoo",
		},
		{
			`{
				"topic": "tenant1/dummy-device:edge:containers/things/live/messages/simplySend",
				"headers": {
					"response-required": false
				},
				"path": "/features/ContainerOrchestator/outbox/messages/simplySend",
				"value": {
					"message_id": "some-message-id-1234",
					"text": "simple text added",
					"version": "1.0.8"
				}
			}`,
			1,
			"simple.message",
			"",
			"ChRzb21lLW1lc3NhZ2UtaWQtMTIzNBIRc2ltcGxlIHRleHQgYWRkZWQaBTEuMC44",
		},
	}
	for _, testValues := range testData {
		t.Run(testValues.messageSubType, func(t *testing.T) {
			convertedMessages, err := handler.HandleMessage(createWatermillMessageForD2C([]byte(testValues.jsonPayload)))
			require.NoError(t, err)

			d2cMessage := &connectormessage.TelemetryMessage{}
			err = json.Unmarshal(convertedMessages[0].Payload, d2cMessage)
			require.NoError(t, err)

			protobufPayload, ok := d2cMessage.Payload.(string)
			assert.True(t, ok)

			assert.Equal(t, "", d2cMessage.ApplicationID)
			assert.Equal(t, testValues.messageType, d2cMessage.MessageType)
			assert.Equal(t, testValues.messageSubType, d2cMessage.MessageSubType)
			assert.Equal(t, testValues.correlationId, d2cMessage.CorrelationID)
			assert.Equal(t, "1.0.0", d2cMessage.EnvelopeVersion)
			assert.Equal(t, "1.0.0", d2cMessage.PayloadVersion)
			assert.Equal(t, testValues.protobufPayload, protobufPayload)
		})
	}
}

func TestUnsupportedMessageType(t *testing.T) {
	handler := createTelemetryMessageHandler(t)
	jsonPayload := `{
		"topic": "tenant1/dummy-device:edge:containers/things/live/messages/deleted",
		"headers": {
		  "response-required": false
		},
		"path": "/features/ContainerOrchestator/outbox/messages/deleted",
		"value": {
		  "name": "influxdb"
		}
	}`
	_, err := handler.HandleMessage(createWatermillMessageForD2C([]byte(jsonPayload)))
	require.Error(t, err)
}

func TestMessageTypeMappingWithoutSpecificDescriptorMappingFields(t *testing.T) {
	handler := createTelemetryMessageHandler(t)
	var testData = []struct {
		jsonPayload     string
		messageType     int
		messageSubType  string
		protobufPayload string
	}{
		{
			`{
				"topic": "tenant1/dummy-device:edge:containers/things/live/messages/noTopic",
				"headers": {
					"response-required": false
				},
				"path": "/features/ContainerOrchestator/outbox/messages/noTopic",
				"value": {
					"message_id": "some-message-id-1234",
					"text": "simple text added",
					"version": "1.0.8"
				}
			}`,
			1,
			"no.topic",
			"ChRzb21lLW1lc3NhZ2UtaWQtMTIzNBIRc2ltcGxlIHRleHQgYWRkZWQaBTEuMC44",
		},
		{
			`{
				"topic": "tenant1/dummy-device:edge:containers/things/live/messages/noPath",
				"headers": {
					"response-required": false
				},
				"path": "/features/ContainerOrchestator/outbox/messages/noPath",
				"value": {
					"message_id": "some-message-id-1234",
					"text": "simple text added",
					"version": "1.0.8"
				}
			}`,
			1,
			"no.path",
			"ChRzb21lLW1lc3NhZ2UtaWQtMTIzNBIRc2ltcGxlIHRleHQgYWRkZWQaBTEuMC44",
		},
	}
	for _, testValues := range testData {
		t.Run(testValues.messageSubType, func(t *testing.T) {
			convertedMessages, err := handler.HandleMessage(createWatermillMessageForD2C([]byte(testValues.jsonPayload)))
			require.NoError(t, err)

			d2cMessage := &connectormessage.TelemetryMessage{}
			err = json.Unmarshal(convertedMessages[0].Payload, d2cMessage)
			require.NoError(t, err)

			protobufPayload, ok := d2cMessage.Payload.(string)
			assert.True(t, ok)

			assert.Equal(t, testValues.messageType, d2cMessage.MessageType)
			assert.Equal(t, testValues.messageSubType, d2cMessage.MessageSubType)
			assert.Equal(t, "1.0.0", d2cMessage.EnvelopeVersion)
			assert.Equal(t, "1.0.0", d2cMessage.PayloadVersion)
			assert.Equal(t, testValues.protobufPayload, protobufPayload)
		})
	}
}

func TestNonExistingMessageTypeDescriptor(t *testing.T) {
	handler := createTelemetryMessageHandler(t)
	jsonPayload := `{
		"topic": "tenant1/dummy-device:edge:containers/things/live/messages/noDescriptor",
		"headers": {
			"response-required": false
		},
		"path": "/features/ContainerOrchestator/outbox/messages/noDescriptor",
		"value": {
			"message_id": "some-message-id-1234",
			"text": "simple text added",
			"version": "1.0.8"
		}
	}`
	_, err := handler.HandleMessage(createWatermillMessageForD2C([]byte(jsonPayload)))
	require.Error(t, err)
}

func TestNonExistingDittoMessageTopic(t *testing.T) {
	handler := createTelemetryMessageHandler(t)
	jsonPayload := `{
		"topic": "tenant1/dummy-device:edge:containers/things/live/messages/non.matching.topic",
		"headers": {
			"response-required": false
		},
		"path": "/features/ContainerOrchestator/outbox/messages/non.matching.topic",
		"value": {
			"message_id": "some-message-id-1234",
			"text": "simple text added",
			"version": "1.0.8"
		}
	}`
	_, err := handler.HandleMessage(createWatermillMessageForD2C([]byte(jsonPayload)))
	require.Error(t, err)
}

func TestNonExistingDittoMessagePath(t *testing.T) {
	handler := createTelemetryMessageHandler(t)
	jsonPayload := `{
		"topic": "tenant1/dummy-device:edge:containers/things/live/messages/non.matching.path",
		"headers": {
			"response-required": false
		},
		"path": "/features/ContainerOrchestator/outbox/messages/non.matching.path",
		"value": {
			"message_id": "some-message-id-1234",
			"text": "simple text added",
			"version": "1.0.8"
		}
	}`
	_, err := handler.HandleMessage(createWatermillMessageForD2C([]byte(jsonPayload)))
	require.Error(t, err)
}

func TestNonExistingMessageTopicKey(t *testing.T) {
	handler := createTelemetryMessageHandler(t)
	jsonPayload := `{
		"headers": {
			"response-required": false
		},
		"path": "/features/ContainerOrchestator/outbox/messages/noDescriptor",
		"value": {
			"message_id": "some-message-id-1234",
			"text": "simple text added",
			"version": "1.0.8"
		}	setUpMethod(t)
	}`
	_, err := handler.HandleMessage(createWatermillMessageForD2C([]byte(jsonPayload)))
	require.Error(t, err)
}

func TestNonExistingMessagePath(t *testing.T) {
	handler := createTelemetryMessageHandler(t)
	jsonPayload := `{
		"topic": "tenant1/dummy-device:edge:containers/things/live/messages/noDescriptor",
		"headers": {
			"response-required": false
		},
		"value": {
			"message_id": "some-message-id-1234",
			"text": "simple text added",
			"version": "1.0.8"
		}
	}`
	_, err := handler.HandleMessage(createWatermillMessageForD2C([]byte(jsonPayload)))
	require.Error(t, err)
}

func TestUnsupportedMessagePayload(t *testing.T) {
	handler := createTelemetryMessageHandler(t)
	jsonPayload := "topic=tenant1/dummy-device:edge:containers/things/live/messages/removed, name=influxdb"
	_, err := handler.HandleMessage(createWatermillMessageForD2C([]byte(jsonPayload)))
	require.Error(t, err)
}

func TestMissingDittoTopic(t *testing.T) {
	handler := createTelemetryMessageHandler(t)
	jsonPayload := `{
		"headers": {
			"response-required": false
		},
		"path": "/features/ContainerOrchestator/outbox/messages/noDescriptor",
		"value": {
			"message_id": "some-message-id-1234",
			"text": "simple text added",
			"version": "1.0.8"
		}
	}`
	_, err := handler.HandleMessage(createWatermillMessageForD2C([]byte(jsonPayload)))
	require.Error(t, err)
}

func createTelemetryMessageHandler(t *testing.T) handlers.MessageHandler {
	settings := &config.AzureSettings{
		ConnectionString:    "HostName=dummy-hub.azure-devices.net;DeviceId=dummy-device;SharedAccessKey=dGVzdGF6dXJlc2hhcmVkYWNjZXNza2V5",
		MessageMapperConfig: "../internal/testdata/handlers-mapper-config.json",
	}
	logger := logger.NewLogger(log.New(io.Discard, "", log.Ldate), logger.INFO)
	connSettings, err := config.CreateAzureConnectionSettings(settings, logger)
	require.NoError(t, err)
	messageHandler := &telemetryThingsMessageHandler{}
	messageHandler.Init(settings, connSettings)
	return messageHandler
}

func createWatermillMessageForD2C(payload []byte) *message.Message {
	watermillMessage := &message.Message{
		Payload: []byte(payload),
	}
	return watermillMessage
}
