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
	"fmt"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/eclipse-kanto/azure-connector/config"
	mapperconfig "github.com/eclipse-kanto/azure-connector/routing/message/config"
	"github.com/eclipse-kanto/azure-connector/routing/message/protobuf"
	"github.com/eclipse-kanto/suite-connector/connector"
	"github.com/eclipse/ditto-clients-golang/protocol"
	"github.com/pkg/errors"
)

const (
	commandThingsHandlerName     = "command_things_handler"
	dittoNamespace               = "azure.edge"
	messageTopicPattern          = "command///req/%s/%s"
	messageTopicPatternWithThing = "command//%s:%s:%s/req/%s/%s"
	dittoTopicPattern            = `"%s/%s/things/live/messages/%s"`
	dittoTopicPatternWithThing   = `"%s/%s:%s/things/live/messages/%s"`
)

type commandThingsMessageHandler struct {
	settings     *config.AzureSettings
	connSettings *config.AzureConnectionSettings
	mapperConfig *mapperconfig.MessageMapperConfig
	marshaller   protobuf.Marshaller
	topics       []string
}

func (h *commandThingsMessageHandler) Init(settings *config.AzureSettings, connSettings *config.AzureConnectionSettings) error {
	h.settings = settings
	h.connSettings = connSettings
	messageMapperConfig, err := mapperconfig.LoadMessageMapperConfig(h.settings.MessageMapperConfig)
	if err != nil {
		return err
	}
	h.mapperConfig = messageMapperConfig
	h.marshaller = protobuf.NewProtobufJSONMarshaller(h.mapperConfig)
	h.topics = getSupportedCommandNames(h.mapperConfig)
	return nil
}

func (h *commandThingsMessageHandler) HandleMessage(msg *message.Message) ([]*message.Message, error) {
	cloudMessage := GetCommandMessageFromContext(msg)
	if cloudMessage == nil {
		return nil, errors.New("cannot deserialize cloud message")
	}
	messageMapping, err := h.mapperConfig.GetCommandMessageMapping(cloudMessage.CommandName)
	if err != nil {
		return nil, err
	}
	mappingProperties := messageMapping.MappingProperties

	deviceID := h.connSettings.HubName + ":" + h.connSettings.DeviceID
	topicJSONStr := createDittoTopic(mappingProperties, deviceID)
	var topic *protocol.Topic
	json.Unmarshal([]byte(topicJSONStr), &topic)

	headers := protocol.NewHeaders(
		protocol.WithContentType("application/json"),
		protocol.WithCorrelationID(cloudMessage.CorrelationID))

	payload, _ := cloudMessage.Payload.(string)
	payloadFromProtobuf, err := h.marshaller.Unmarshal(cloudMessage.CommandName, payload)
	if err != nil {
		return nil, err
	}

	dittoMessage := &protocol.Envelope{
		Topic:   topic,
		Headers: headers,
		Path:    mappingProperties.Path,
		Value:   payloadFromProtobuf,
	}

	outgoingPayload, err := json.Marshal(dittoMessage)
	if err != nil {
		return nil, errors.Wrap(err, "cannot serialize C2D message")
	}
	outgoingMessage := message.NewMessage(watermill.NewUUID(), outgoingPayload)
	outgoingTopic := createMessageTopic(mappingProperties, deviceID, cloudMessage.CorrelationID)
	outgoingMessage.SetContext(connector.SetTopicToCtx(outgoingMessage.Context(), outgoingTopic))

	return []*message.Message{outgoingMessage}, nil
}

func createMessageTopic(mappingProperties *mapperconfig.CommandMappingProperties, deviceID, reqID string) string {
	if mappingProperties.Thing == "" {
		return fmt.Sprintf(messageTopicPattern, reqID, mappingProperties.Action)
	}
	return fmt.Sprintf(messageTopicPatternWithThing, dittoNamespace, deviceID, mappingProperties.Thing, reqID, mappingProperties.Action)
}

func createDittoTopic(mappingProperties *mapperconfig.CommandMappingProperties, deviceID string) string {
	if mappingProperties.Thing == "" {
		return fmt.Sprintf(dittoTopicPattern, dittoNamespace, deviceID, mappingProperties.Action)
	}
	return fmt.Sprintf(dittoTopicPatternWithThing, dittoNamespace, deviceID, mappingProperties.Thing, mappingProperties.Action)
}

func getSupportedCommandNames(mapperConfig *mapperconfig.MessageMapperConfig) []string {
	commandNames := []string{}
	for commandName := range mapperConfig.MessageMappings.Command {
		commandNames = append(commandNames, commandName)
	}
	return commandNames
}

func (h *commandThingsMessageHandler) Name() string {
	return commandThingsHandlerName
}

func (h *commandThingsMessageHandler) Topics() []string {
	return h.topics
}

func init() {
	registerMessageHandler(&commandThingsMessageHandler{})
}
