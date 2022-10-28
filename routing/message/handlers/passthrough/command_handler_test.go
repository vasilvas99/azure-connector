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
	"strings"
	"testing"

	"github.com/ThreeDotsLabs/watermill/message"

	"github.com/eclipse-kanto/suite-connector/connector"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateDefaultCommandHandler(t *testing.T) {
	messageHandler := CreateDefaultCommandHandler()
	require.NoError(t, messageHandler.Init(nil))
	assert.Equal(t, commandHandlerName, messageHandler.Name())
}

func TestHandleOneWayCommand(t *testing.T) {
	messageHandler := CreateDefaultCommandHandler()
	require.NoError(t, messageHandler.Init(nil))

	payload := `{
		"topic": "org.eclipse.kanto/Test:testing/things/live/messages/toggle",
		"headers": {
			"version": 2,
			"response-required": false,
			"content-type": "application/json"
		},
		"path": "/features/kanto:testing:BinarySwitchExt:1/inbox/messages/toggle",
		"value": {}
	}`

	azureMessages, err := messageHandler.HandleMessage(&message.Message{Payload: []byte(payload)})
	require.NoError(t, err)
	assert.Equal(t, 2, len(azureMessages))

	azureMsg := azureMessages[0]
	azureMsgTopic, ok := connector.TopicFromCtx(azureMsg.Context())
	assert.True(t, ok)
	assert.True(t, strings.Contains(azureMsgTopic, "req//toggle"))

	azureMsg = azureMessages[1]
	azureMsgTopic, ok = connector.TopicFromCtx(azureMsg.Context())
	assert.True(t, ok)
	assert.True(t, strings.Contains(azureMsgTopic, "q//toggle"))
}

func TestHandleCommand(t *testing.T) {
	messageHandler := CreateDefaultCommandHandler()
	require.NoError(t, messageHandler.Init(nil))

	payload := `{
		"topic": "org.eclipse.kanto/Test:testing/things/live/messages/toggle",
		"headers": {
			"correlation-id": "f0a03c95-9526-4995-9718-2fb5cc866200",
			"version": 2,
			"response-required": true,
			"content-type": "application/json"
		},
		"path": "/features/kanto:testing:BinarySwitchExt:1/inbox/messages/toggle",
		"value": {}
	}`

	azureMessages, err := messageHandler.HandleMessage(&message.Message{Payload: []byte(payload)})
	require.NoError(t, err)
	assert.Equal(t, 2, len(azureMessages))

	azureMsg := azureMessages[0]
	azureMsgTopic, ok := connector.TopicFromCtx(azureMsg.Context())
	assert.True(t, ok)
	assert.True(t, strings.HasPrefix(azureMsgTopic, "command//"))
	assert.Equal(t, payload, string(azureMsg.Payload))

	azureMsg = azureMessages[1]
	azureMsgTopic, ok = connector.TopicFromCtx(azureMsg.Context())
	assert.True(t, ok)
	assert.True(t, strings.HasPrefix(azureMsgTopic, "c//"))
	assert.Equal(t, payload, string(azureMsg.Payload))
}
