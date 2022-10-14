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
	"errors"
	"reflect"
	"testing"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"

	"github.com/eclipse-kanto/azure-connector/routing"
	test "github.com/eclipse-kanto/azure-connector/routing/bus/internal/testing"
	"github.com/eclipse-kanto/azure-connector/routing/message/handlers"

	conn "github.com/eclipse-kanto/suite-connector/connector"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	commandName             = "command-name"
	notSupportedCommandName = "not-supported"
	testCommandHandlerName  = "test_command_handler"
)

func TestRegisterCommandMessageHandler(t *testing.T) {
	router, connInfo := setupTestRouter("dummy-device")

	commandHandlers := []handlers.CommandHandler{}
	CommandBus(router, conn.NullPublisher(), test.NewDummySubscriber(), connInfo, commandHandlers)
	refRouterPtr := reflect.ValueOf(router)
	refRouter := reflect.Indirect(refRouterPtr)
	refHandlers := refRouter.FieldByName(fieldHandlers)
	assert.Equal(t, 1, refHandlers.Len())
	refHandler := refHandlers.MapIndex(refHandlers.MapKeys()[0])
	test.AssertRouterHandler(t, commandHandlerName, routing.CreateRemoteCloudTopic("dummy-device"), "", reflect.Indirect(refHandler))
}

func TestRegisterCommandMessageHandlerInitializationError(t *testing.T) {
	router, connInfo := setupTestRouter("dummy-device")

	commandHandler := test.NewDummyCommandHandler(testCommandHandlerName, errors.New(""), nil)
	commandHandlers := []handlers.CommandHandler{commandHandler}
	CommandBus(router, conn.NullPublisher(), test.NewDummySubscriber(), connInfo, commandHandlers)
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

func TestNoCommandHandlerForMessage(t *testing.T) {
	busHandler := &commandBusHandler{}
	_, err := busHandler.HandleMessage(message.NewMessage(watermill.NewUUID(), message.Payload("dummy_payload")))
	require.Error(t, err)
}

func TestFirstValidCommandMessageHandler(t *testing.T) {
	commandHandler1 := test.NewDummyCommandHandler(testCommandHandlerName+"_1", nil, errors.New(""))
	commandHandler2 := test.NewDummyCommandHandler(testCommandHandlerName+"_2", nil, nil)
	commandHandler3 := test.NewDummyCommandHandler(testCommandHandlerName+"_3", nil, errors.New(""))
	commandHandlers := []handlers.CommandHandler{commandHandler1, commandHandler2, commandHandler3}

	busHandler := &commandBusHandler{logger: watermill.NopLogger{}, commandHandlers: commandHandlers}

	outgoingMessages, err := busHandler.HandleMessage(message.NewMessage(watermill.NewUUID(), message.Payload("dummy_payload")))
	require.NoError(t, err)
	assert.NotNil(t, outgoingMessages)
	assert.Equal(t, len(outgoingMessages), 1)
	assert.Equal(t, "test_command_handler_2", outgoingMessages[0].Metadata["handler_name"])
}

func TestMultipleCommandMessageHandlers(t *testing.T) {
	commandHandler1 := test.NewDummyCommandHandler(testCommandHandlerName+"_1", nil, nil)
	commandHandler2 := test.NewDummyCommandHandler(testCommandHandlerName+"_2", nil, nil)
	commandHandler3 := test.NewDummyCommandHandler(testCommandHandlerName+"_3", nil, nil)
	commandHandlers := []handlers.CommandHandler{commandHandler1, commandHandler2, commandHandler3}

	busHandler := &commandBusHandler{logger: watermill.NopLogger{}, commandHandlers: commandHandlers}

	outgoingMessages, err := busHandler.HandleMessage(message.NewMessage(watermill.NewUUID(), message.Payload("dummy_payload")))
	require.NoError(t, err)
	assert.NotNil(t, outgoingMessages)
	assert.Equal(t, len(outgoingMessages), 1)
	assert.Equal(t, "test_command_handler_1", outgoingMessages[0].Metadata["handler_name"])
}
