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

package config_test

import (
	"testing"

	"github.com/eclipse-kanto/azure-connector/routing/message/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEmptyMessageMapperConfigFile(t *testing.T) {
	_, err := config.LoadMessageMapperConfig("")
	require.Error(t, err)
}

func TestInvalidMessageMapperConfigFile(t *testing.T) {
	_, err := config.LoadMessageMapperConfig("testdata/invalid-message-mapper-config.json")
	require.Error(t, err)
}

func TestEmptyMappings(t *testing.T) {
	mapperConfig, err := config.LoadMessageMapperConfig("testdata/empty-mappings-config.json")
	require.NoError(t, err)
	assertConfigGettersError(t, mapperConfig)
}

func TestInvalidJSONMappingsFile(t *testing.T) {
	_, err := config.LoadMessageMapperConfig("testdata/invalid-json-config.json")
	require.Error(t, err)
}

func TestMissingMappings(t *testing.T) {
	mapperConfig, err := config.LoadMessageMapperConfig("testdata/missing-mappings-config.json")
	require.NoError(t, err)
	assertConfigGettersError(t, mapperConfig)
}

func assertConfigGettersError(t *testing.T, mapperConfig *config.MessageMapperConfig) {
	_, err := mapperConfig.GetCommandMessageMappings()
	require.Error(t, err)
	_, err = mapperConfig.GetCommandMessageMapping("command-mapping")
	require.Error(t, err)
	_, err = mapperConfig.GetTelemetryMessageMappings()
	require.Error(t, err)
	_, err = mapperConfig.GetTelemetryMessageMapping(1, "telemetry-mapping")
	require.Error(t, err)
}

func TestMissingCommandMappings(t *testing.T) {
	mapperConfig, err := config.LoadMessageMapperConfig("testdata/missing-command-mappings.json")
	require.NoError(t, err)
	_, err = mapperConfig.GetCommandMessageMappings()
	require.Error(t, err)
	_, err = mapperConfig.GetCommandMessageMapping("command-mapping")
	require.Error(t, err)
}

func TestMissingTelemetryMappings(t *testing.T) {
	mapperConfig, err := config.LoadMessageMapperConfig("testdata/missing-telemetry-mappings.json")
	require.NoError(t, err)
	_, err = mapperConfig.GetTelemetryMessageMappings()
	require.Error(t, err)
	_, err = mapperConfig.GetTelemetryMessageMapping(1, "telemetry-mapping")
	require.Error(t, err)
}

func TestEmptyTelemetryMessageTypeMappings(t *testing.T) {
	mapperConfig, err := config.LoadMessageMapperConfig("testdata/empty-telemetry-message-types.json")
	require.NoError(t, err)
	_, err = mapperConfig.GetTelemetryMessageMappings()
	require.NoError(t, err)
	_, err = mapperConfig.GetTelemetryMessageMapping(1, "telemetry-mapping")
	require.Error(t, err)
}

func TestEmptyTelemetryMappings(t *testing.T) {
	mapperConfig, err := config.LoadMessageMapperConfig("testdata/empty-telemetry-mappings.json")
	require.NoError(t, err)
	_, err = mapperConfig.GetTelemetryMessageMappings()
	require.NoError(t, err)
	_, err = mapperConfig.GetTelemetryMessageMapping(1, "telemetry-mapping")
	require.Error(t, err)
}

func TestTelemetryMessageMappings(t *testing.T) {
	mapperConfig, err := config.LoadMessageMapperConfig("testdata/message-mappings.json")
	require.NoError(t, err)
	telemetryMessageTypeMappings, err := mapperConfig.GetTelemetryMessageMappings()
	require.NoError(t, err)
	telemetryMappings := telemetryMessageTypeMappings[1]
	expectedMappings := make(map[string]*config.TelemetryMessageMapping)
	expectedMappings["telemetry-mapping"] = &config.TelemetryMessageMapping{
		ProtoFile:    "dummy_message.proto",
		ProtoMessage: "DummyMessage",
		MappingProperties: &config.TelemetryMappingProperties{
			Topic: "/telemetry",
			Path:  "/outbox/messages/action",
		},
	}
	expectedMappings["missing-proto-message"] = &config.TelemetryMessageMapping{
		ProtoFile: "dummy_message.proto",
		MappingProperties: &config.TelemetryMappingProperties{
			Topic: "/missing_proto_message",
			Path:  "/outbox/messages/action",
		},
	}
	expectedMappings["missing-mapping-topic"] = &config.TelemetryMessageMapping{
		ProtoFile:         "dummy_message.proto",
		ProtoMessage:      "DummyMessage",
		MappingProperties: &config.TelemetryMappingProperties{Path: "/outbox/messages/action"},
	}
	expectedMappings["missing-mapping-path"] = &config.TelemetryMessageMapping{
		ProtoFile:         "dummy_message.proto",
		ProtoMessage:      "DummyMessage",
		MappingProperties: &config.TelemetryMappingProperties{Topic: "/missing_mapping_path"},
	}
	assertTelemetryMappings(t, expectedMappings, telemetryMappings)

	telemetryMapping, err := mapperConfig.GetTelemetryMessageMapping(1, "telemetry-mapping")
	require.NoError(t, err)
	assert.Equal(t, "dummy_message.proto", telemetryMapping.ProtoFile)
	assert.Equal(t, "DummyMessage", telemetryMapping.ProtoMessage)
	assertTelemetryMappingProperties(t, expectedMappings["telemetry-mapping"].MappingProperties, telemetryMapping.MappingProperties)

	_, err = mapperConfig.GetTelemetryMessageMapping(1, "invalid-telemetry-mapping")
	require.Error(t, err)
	_, err = mapperConfig.GetTelemetryMessageMapping(2, "invalid-telemetry-mapping")
	require.Error(t, err)
}

