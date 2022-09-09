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

package telemetry

import (
	handlers "github.com/eclipse-kanto/azure-connector/routing/message/handlers/common"
)

var messageHandlers = []handlers.MessageHandler{}

// MessageHandlers returns the telemetry message handlers.
func MessageHandlers() []handlers.MessageHandler {
	return messageHandlers
}

func registerMessageHandler(messageHandler handlers.MessageHandler) {
	messageHandlers = append(messageHandlers, messageHandler)
}
