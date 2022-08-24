// Copyright (c) 2022 Contributors to the Eclipse Foundation
//
// See the NOTICE file(s) distributed with this work for additional
// information regarding copyright ownership.
//
// This program and the accompanying materials are made available under the
// terms of the Eclipse Public License 2.0 which is available at
// http://www.eclipse.org/legal/epl-2.0
//
// SPDX-License-Identifier: EPL-2.0

package message

// CloudMessage represents the envelope for the cloud-to-device messages.
type CloudMessage struct {
	CommandName     string      `json:"cmdName,omitempty"`
	ApplicationID   string      `json:"appId,omitempty"`
	CorrelationID   string      `json:"cId,omitempty"`
	Timestamp       int64       `json:"ts,omitempty"`
	EnvelopeVersion string      `json:"eVer,omitempty"`
	Payload         interface{} `json:"p,omitempty"`
	PayloadVersion  string      `json:"pVer,omitempty"`
}

// TelemetryMessage represents the envelope for the telemetry messages.
type TelemetryMessage struct {
	MessageType     int         `json:"mt,omitempty"`
	MessageSubType  string      `json:"mst,omitempty"`
	ApplicationID   string      `json:"appId,omitempty"`
	CorrelationID   string      `json:"cId,omitempty"`
	Timestamp       int64       `json:"ts,omitempty"`
	EnvelopeVersion string      `json:"eVer,omitempty"`
	Payload         interface{} `json:"p,omitempty"`
	PayloadVersion  string      `json:"pVer,omitempty"`
}