func assertTelemetryMappings(t *testing.T, expectedMappings, telemetryMappings map[string]*config.TelemetryMessageMapping) {
	assert.Equal(t, len(expectedMappings), len(telemetryMappings))
	for messageType, expectedMapping := range expectedMappings {
		telemetryMapping := telemetryMappings[messageType]
		assert.NotEqual(t, telemetryMapping, nil)
		assert.Equal(t, expectedMapping.ProtoFile, telemetryMapping.ProtoFile)
		assert.Equal(t, expectedMapping.ProtoMessage, telemetryMapping.ProtoMessage)
		assertTelemetryMappingProperties(t, expectedMapping.MappingProperties, telemetryMapping.MappingProperties)
	}
}

func assertTelemetryMappingProperties(t *testing.T, expectedMappingProperties, telemetryMappingProperties *config.TelemetryMappingProperties) {
	assert.Equal(t, expectedMappingProperties.Topic, telemetryMappingProperties.Topic)
	assert.Equal(t, expectedMappingProperties.Path, telemetryMappingProperties.Path)
}

func TestCommandMessageMappings(t *testing.T) {
	mapperConfig, err := config.LoadMessageMapperConfig("testdata/message-mappings.json")
	require.NoError(t, err)
	commandMappings, err := mapperConfig.GetCommandMessageMappings()
	require.NoError(t, err)
	expectedMappings := make(map[string]*config.CommandMessageMapping)
	expectedMappings["command-mapping"] = &config.CommandMessageMapping{
		ProtoFile:    "dummy_message.proto",
		ProtoMessage: "DummyMessage",
		MappingProperties: &config.CommandMappingProperties{
			Thing:  "command",
			Path:   "/outbox/messages/action",
			Action: "action",
		},
	}
	expectedMappings["missing-proto-message"] = &config.CommandMessageMapping{
		ProtoFile: "dummy_message.proto",
		MappingProperties: &config.CommandMappingProperties{
			Thing:  "missing_proto_message",
			Path:   "/outbox/messages/action",
			Action: "action",
		},
	}
	expectedMappings["missing-mapping-thing"] = &config.CommandMessageMapping{
		ProtoFile:    "dummy_message.proto",
		ProtoMessage: "DummyMessage",
		MappingProperties: &config.CommandMappingProperties{
			Path:   "/outbox/messages/action",
			Action: "action",
		},
	}
	expectedMappings["missing-mapping-path"] = &config.CommandMessageMapping{
		ProtoFile:    "dummy_message.proto",
		ProtoMessage: "DummyMessage",
		MappingProperties: &config.CommandMappingProperties{
			Thing:  "missing_mapping_path",
			Action: "action",
		},
	}
	expectedMappings["missing-mapping-action"] = &config.CommandMessageMapping{
		ProtoFile:    "dummy_message.proto",
		ProtoMessage: "DummyMessage",
		MappingProperties: &config.CommandMappingProperties{
			Thing: "missing_mapping_action",
			Path:  "/outbox/messages/action",
		},
	}
	assertCommandMappings(t, expectedMappings, commandMappings)

	commandMapping, err := mapperConfig.GetCommandMessageMapping("command-mapping")
	require.NoError(t, err)
	assert.Equal(t, "dummy_message.proto", commandMapping.ProtoFile)
	assert.Equal(t, "DummyMessage", commandMapping.ProtoMessage)
	assertCommandMappingProperties(t, expectedMappings["command-mapping"].MappingProperties, commandMapping.MappingProperties)

	_, err = mapperConfig.GetCommandMessageMapping("invalid-command-mapping")
	require.Error(t, err)
}

func assertCommandMappings(t *testing.T, expectedMappings, commandMappings map[string]*config.CommandMessageMapping) {
	assert.Equal(t, len(expectedMappings), len(commandMappings))
	for messageType, expectedMapping := range expectedMappings {
		commandMapping := commandMappings[messageType]
		assert.NotEqual(t, commandMapping, nil)
		assert.Equal(t, expectedMapping.ProtoFile, commandMapping.ProtoFile)
		assert.Equal(t, expectedMapping.ProtoMessage, commandMapping.ProtoMessage)
		assertCommandMappingProperties(t, expectedMapping.MappingProperties, commandMapping.MappingProperties)
	}
}

func assertCommandMappingProperties(t *testing.T, expectedMappingProperties, commandMappingProperties *config.CommandMappingProperties) {
	assert.Equal(t, expectedMappingProperties.Thing, commandMappingProperties.Thing)
	assert.Equal(t, expectedMappingProperties.Path, commandMappingProperties.Path)
	assert.Equal(t, expectedMappingProperties.Action, commandMappingProperties.Action)
}
