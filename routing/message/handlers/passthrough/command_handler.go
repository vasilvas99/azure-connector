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

package passthrough

import (
	"encoding/json"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/pkg/errors"

	"github.com/eclipse-kanto/suite-connector/connector"

	"github.com/eclipse-kanto/azure-connector/config"
	"github.com/eclipse-kanto/azure-connector/routing"
	"github.com/eclipse-kanto/azure-connector/routing/message/handlers"

	"github.com/eclipse/ditto-clients-golang/protocol"
)

const (
	commandHandlerName = "passthrough_command_handler"

	msgInvalidCloudCommand = "invalid cloud command"
)

type commandHandler struct{}

// CreateDefaultCommandHandler instantiates a new command handler that forwards cloud-to-device messages to the local message broker as Hono commands.
func CreateDefaultCommandHandler() handlers.CommandHandler {
	return new(commandHandler)
}

// Init does nothing.
func (h *commandHandler) Init(connInfo *config.RemoteConnectionInfo) error {
	return nil
}

// HandleMessage creates 2 new messages with the same payload, using Hono command topic and its shorthand version
func (h *commandHandler) HandleMessage(msg *message.Message) ([]*message.Message, error) {
	command := protocol.Envelope{Headers: protocol.NewHeaders()}

	if err := json.Unmarshal(msg.Payload, &command); err != nil {
		return nil, errors.Wrap(err, msgInvalidCloudCommand)
	}

	l := message.NewMessage(watermill.NewUUID(), msg.Payload)
	l.SetContext(connector.SetTopicToCtx(l.Context(), routing.CreateLocalCmdTopicLong(&command)))

	s := message.NewMessage(watermill.NewUUID(), msg.Payload)
	s.SetContext(connector.SetTopicToCtx(s.Context(), routing.CreateLocalCmdTopicShort(&command)))

	return []*message.Message{l, s}, nil
}

// Name returns the message handler name.
func (h *commandHandler) Name() string {
	return commandHandlerName
}
