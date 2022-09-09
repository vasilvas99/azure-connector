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

package bus

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"reflect"
	"strconv"
	"testing"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"

	"github.com/eclipse-kanto/azure-connector/config"
	"github.com/eclipse-kanto/azure-connector/routing"
	test "github.com/eclipse-kanto/azure-connector/routing/bus/internal/testing"
	routingmessage "github.com/eclipse-kanto/azure-connector/routing/message"
	"github.com/eclipse-kanto/azure-connector/routing/message/handlers/command"
	handlers "github.com/eclipse-kanto/azure-connector/routing/message/handlers/common"

	conn "github.com/eclipse-kanto/suite-connector/connector"
	"github.com/eclipse-kanto/suite-connector/logger"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	commandName             = "command-name"
	notSupportedCommandName = "not-supported"
	testCommandHandlerName  = "test_command_handler"
)

func TestRegisterCommandMessageHandler(t *testing.T) {
	settings := &config.AzureSettings{
		ConnectionString: "HostName=dummy-hub.azure-devices.net;DeviceId=dummy-device;SharedAccessKey=dGVzdGF6dXJlc2hhcmVkYWNjZXNza2V5",
	}
	logger := logger.NewLogger(log.New(io.Discard, "", log.Ldate), logger.INFO)
	connSettings, _ := config.CreateAzureConnectionSettings(settings, logger)
	router, _ := message.NewRouter(message.RouterConfig{}, watermill.NopLogger{})

	commandHandlers := []handlers.MessageHandler{}
	CommandBus(router, conn.NullPublisher(), test.NewDummySubscriber(), settings, connSettings, commandHandlers)
	refRouterPtr := reflect.ValueOf(router)
	refRouter := reflect.Indirect(refRouterPtr)
	refHandlers := refRouter.FieldByName(fieldHandlers)
	assert.Equal(t, 1, refHandlers.Len())
	refHandler := refHandlers.MapIndex(refHandlers.MapKeys()[0])
	test.AssertRouterHandler(t, commandHandlerName, routing.CreateRemoteCloudTopic("dummy-device"), "", reflect.Indirect(refHandler))
}

func TestRegisterCommandMessageHandlerInitializationError(t *testing.T) {
	settings := &config.AzureSettings{
		ConnectionString: "HostName=dummy-hub.azure-devices.net;DeviceId=dummy-device;SharedAccessKey=dGVzdGF6dXJlc2hhcmVkYWNjZXNza2V5",
	}
	logger := logger.NewLogger(log.New(io.Discard, "", log.Ldate), logger.INFO)
	connSettings, _ := config.CreateAzureConnectionSettings(settings, logger)
	router, _ := message.NewRouter(message.RouterConfig{}, watermill.NopLogger{})

	commandHandler := test.NewDummyMessageHandler(testCommandHandlerName, []string{commandName}, errors.New(""))
	commandHandlers := []handlers.MessageHandler{commandHandler}
	CommandBus(router, conn.NullPublisher(), test.NewDummySubscriber(), settings, connSettings, commandHandlers)
	refRouterPtr := reflect.ValueOf(router)
	refRouter := reflect.Indirect(refRouterPtr)
	refHandlers := refRouter.FieldByName(fieldHandlers)
	assert.Equal(t, 1, refHandlers.Len())
	refHandler := refHandlers.MapIndex(refHandlers.MapKeys()[0])
	test.AssertRouterHandler(t, commandHandlerName, routing.CreateRemoteCloudTopic("dummy-device"), "", reflect.Indirect(refHandler))
}

func TestInvalidCloudMessagePayload(t *testing.T) {
	busHandler := &commandBusHandler{}
	payload := "invalid-cloud-message-payload"
	msg := message.NewMessage(watermill.NewUUID(), message.Payload(payload))
	_, err := busHandler.HandleMessage(msg)
	require.Error(t, err)
}

func TestNoCommandMessageHandlerForMessage(t *testing.T) {
	busHandler := &commandBusHandler{}
	msg := createWatermillMessage(commandName)
	_, err := busHandler.HandleMessage(msg)
	require.Error(t, err)
}
func TestSupportedCommandName(t *testing.T) {
	commandHandler := test.NewDummyMessageHandler(testCommandHandlerName, []string{commandName}, nil)
	commandHandlers := []handlers.MessageHandler{commandHandler}
	busHandler := &commandBusHandler{commandHandlers: commandHandlers}

	outgoingMessages, err := busHandler.HandleMessage(createWatermillMessage(commandName))
	require.NoError(t, err)
	assert.NotNil(t, outgoingMessages)
	assert.Equal(t, len(outgoingMessages), 1)
	assert.NotNil(t, command.MessageFromContext(outgoingMessages[0]))
}

func TestFirstMatchedCommandMessageHandler(t *testing.T) {
	var commandHandlers []handlers.MessageHandler
	commandNames := []string{"command.name.1", "command.name.2", "command.name.3"}
	for i, commandName := range commandNames {
		commandHandler := test.NewDummyMessageHandler(testCommandHandlerName+"_"+strconv.Itoa(i+1), []string{commandName}, nil)
		commandHandlers = append(commandHandlers, commandHandler)
	}
	busHandler := &commandBusHandler{commandHandlers: commandHandlers}

	outgoingMessages, err := busHandler.HandleMessage(createWatermillMessage("command.name.2"))
	require.NoError(t, err)
	assert.NotNil(t, outgoingMessages)
	assert.Equal(t, len(outgoingMessages), 1)
	assert.NotNil(t, command.MessageFromContext(outgoingMessages[0]))
	assert.Equal(t, "test_command_handler_2", outgoingMessages[0].Metadata["handler_name"])
}

func TestMultipleCommandMessageHandlersForCommand(t *testing.T) {
	var commandHandlers []handlers.MessageHandler
	for i := 0; i < 3; i++ {
		commandHandler := test.NewDummyMessageHandler(testCommandHandlerName+"_"+strconv.Itoa(i+1), []string{commandName}, nil)
		commandHandlers = append(commandHandlers, commandHandler)
	}
	busHandler := &commandBusHandler{commandHandlers: commandHandlers}

	outgoingMessages, err := busHandler.HandleMessage(createWatermillMessage(commandName))
	require.NoError(t, err)
	assert.NotNil(t, outgoingMessages)
	assert.Equal(t, len(outgoingMessages), 1)
	assert.NotNil(t, command.MessageFromContext(outgoingMessages[0]))
	assert.Equal(t, "test_command_handler_1", outgoingMessages[0].Metadata["handler_name"])
}

func TestNotSupportedCommandName(t *testing.T) {
	commandHandler := test.NewDummyMessageHandler(testCommandHandlerName, []string{commandName}, nil)
	commandHandlers := []handlers.MessageHandler{commandHandler}
	busHandler := &commandBusHandler{commandHandlers: commandHandlers}
	_, err := busHandler.HandleMessage(createWatermillMessage(notSupportedCommandName))
	require.Error(t, err)
}

func createWatermillMessage(commandName string) *message.Message {
	cloudMessage := &routingmessage.CloudMessage{
		ApplicationID:   "app1",
		CommandName:     commandName,
		EnvelopeVersion: "1.0.0",
		PayloadVersion:  "1.0.0",
		Payload:         "{}",
	}
	jsonPayload, _ := json.Marshal(cloudMessage)
	return message.NewMessage(watermill.NewUUID(), message.Payload(jsonPayload))
}
