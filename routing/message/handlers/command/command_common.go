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

package command

import (
	"context"

	"github.com/ThreeDotsLabs/watermill/message"
	routingmessage "github.com/eclipse-kanto/azure-connector/routing/message"
	handlers "github.com/eclipse-kanto/azure-connector/routing/message/handlers/common"
)

type contextKey int

const (
	commandMessageContextKey contextKey = 4 + iota //the rest of the context keys are defined in the connector package
)

var messageHandlers = []handlers.MessageHandler{}

// MessageHandlers returns the command message handlers.
func MessageHandlers() []handlers.MessageHandler {
	return messageHandlers
}

// SetMessageToContext sets a command message instance as a value to a context.
func SetMessageToContext(msg *message.Message, commandMessage *routingmessage.CloudMessage) context.Context {
	return context.WithValue(msg.Context(), commandMessageContextKey, commandMessage)
}

// MessageFromContext gets a command message instance from a context.
func MessageFromContext(msg *message.Message) *routingmessage.CloudMessage {
	value, ok := msg.Context().Value(commandMessageContextKey).(*routingmessage.CloudMessage)
	if ok {
		return value
	}
	return nil
}

// GetCommandMessageFromContext gets a command message instance from a context.
func GetCommandMessageFromContext(msg *message.Message) *routingmessage.CloudMessage {
	value, ok := msg.Context().Value(commandMessageContextKey).(*routingmessage.CloudMessage)
	if ok {
		return value
	}
	return nil
}

func registerMessageHandler(messageHandler handlers.MessageHandler) {
	messageHandlers = append(messageHandlers, messageHandler)
}
