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

package protobuf

import (
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/eclipse-kanto/azure-connector/routing/message/config"
	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/desc/protoparse"
	"github.com/jhump/protoreflect/dynamic"
	"github.com/pkg/errors"
)

// Marshaller is an interface for marshalling/unmarshalling C2D & D2C messages to/from protobuf message payload.
type Marshaller interface {
	Marshal(messageType int, messageSubType string, payload []byte) ([]byte, error)
	Unmarshal(messageType string, protobufPayload string) ([]byte, error)
}

type jsonProtobufMarshaller struct {
	mapperConfig                *config.MessageMapperConfig
	commandMessageDescriptors   map[string]*desc.MessageDescriptor
	telemetryMessageDescriptors map[int]map[string]*desc.MessageDescriptor
}

// NewProtobufJSONMarshaller creates a protobuf marshaller instance.
func NewProtobufJSONMarshaller(mapperConfig *config.MessageMapperConfig) Marshaller {
	return &jsonProtobufMarshaller{
		mapperConfig:                mapperConfig,
		commandMessageDescriptors:   make(map[string]*desc.MessageDescriptor),
		telemetryMessageDescriptors: make(map[int]map[string]*desc.MessageDescriptor),
	}
}

func (m *jsonProtobufMarshaller) Marshal(messageType int, messageSubType string, jsonPayload []byte) ([]byte, error) {
	errorMsg := "cannot serialize D2C message payload to protobuf format for message type '%v' and message subtype '%s'"
	dynamicMessage, err := m.getD2CProtoMessage(messageType, messageSubType)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf(errorMsg, messageType, messageSubType))
	}
	err = dynamicMessage.UnmarshalJSON(jsonPayload)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf(errorMsg, messageType, messageSubType))
	}
	protobufPayload, err := dynamicMessage.Marshal()
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf(errorMsg, messageType, messageSubType))
	}
	return protobufPayload, nil
}

func (m *jsonProtobufMarshaller) Unmarshal(messageType string, payload string) ([]byte, error) {
	errorMsg := "Cannot deserialize C2D message protobuf payload format to JSON for message type '%s'!"
	dynamicMessage, err := m.getC2DProtoMessage(messageType)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf(errorMsg, messageType))
	}
	decodedPayload, err := base64.StdEncoding.DecodeString(payload)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf(errorMsg, messageType))
	}
	err = dynamicMessage.Unmarshal(decodedPayload)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf(errorMsg, messageType))
	}
	jsonPayload, err := dynamicMessage.MarshalJSON()
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf(errorMsg, messageType))
	}
	return jsonPayload, nil
}

func (m *jsonProtobufMarshaller) getD2CProtoMessage(messageType int, messageSubType string) (*dynamic.Message, error) {
	messageDescriptor, err := m.getTelemetryMessageDescriptor(messageType, messageSubType)
	if err != nil {
		return nil, err
	}
	return dynamic.NewMessage(messageDescriptor), nil
}

func (m *jsonProtobufMarshaller) getC2DProtoMessage(messageType string) (*dynamic.Message, error) {
	messageDescriptor, err := m.getCommandMessageDescriptor(messageType)
	if err != nil {
		return nil, err
	}
	return dynamic.NewMessage(messageDescriptor), nil
}

func (m *jsonProtobufMarshaller) getTelemetryMessageDescriptor(messageType int, messageSubType string) (*desc.MessageDescriptor, error) {
	messageSubTypeDescriptors, ok := m.telemetryMessageDescriptors[messageType]
	if ok {
		if messageDescriptor, ok := messageSubTypeDescriptors[messageSubType]; ok {
			return messageDescriptor, nil
		}
		messageMapping, err := m.mapperConfig.GetTelemetryMessageMapping(messageType, messageSubType)
		if messageMapping == nil {
			return nil, err
		}
		messageDescriptor, err := m.loadMessageDescriptor(messageMapping.ProtoMessage, messageMapping.ProtoFile)
		if err != nil {
			return nil, err
		}
		messageSubTypeDescriptors[messageSubType] = messageDescriptor
		return messageDescriptor, nil
	}
	messageMapping, err := m.mapperConfig.GetTelemetryMessageMapping(messageType, messageSubType)
	if messageMapping == nil {
		return nil, err
	}
	messageDescriptor, err := m.loadMessageDescriptor(messageMapping.ProtoMessage, messageMapping.ProtoFile)
	if err != nil {
		return nil, err
	}
	messageSubTypeDescriptors = make(map[string]*desc.MessageDescriptor)
	messageSubTypeDescriptors[messageSubType] = messageDescriptor
	m.telemetryMessageDescriptors[messageType] = messageSubTypeDescriptors
	return messageDescriptor, nil
}

func (m *jsonProtobufMarshaller) getCommandMessageDescriptor(messageType string) (*desc.MessageDescriptor, error) {
	if messageDescriptor, ok := m.commandMessageDescriptors[messageType]; ok {
		return messageDescriptor, nil
	}
	messageMapping, err := m.mapperConfig.GetCommandMessageMapping(messageType)
	if messageMapping == nil {
		return nil, err
	}
	messageDescriptor, err := m.loadMessageDescriptor(messageMapping.ProtoMessage, messageMapping.ProtoFile)
	if err != nil {
		return nil, err
	}
	m.commandMessageDescriptors[messageType] = messageDescriptor
	return messageDescriptor, nil
}

func (m *jsonProtobufMarshaller) loadMessageDescriptor(protoMessage, protoFile string) (*desc.MessageDescriptor, error) {
	index := strings.LastIndex(protoFile, "/")
	var protoFilesPath string
	if index == -1 {
		protoFilesPath = protoFile
	} else {
		protoFilesPath = protoFile[:index]
	}
	fileAccessor := func(fileName string) (io.ReadCloser, error) {
		if !strings.HasPrefix(fileName, protoFilesPath) {
			fileName = protoFilesPath + "/" + fileName
		}
		return os.Open(fileName)
	}
	parser := protoparse.Parser{Accessor: fileAccessor}
	fileDescriptors, err := parser.ParseFiles(protoFile)
	if err != nil {
		return nil, err
	}
	messageDescriptors := fileDescriptors[0].GetMessageTypes()
	if len(messageDescriptors) == 1 {
		if len(protoMessage) == 0 || messageDescriptors[0].GetName() == protoMessage {
			return messageDescriptors[0], nil
		}
		return nil, errors.New(fmt.Sprintf("no proto message '%s' in proto file '%s'", protoMessage, protoFile))
	}
	for _, messageDescriptor := range messageDescriptors {
		if messageDescriptor.GetName() == protoMessage {
			return messageDescriptor, nil
		}
	}
	return nil, errors.New(fmt.Sprintf("no proto message '%s' in proto file '%s'", protoMessage, protoFile))
}
