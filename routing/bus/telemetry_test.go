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
	"io"
	"log"
	"reflect"
	"testing"

	"github.com/pkg/errors"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"

	"github.com/eclipse-kanto/azure-connector/config"
	"github.com/eclipse-kanto/azure-connector/util"

	test "github.com/eclipse-kanto/azure-connector/routing/bus/internal/testing"
	"github.com/eclipse-kanto/azure-connector/routing/message/handlers"

	conn "github.com/eclipse-kanto/suite-connector/connector"
	"github.com/eclipse-kanto/suite-connector/logger"

	"github.com/stretchr/testify/assert"
)

const (
	fieldHandlers            = "handlers"
	testTelemetryHandlerName = "test_telemetry_handler"
)

func TestNoTelemetryMessageHandlers(t *testing.T) {
	settings := &config.AzureSettings{
		ConnectionString: "HostName=dummy-hub.azure-devices.net;DeviceId=dummy-device;SharedAccessKey=dGVzdGF6dXJlc2hhcmVkYWNjZXNza2V5",
	}
	logger := logger.NewLogger(log.New(io.Discard, "", log.Ldate), logger.INFO)
	connSettings, _ := config.PrepareAzureConnectionSettings(settings, nil, logger)
	router, _ := message.NewRouter(message.RouterConfig{}, watermill.NopLogger{})

	telemetryHandlers := []handlers.MessageHandler{}
	TelemetryBus(router, conn.NullPublisher(), test.NewDummySubscriber(), settings, connSettings, telemetryHandlers)
	test.AssertNoRouterHandlers(t, router)
}

func TestTelemetryMessageHandlerWithoutTopics(t *testing.T) {
	settings := &config.AzureSettings{
		ConnectionString: "HostName=dummy-hub.azure-devices.net;DeviceId=dummy-device;SharedAccessKey=dGVzdGF6dXJlc2hhcmVkYWNjZXNza2V5",
	}
	logger := logger.NewLogger(log.New(io.Discard, "", log.Ldate), logger.INFO)
	connSettings, _ := config.PrepareAzureConnectionSettings(settings, nil, logger)
	router, _ := message.NewRouter(message.RouterConfig{}, watermill.NopLogger{})

	telemetryHandler := test.NewDummyMessageHandler(testTelemetryHandlerName, []string{}, nil)
	telemetryHandlers := []handlers.MessageHandler{telemetryHandler}
	TelemetryBus(router, conn.NullPublisher(), test.NewDummySubscriber(), settings, connSettings, telemetryHandlers)
	test.AssertNoRouterHandlers(t, router)
}

func TestSingleTelemetryMessageHandler(t *testing.T) {
	settings := &config.AzureSettings{
		ConnectionString: "HostName=dummy-hub.azure-devices.net;DeviceId=dummy-device;SharedAccessKey=dGVzdGF6dXJlc2hhcmVkYWNjZXNza2V5",
	}
	logger := logger.NewLogger(log.New(io.Discard, "", log.Ldate), logger.INFO)
	connSettings, _ := config.PrepareAzureConnectionSettings(settings, nil, logger)
	router, _ := message.NewRouter(message.RouterConfig{}, watermill.NopLogger{})

	telemetryHandler := test.NewDummyMessageHandler(testTelemetryHandlerName, []string{"telemetry/#"}, nil)
	telemetryHandlers := []handlers.MessageHandler{telemetryHandler}
	TelemetryBus(router, conn.NullPublisher(), test.NewDummySubscriber(), settings, connSettings, telemetryHandlers)
	refRouterPtr := reflect.ValueOf(router)
	refRouter := reflect.Indirect(refRouterPtr)
	refHandlers := refRouter.FieldByName(fieldHandlers)
	assert.Equal(t, 1, refHandlers.Len())
	refHandler := refHandlers.MapIndex(refHandlers.MapKeys()[0])
	test.AssertRouterHandler(t, testTelemetryHandlerName, "telemetry/#", "", reflect.Indirect(refHandler))
}

func TestMultipleTelemetryMessageHandlers(t *testing.T) {
	settings := &config.AzureSettings{
		ConnectionString: "HostName=dummy-hub.azure-devices.net;DeviceId=dummy-device;SharedAccessKey=dGVzdGF6dXJlc2hhcmVkYWNjZXNza2V5",
	}
	logger := logger.NewLogger(log.New(io.Discard, "", log.Ldate), logger.INFO)
	connSettings, _ := config.PrepareAzureConnectionSettings(settings, nil, logger)
	router, _ := message.NewRouter(message.RouterConfig{}, watermill.NopLogger{})

	handlerNames := []string{"test_handler_1", "test_handler_2", "test_handler_3"}
	var telemetryHandlers []handlers.MessageHandler
	for _, handlerName := range handlerNames {
		telemetryHandlers = append(telemetryHandlers, test.NewDummyMessageHandler(handlerName, []string{"telemetry/#"}, nil))
	}
	TelemetryBus(router, conn.NullPublisher(), test.NewDummySubscriber(), settings, connSettings, telemetryHandlers)
	refRouterPtr := reflect.ValueOf(router)
	refRouter := reflect.Indirect(refRouterPtr)
	refHandlers := refRouter.FieldByName(fieldHandlers)
	assert.Equal(t, 3, refHandlers.Len())
	for i := 0; i < 3; i++ {
		refHandler := refHandlers.MapIndex(refHandlers.MapKeys()[i])
		handlerName := reflect.Indirect(refHandler).FieldByName("name").String()
		assert.True(t, util.ContainsString(handlerNames, handlerName))
		test.AssertRouterHandler(t, handlerName, "telemetry/#", "", reflect.Indirect(refHandler))
	}
}

func TestTelemetryMessageHandlerInitializationError(t *testing.T) {
	settings := &config.AzureSettings{
		ConnectionString: "HostName=dummy-hub.azure-devices.net;DeviceId=dummy-device;SharedAccessKey=dGVzdGF6dXJlc2hhcmVkYWNjZXNza2V5",
	}
	logger := logger.NewLogger(log.New(io.Discard, "", log.Ldate), logger.INFO)
	connSettings, _ := config.PrepareAzureConnectionSettings(settings, nil, logger)
	router, _ := message.NewRouter(message.RouterConfig{}, watermill.NopLogger{})

	telemetryHandler := test.NewDummyMessageHandler(testTelemetryHandlerName, []string{"telemetry/#"}, errors.New("init error"))
	telemetryHandlers := []handlers.MessageHandler{telemetryHandler}
	TelemetryBus(router, conn.NullPublisher(), test.NewDummySubscriber(), settings, connSettings, telemetryHandlers)
	test.AssertNoRouterHandlers(t, router)
}
