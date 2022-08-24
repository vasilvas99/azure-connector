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

package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/pkg/errors"
)

// MessageMapperConfig represents the configuration data for the message mappings.
type MessageMapperConfig struct {
	MessageMappings *MessageMappings `json:"messageMappings,omitempty"`
}

// MessageMappings represents the message mappings.
type MessageMappings struct {
	Command   map[string]*CommandMessageMapping           `json:"command,omitempty"`
	Telemetry map[int]map[string]*TelemetryMessageMapping `json:"telemetry,omitempty"`
}

// CommandMessageMapping contains the configuration data for a command message mapping.
type CommandMessageMapping struct {
	ProtoFile         string                    `json:"protoFile,omitempty"`
	ProtoMessage      string                    `json:"protoMessage,omitempty"`
	MappingProperties *CommandMappingProperties `json:"dittoMapping,omitempty"`
}

// TelemetryMessageMapping contains the configuration data for a telemetry message mapping.
type TelemetryMessageMapping struct {
	ProtoFile         string                      `json:"protoFile,omitempty"`
	ProtoMessage      string                      `json:"protoMessage,omitempty"`
	MappingProperties *TelemetryMappingProperties `json:"dittoMapping,omitempty"`
}

// CommandMappingProperties defines the mapping properties for a command message mapping.
type CommandMappingProperties struct {
	Thing  string `json:"thing,omitempty"`
	Action string `json:"action,omitempty"`
	Path   string `json:"path,omitempty"`
}

// TelemetryMappingProperties defines the mapping properties for a telemetry message mapping.
type TelemetryMappingProperties struct {
	Topic string `json:"topic,omitempty"`
	Path  string `json:"path,omitempty"`
}

// GetCommandMessageMapping returns the command message mapping for a specific command name.
func (config *MessageMapperConfig) GetCommandMessageMapping(messageType string) (*CommandMessageMapping, error) {
	commandMappings, err := config.GetCommandMessageMappings()
	if err != nil {
		return nil, err
	}
	messageMapping := commandMappings[messageType]
	if messageMapping == nil {
		return nil, errors.New(fmt.Sprintf("no supported command message mapping configuration for message type '%s'", messageType))
	}
	return messageMapping, nil
}

// GetCommandMessageMappings returns all available command message mappings.
func (config *MessageMapperConfig) GetCommandMessageMappings() (map[string]*CommandMessageMapping, error) {
	messageMappings := config.MessageMappings
	if messageMappings == nil {
		return nil, errors.New("no message mapping configurations")
	}
	commandMappings := messageMappings.Command
	if commandMappings == nil {
		return nil, errors.New("no command message mapping configurations")
	}
	return commandMappings, nil
}

// GetTelemetryMessageMapping returns the telemetry message mapping for a specific message type + sub type pair.
func (config *MessageMapperConfig) GetTelemetryMessageMapping(messageType int, messageSubType string) (*TelemetryMessageMapping, error) {
	telemetryMappings, err := config.GetTelemetryMessageMappings()
	if err != nil {
		return nil, errors.New("no telemetry message mapping configurations")
	}
	telemetryMessageTypeMappings := telemetryMappings[messageType]
	if telemetryMessageTypeMappings == nil {
		return nil, errors.New(fmt.Sprintf("no telemetry message mapping configurations for message type '%v'", messageType))
	}
	messageMapping := telemetryMessageTypeMappings[messageSubType]
	if messageMapping == nil {
		return nil, errors.New(fmt.Sprintf("no supported telemetry message mapping configuration for message type '%v' and message subtype '%s'", messageType, messageSubType))
	}
	return messageMapping, nil
}

// GetTelemetryMessageMappings returns all available telemetry message mappings.
func (config *MessageMapperConfig) GetTelemetryMessageMappings() (map[int]map[string]*TelemetryMessageMapping, error) {
	messageMappings := config.MessageMappings
	if messageMappings == nil {
		return nil, errors.New("no message mapping configurations")
	}
	telemetryMappings := messageMappings.Telemetry
	if telemetryMappings == nil {
		return nil, errors.New("no telemetry message mapping configurations")
	}
	return telemetryMappings, nil
}

// LoadMessageMapperConfig loads the message mappings configuration data from a file on the file system.
func LoadMessageMapperConfig(mapperConfigFile string) (*MessageMapperConfig, error) {
	jsonContent, err := ioutil.ReadFile(mapperConfigFile)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("cannot load message mapper config file '%s'", mapperConfigFile))
	}
	config := &MessageMapperConfig{}
	err = json.Unmarshal(jsonContent, config)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("cannot parse message mapper config file '%s'", mapperConfigFile))
	}
	return config, nil
}
