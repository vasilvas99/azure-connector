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
	"fmt"
	"strings"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/eclipse-kanto/azure-connector/config"
	"github.com/eclipse-kanto/azure-connector/util"
	"github.com/eclipse-kanto/suite-connector/connector"
	"github.com/pkg/errors"
)

const commandPassthroughHandlerName = "command_passthrough_handler"

type commandPassthroughMessageHandler struct {
	topics   []string
	settings *config.AzureSettings
}

func (h *commandPassthroughMessageHandler) Init(settings *config.AzureSettings, connSettings *config.AzureConnectionSettings) error {
	h.settings = settings
	h.topics = strings.Split(settings.AllowedCloudMessageTypesList, ",")
	return nil
}

// TODO: keep it generic!
func (h *commandPassthroughMessageHandler) HandleMessage(msg *message.Message) ([]*message.Message, error) {
	cloudMessage := MessageFromContext(msg)
	if cloudMessage == nil {
		return nil, errors.New("cannot deserialize cloud message")
	}
	if util.ContainsString(h.topics, cloudMessage.CommandName) {
		msg.SetContext(connector.SetTopicToCtx(msg.Context(), cloudMessage.ApplicationID+"/"+cloudMessage.CommandName))
		return []*message.Message{msg}, nil
	}
	return nil, fmt.Errorf("cloud command name '%s' is not supported", cloudMessage.CommandName)
}

func (h *commandPassthroughMessageHandler) Name() string {
	return commandPassthroughHandlerName
}

func (h *commandPassthroughMessageHandler) Topics() []string {
	return h.topics
}

func init() {
	registerMessageHandler(&commandPassthroughMessageHandler{})
}
