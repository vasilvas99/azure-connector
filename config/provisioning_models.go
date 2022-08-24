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

package config

import "github.com/pkg/errors"

// AzureDpsRegisterDeviceRequest represents the registration ID for a device in the Azure DPS.
type AzureDpsRegisterDeviceRequest struct {
	RegistrationID string `json:"registrationId,omitempty"`
}

// AzureDpsRegistrationState represents the device registration state for a registered device in the Azure DPS.
type AzureDpsRegistrationState struct {
	X509                   interface{} `json:"x509,omitempty"`
	RegistrationID         string      `json:"registrationId,omitempty"`
	CreatedDateTimeUtc     string      `json:"createdDateTimeUtc,omitempty"`
	AssignedHub            string      `json:"assignedHub,omitempty"`
	DeviceID               string      `json:"deviceId,omitempty"`
	Status                 string      `json:"status,omitempty"`
	Substatus              string      `json:"substatus,omitempty"`
	LastUpdatedDateTimeUtc string      `json:"lastUpdatedDateTimeUtc,omitempty"`
	Etag                   string      `json:"etag,omitempty"`
}

// AzureDpsDeviceInfoResponse contains the device connection information, returned by the Azure DPS.
type AzureDpsDeviceInfoResponse struct {
	OperationID       string                    `json:"operationId,omitempty"`
	Status            string                    `json:"status,omitempty"`
	RegistrationState AzureDpsRegistrationState `json:"registrationState,omitempty"`
}

// AzureDeviceData contains the basic device connection information (assigned hub + device ID), returned by the Azure DPS.
type AzureDeviceData struct {
	// TODO: add support per https://docs.microsoft.com/en-us/azure/iot-dps/concepts-service
	AssignedHub string `json:"assignedHub,omitempty"`
	DeviceID    string `json:"deviceId,omitempty"`
}

func (d *AzureDeviceData) validate() error {
	if len(d.AssignedHub) == 0 {
		return errors.New("error occurred: missing assignedHub")
	}
	if len(d.DeviceID) == 0 {
		return errors.New("error occurred: missing deviceId")
	}
	return nil
}

// ResponseError represents the error response information from a provisioning service.
type ResponseError struct {
	Message string `json:"message,omitempty"`
	Detail  string `json:"detail,omitempty"`
}
