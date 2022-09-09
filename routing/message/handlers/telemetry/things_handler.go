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

package telemetry

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/eclipse-kanto/azure-connector/config"
	"github.com/eclipse-kanto/azure-connector/routing"
	routingmessage "github.com/eclipse-kanto/azure-connector/routing/message"
	mapperconfig "github.com/eclipse-kanto/azure-connector/routing/message/config"
	"github.com/eclipse-kanto/azure-connector/routing/message/protobuf"
	"github.com/eclipse-kanto/suite-connector/connector"
	"github.com/eclipse/ditto-clients-golang/protocol"
	"github.com/pkg/errors"
)

var localTopics []string = []string{"event/#", "e/#", "telemetry/#", "t/#"}

const (
	envelopeVersion      = "1.0.0"
	payloadVersion       = "1.0.0"
	telemetryHandlerName = "things_telemetry_handler"
)

type telemetryThingsMessageHandler struct {
	settings     *config.AzureSettings
	connSettings *config.AzureConnectionSettings
	mapperConfig *mapperconfig.MessageMapperConfig
	marshaller   protobuf.Marshaller
}

func (h *telemetryThingsMessageHandler) Init(settings *config.AzureSettings, connSettings *config.AzureConnectionSettings) error {
	h.settings = settings
	h.connSettings = connSettings
	messageMapperConfig, err := mapperconfig.LoadMessageMapperConfig(h.settings.MessageMapperConfig)
	if err != nil {
		return err
	}
	h.mapperConfig = messageMapperConfig
	h.marshaller = protobuf.NewProtobufJSONMarshaller(h.mapperConfig)
	return nil
}

func (h *telemetryThingsMessageHandler) HandleMessage(msg *message.Message) ([]*message.Message, error) {
	dittoMessage := &protocol.Envelope{}
	if err := json.Unmarshal(msg.Payload, dittoMessage); err != nil {
		return nil, errors.Wrap(err, "cannot deserialize Ditto message!")
	}

	messageType, messageSubType, err := getMessageTypes(dittoMessage, h.mapperConfig)
	if err != nil {
		return nil, err
	}

	// TODO: quick workaround for ditto lib
	data, err := json.Marshal(dittoMessage.Value)
	if err != nil {
		data = make([]byte, 0)
	}
	protobufPayload, err := h.marshaller.Marshal(messageType, messageSubType, data)
	if err != nil {
		return nil, err
	}

	d2cMessage := &routingmessage.TelemetryMessage{
		MessageType:     messageType,
		MessageSubType:  messageSubType,
		Timestamp:       getUnixTimestampMs(),
		EnvelopeVersion: envelopeVersion,
		PayloadVersion:  payloadVersion,
		Payload:         protobufPayload,
	}

	correlationID := dittoMessage.Headers.CorrelationID()
	if len(correlationID) > 0 {
		d2cMessage.CorrelationID = correlationID
	}

	outgoingPayload, err := json.Marshal(d2cMessage)
	if err != nil {
		return nil, errors.Wrap(err, "cannot serialize D2C message")
	}

	msgID := watermill.NewUUID()
	outgoingMessage := message.NewMessage(msgID, outgoingPayload)
	outgoingTopic := routing.CreateTelemetryTopic(h.connSettings.DeviceID, msgID)
	outgoingMessage.SetContext(connector.SetTopicToCtx(outgoingMessage.Context(), outgoingTopic))
	return []*message.Message{outgoingMessage}, nil
}

func getMessageTypes(dittoMessage *protocol.Envelope, mapperConfig *mapperconfig.MessageMapperConfig) (int, string, error) {
	if dittoMessage.Topic == nil {
		return -1, "", errors.New("missing Ditto topic in message")
	}

	topic := dittoMessage.Topic.String()
	path := dittoMessage.Path
	telemetryMappings, err := mapperConfig.GetTelemetryMessageMappings()
	if err != nil {
		return -1, "", errors.Wrap(err, fmt.Sprintf("cannot map Ditto topic '%s' & Ditto path '%s' to D2C message sub type", topic, path))
	}

	for messageType, telemetryMessageTypeMappings := range telemetryMappings {
		for messageSubType, messageMapping := range telemetryMessageTypeMappings {
			mappingProperties := messageMapping.MappingProperties
			if mappingProperties.Topic != "" {
				if mappingProperties.Path != "" {
					if strings.Contains(topic, mappingProperties.Topic) && strings.Contains(path, mappingProperties.Path) {
						return messageType, messageSubType, nil
					}
				} else {
					if strings.Contains(topic, mappingProperties.Topic) {
						return messageType, messageSubType, nil
					}
				}
			} else if mappingProperties.Path != "" {
				if strings.Contains(path, mappingProperties.Path) {
					return messageType, messageSubType, nil
				}
			}
		}
	}
	return -1, "", fmt.Errorf("cannot map Ditto topic '%s' & Ditto path '%s' to D2C message sub type", topic, path)
}

func getUnixTimestampMs() int64 {
	return time.Now().UnixNano() / (int64(time.Millisecond) / int64(time.Nanosecond))
}

func (h *telemetryThingsMessageHandler) Name() string {
	return telemetryHandlerName
}

func (h *telemetryThingsMessageHandler) Topics() []string {
	return localTopics
}

func init() {
	registerMessageHandler(&telemetryThingsMessageHandler{})
}
