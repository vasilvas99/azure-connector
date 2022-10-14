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
	"testing"

	"github.com/eclipse-kanto/suite-connector/connector"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateCommandHandler(t *testing.T) {
	messageHandler := CreateCommandHandler("command-topic")
	require.NoError(t, messageHandler.Init(nil))
	assert.Equal(t, commandHandlerName, messageHandler.Name())
}

func TestHandleCommand(t *testing.T) {
	topic := "command-topic"
	messageHandler := CreateCommandHandler(topic)
	require.NoError(t, messageHandler.Init(nil))

	payload := "dummy_payload"
	azureMessages, err := messageHandler.HandleMessage(&message.Message{Payload: []byte(payload)})
	require.NoError(t, err)

	assert.Equal(t, 1, len(azureMessages))
	azureMsg := azureMessages[0]
	azureMsgTopic, _ := connector.TopicFromCtx(azureMsg.Context())

	assert.Equal(t, topic, azureMsgTopic)
	assert.Equal(t, payload, string(azureMsg.Payload))
}
