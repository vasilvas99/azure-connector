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

package handlers

import (
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/eclipse-kanto/azure-connector/config"
)

// messageHandler represents the internal interface for implementing a message handler.
type messageHandler interface {
	Init(connInfo *config.RemoteConnectionInfo) error
	HandleMessage(message *message.Message) ([]*message.Message, error)
	Name() string
}

// CommandHandler interface allows pluggability for handlers for cloud-to-device messages from Azure IoT Hub.
type CommandHandler interface {
	messageHandler
}

// TelemetryHandler interface allows pluggability for handlers for device-to-cloud messages to Azure IoT Hub.
type TelemetryHandler interface {
	messageHandler
	Topics() string
}
