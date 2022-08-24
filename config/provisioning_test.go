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

package config_test

import (
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"testing"

	"github.com/eclipse-kanto/azure-connector/config"
	mock "github.com/eclipse-kanto/azure-connector/config/internal/mock"
	"github.com/eclipse-kanto/suite-connector/logger"

	test "github.com/eclipse-kanto/azure-connector/config/internal/testing"
	gomock "github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	provisioningFileDefaultContent = `{"assignedHub":"test-iot.azure-devices.net","deviceId":"test-demo-device"}`
	azureRegisterDefaultJson       = `{
		"operationId": "5.b4ba454a90f38510.17d1dba6-67df-4b6c-bbdd-bf94a5507404",
		"status": "assigned"
	}`
	azureGetInfoDefaultJson = `{
		"operationId": "5.b4ba454a90f38510.17d1dba7-18df-4b3c-bbdd-bf94a5107404",
		"status": "assigned",
		"registrationState": {
			"x509": {},
			"registrationId": "test-demo-device",
			"createdDateTimeUtc": "2021-11-05T13:05:11.0655047Z",
			"assignedHub": "test-iot.azure-devices.net",
			"deviceId": "test-demo-device",
			"status": "assigned",
			"substatus": "reprovisionedToInitialAssignment",
			"lastUpdatedDateTimeUtc": "2021-11-05T13:05:11.3481733Z",
			"etag": "IjE4MDFhOGI5LTAwMDAtMGQwMC0wMDAwLTYxODUyYzA3MDAwMCI="
		}
	}`
	provisioningAssignedHub = "test-iot.azure-devices.net"
	provisioningDeviceId    = "test-demo-device"

	retryAfterHeaderKey = "Retry-After"
	testScopeId         = "1ne113B8627"
	responseError       = "response-error"
)

func TestDeviceDataFromRequestCorrectly(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	provisioningService := config.NewProvisioningService(nil)

	mockClient := mock.NewMockProvisioningHTTPClient(mockCtrl)
	var writer bytes.Buffer
	provisioningService.Init(mockClient, &writer)

	connSettings := &config.AzureConnectionSettings{}
	connSettings.DeviceKey = test.CertificateKey()
	_, err := provisioningService.GetDeviceData("", connSettings)
	require.Error(t, err)
}

func TestDeviceDataFromRequestSkippingPSSCorrectly(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	provisioningService := config.NewProvisioningService(nil)

	mockClient := mock.NewMockProvisioningHTTPClient(mockCtrl)
	var writer bytes.Buffer
	provisioningService.Init(mockClient, &writer)

	mockAzureDeviceRegisterRes := mockRequest(
		bodyFromStr(azureRegisterDefaultJson),
		http.StatusAccepted,
		map[string][]string{retryAfterHeaderKey: {"0"}},
	)
	mockAzureGetInfoRes := mockRequest(bodyFromStr(azureGetInfoDefaultJson), http.StatusOK, nil)

	mockClient.EXPECT().Post(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, nil).Times(0)
	mockClient.EXPECT().Do(gomock.Any()).Return(mockAzureDeviceRegisterRes, nil).Times(1)
	mockClient.EXPECT().Get(gomock.Any()).Return(mockAzureGetInfoRes, nil).Times(1)

	connSettings := &config.AzureConnectionSettings{}
	deviceData, err := provisioningService.GetDeviceData(testScopeId, connSettings)
	require.NoError(t, err)
	assert.Equal(t, deviceData.AssignedHub, provisioningAssignedHub)
	assert.Equal(t, deviceData.DeviceID, provisioningDeviceId)
	assert.Equal(t, provisioningFileDefaultContent, writer.String())
}

func TestDeviceDataFromRequestWithPersistToDiskError(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	logger := logger.NewLogger(log.New(io.Discard, "", log.Ldate), logger.INFO)
	provisioningService := config.NewProvisioningService(logger)

	writer := test.CreateErrorWriter()
	mockClient := mock.NewMockProvisioningHTTPClient(mockCtrl)
	provisioningService.Init(mockClient, writer)

	mockAzureDeviceRegisterRes := mockRequest(
		bodyFromStr(azureRegisterDefaultJson),
		http.StatusAccepted,
		map[string][]string{retryAfterHeaderKey: {"0"}},
	)
	mockAzureGetInfoRes := mockRequest(bodyFromStr(azureGetInfoDefaultJson), http.StatusOK, nil)

	mockClient.EXPECT().Post(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, nil).Times(0)
	mockClient.EXPECT().Do(gomock.Any()).Return(mockAzureDeviceRegisterRes, nil).Times(1)
	mockClient.EXPECT().Get(gomock.Any()).Return(mockAzureGetInfoRes, nil).Times(1)

	connSettings := &config.AzureConnectionSettings{}
	deviceData, err := provisioningService.GetDeviceData(testScopeId, connSettings)
	assert.NoError(t, err)
	assert.Equal(t, deviceData.AssignedHub, provisioningAssignedHub)
	assert.Equal(t, deviceData.DeviceID, provisioningDeviceId)
}

func TestDeviceDataFromAzureDPSRegisterRequestCorrectly(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	provisioningService := config.NewProvisioningService(nil)

	mockClient := mock.NewMockProvisioningHTTPClient(mockCtrl)
	var writer bytes.Buffer
	provisioningService.Init(mockClient, &writer)

	connSettings := &config.AzureConnectionSettings{}
	connSettings.DeviceKey = test.CertificateKey()
	_, err := provisioningService.GetDeviceData("", connSettings)
	require.Error(t, err)
}

func TestDeviceDataFromDiskCorrectly(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	provisioningService := config.NewProvisioningService(nil)
	mockClient := mock.NewMockProvisioningHTTPClient(mockCtrl)

	var writer bytes.Buffer
	writer.WriteString(provisioningFileDefaultContent)
	provisioningService.Init(nil, &writer)

	mockClient.EXPECT().Post(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, nil).Times(0)
	mockClient.EXPECT().Do(gomock.Any()).Return(nil, nil).Times(0)
	mockClient.EXPECT().Get(gomock.Any()).Return(nil, nil).Times(0)

	connSettings := &config.AzureConnectionSettings{}
	deviceData, err := provisioningService.GetDeviceData("", connSettings)
	require.NoError(t, err)
	assert.Equal(t, deviceData.AssignedHub, provisioningAssignedHub)
	assert.Equal(t, deviceData.DeviceID, provisioningDeviceId)
}

func TestDeviceDataFromRequestWithEmptyDeviceInfo(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	logger := logger.NewLogger(log.New(io.Discard, "", log.Ldate), logger.INFO)
	provisioningService := config.NewProvisioningService(logger)

	mockClient := mock.NewMockProvisioningHTTPClient(mockCtrl)
	var writer bytes.Buffer
	provisioningService.Init(mockClient, &writer)

	mockAzureDeviceRegisterRes := mockRequest(
		bodyFromStr(azureRegisterDefaultJson),
		http.StatusAccepted,
		map[string][]string{retryAfterHeaderKey: {"0"}},
	)
	mockAzureGetInfoRes := mockRequest(bodyFromStr(``), http.StatusOK, nil)

	mockClient.EXPECT().Post(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, nil).Times(0)
	mockClient.EXPECT().Do(gomock.Any()).Return(mockAzureDeviceRegisterRes, nil).Times(1)
	mockClient.EXPECT().Get(gomock.Any()).Return(mockAzureGetInfoRes, nil).Times(1)

	connSettings := &config.AzureConnectionSettings{}
	_, err := provisioningService.GetDeviceData(testScopeId, connSettings)
	assert.Error(t, err)
}

func TestDeviceDataFromRequestWithEmptyRegistrationState(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	logger := logger.NewLogger(log.New(io.Discard, "", log.Ldate), logger.INFO)
	provisioningService := config.NewProvisioningService(logger)

	mockClient := mock.NewMockProvisioningHTTPClient(mockCtrl)
	var writer bytes.Buffer
	provisioningService.Init(mockClient, &writer)

	mockAzureDeviceRegisterRes := mockRequest(
		bodyFromStr(azureRegisterDefaultJson),
		http.StatusAccepted,
		map[string][]string{retryAfterHeaderKey: {"0"}},
	)
	azureGetInfoJson := `{
		"operationId": "5.b4ba454a90f38510.17d1dba7-18df-4b3c-bbdd-bf94a5107404",
		"status": "assigned",
		"registrationState": {}
		}`
	mockAzureGetInfoRes := mockRequest(bodyFromStr(azureGetInfoJson), http.StatusOK, nil)

	mockClient.EXPECT().Post(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, nil).Times(0)
	mockClient.EXPECT().Do(gomock.Any()).Return(mockAzureDeviceRegisterRes, nil).Times(1)
	mockClient.EXPECT().Get(gomock.Any()).Return(mockAzureGetInfoRes, nil).Times(1)

	connSettings := &config.AzureConnectionSettings{}
	_, err := provisioningService.GetDeviceData(testScopeId, connSettings)
	assert.Error(t, err)
}

func TestDeviceDataFromRequestWithErrors(t *testing.T) {
	type getDeviceDataRequestErrorTC struct {
		azureGetInfoJson string
	}

	var testData = map[string]getDeviceDataRequestErrorTC{
		"TestDeviceDataFromRequestWithEmptyAssignedHub": {
			azureGetInfoJson: `{
				"operationId": "5.b4ba454a90f38510.17d1dba7-18df-4b3c-bbdd-bf94a5107404",
				"status": "assigned",
				"registrationState": {
					"x509": {},
					"registrationId": "test-demo-device",
					"createdDateTimeUtc": "2021-11-05T13:05:11.0655047Z",
					"assignedHub": "",
					"deviceId": "test-demo-device",
					"status": "assigned",
					"substatus": "reprovisionedToInitialAssignment",
					"lastUpdatedDateTimeUtc": "2021-11-05T13:05:11.3481733Z",
					"etag": "IjE4MDFhOGI5LTAwMDAtMGQwMC0wMDAwLTYxODUyYzA3MDAwMCI="
				}
			}`,
		},
		"TestDeviceDataFromRequestWithEmptyDeviceId": {
			azureGetInfoJson: `{
				"operationId": "5.b4ba454a90f38510.17d1dba7-18df-4b3c-bbdd-bf94a5107404",
				"status": "assigned",
				"registrationState": {
					"x509": {},
					"registrationId": "test-demo-device",
					"createdDateTimeUtc": "2021-11-05T13:05:11.0655047Z",
					"assignedHub": "test-iot.azure-devices.net",
					"deviceId": "",
					"status": "assigned",
					"substatus": "reprovisionedToInitialAssignment",
					"lastUpdatedDateTimeUtc": "2021-11-05T13:05:11.3481733Z",
					"etag": "IjE4MDFhOGI5LTAwMDAtMGQwMC0wMDAwLTYxODUyYzA3MDAwMCI="
				}
			}`,
		},
		"TestDeviceDataFromRequestWithoutAssignedHub": {
			azureGetInfoJson: `{
				"operationId": "5.b4ba454a90f38510.17d1dba7-18df-4b3c-bbdd-bf94a5107404",
				"status": "assigned",
				"registrationState": {
					"x509": {},
					"registrationId": "test-demo-device",
					"createdDateTimeUtc": "2021-11-05T13:05:11.0655047Z",
					"deviceId": "test-demo-device",
					"status": "assigned",
					"substatus": "reprovisionedToInitialAssignment",
					"lastUpdatedDateTimeUtc": "2021-11-05T13:05:11.3481733Z",
					"etag": "IjE4MDFhOGI5LTAwMDAtMGQwMC0wMDAwLTYxODUyYzA3MDAwMCI="
				}
			}`,
		},
		"TestDeviceDataFromRequestWithoutDeviceId": {
			azureGetInfoJson: `{
				"operationId": "5.b4ba454a90f38510.17d1dba7-18df-4b3c-bbdd-bf94a5107404",
				"status": "assigned",
				"registrationState": {
					"x509": {},
					"registrationId": "test-demo-device",
					"createdDateTimeUtc": "2021-11-05T13:05:11.0655047Z",
					"assignedHub": "test-iot.azure-devices.net",
					"status": "assigned",
					"substatus": "reprovisionedToInitialAssignment",
					"lastUpdatedDateTimeUtc": "2021-11-05T13:05:11.3481733Z",
					"etag": "IjE4MDFhOGI5LTAwMDAtMGQwMC0wMDAwLTYxODUyYzA3MDAwMCI="
				}
			}`,
		},
	}

	for testName, testValue := range testData {
		t.Run(testName, func(t *testing.T) {
			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()

			logger := logger.NewLogger(log.New(io.Discard, "", log.Ldate), logger.INFO)
			provisioningService := config.NewProvisioningService(logger)

			mockClient := mock.NewMockProvisioningHTTPClient(mockCtrl)
			var writer bytes.Buffer
			provisioningService.Init(mockClient, &writer)

			mockAzureDeviceRegisterRes := mockRequest(
				bodyFromStr(azureRegisterDefaultJson),
				http.StatusAccepted,
				map[string][]string{retryAfterHeaderKey: {"0"}},
			)
			mockAzureGetInfoRes := mockRequest(bodyFromStr(testValue.azureGetInfoJson), http.StatusOK, nil)

			mockClient.EXPECT().Post(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, nil).Times(0)
			mockClient.EXPECT().Do(gomock.Any()).Return(mockAzureDeviceRegisterRes, nil).Times(1)
			mockClient.EXPECT().Get(gomock.Any()).Return(mockAzureGetInfoRes, nil).Times(1)

			connSettings := &config.AzureConnectionSettings{}
			_, err := provisioningService.GetDeviceData(testScopeId, connSettings)
			assert.Error(t, err)
		})
	}
}

func TestDeviceDataFromRequestWithAzureDPSGetInfoErrorResponse(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	provisioningService := config.NewProvisioningService(nil)

	mockClient := mock.NewMockProvisioningHTTPClient(mockCtrl)
	var writer bytes.Buffer
	provisioningService.Init(mockClient, &writer)

	mockAzureDeviceRegisterRes := mockRequest(bodyFromStr(`{}`), http.StatusAccepted, map[string][]string{retryAfterHeaderKey: {"0"}})

	mockClient.EXPECT().Post(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, nil).Times(0)
	mockClient.EXPECT().Do(gomock.Any()).Return(mockAzureDeviceRegisterRes, nil).Times(1)
	mockClient.EXPECT().Get(gomock.Any()).Return(nil, errors.New(responseError)).Times(1)

	connSettings := &config.AzureConnectionSettings{}
	_, err := provisioningService.GetDeviceData(testScopeId, connSettings)
	assert.Error(t, err)
}

func TestDeviceDataFromRequestWithAzureDPSRegisterErrorReadResponseBody(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	provisioningService := config.NewProvisioningService(nil)

	mockClient := mock.NewMockProvisioningHTTPClient(mockCtrl)
	var writer bytes.Buffer
	provisioningService.Init(mockClient, &writer)

	mockAzureDeviceRegisterRes := mockRequest(test.CreateErrorReadWriterCloser(), http.StatusAccepted, nil)

	mockClient.EXPECT().Post(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, nil).Times(0)
	mockClient.EXPECT().Do(gomock.Any()).Return(mockAzureDeviceRegisterRes, nil).Times(1)
	mockClient.EXPECT().Get(gomock.Any()).Return(nil, nil).Times(0)

	connSettings := &config.AzureConnectionSettings{}
	_, err := provisioningService.GetDeviceData(testScopeId, connSettings)
	assert.Error(t, err)
}

func TestDeviceDataFromRequestWithAzureDPSGetInfoErrorReadResponseBody(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	provisioningService := config.NewProvisioningService(nil)

	mockClient := mock.NewMockProvisioningHTTPClient(mockCtrl)
	var writer bytes.Buffer
	provisioningService.Init(mockClient, &writer)

	mockAzureDeviceRegisterRes := mockRequest(bodyFromStr(azureRegisterDefaultJson), http.StatusAccepted, map[string][]string{retryAfterHeaderKey: {"0"}})
	mockAzureGetInfoRes := mockRequest(test.CreateErrorReadWriterCloser(), http.StatusOK, nil)

	mockClient.EXPECT().Post(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, nil).Times(0)
	mockClient.EXPECT().Do(gomock.Any()).Return(mockAzureDeviceRegisterRes, nil).Times(1)
	mockClient.EXPECT().Get(gomock.Any()).Return(mockAzureGetInfoRes, nil).Times(1)

	connSettings := &config.AzureConnectionSettings{}
	_, err := provisioningService.GetDeviceData(testScopeId, connSettings)
	assert.Error(t, err)
}

func TestDeviceDataFromRequestWithoutClient(t *testing.T) {
	provisioningService := config.NewProvisioningService(nil)
	var writer bytes.Buffer
	provisioningService.Init(nil, &writer)

	_, err := provisioningService.GetDeviceData("", nil)
	assert.Error(t, err)
}

func TestDeviceDataFromDiskWithErrors(t *testing.T) {
	type getDeviceDataDiskErrorTC struct {
		provisioningFileContent string
	}

	var testData = map[string]getDeviceDataDiskErrorTC{
		"TestDeviceDataFromDiskWithMissingAssignedHub": {
			provisioningFileContent: `{"deviceId":"test-demo-device"}`,
		},
		"TestDeviceDataFromDiskWithMissingDeviceId": {
			provisioningFileContent: `{"assignedHub":"test-iot.azure-devices.net"}`,
		},
		"TestDeviceDataFromDiskWithEmptyAssignedHub": {
			provisioningFileContent: `{"assignedHub":"","deviceId":"test-demo-device"}`,
		},
		"TestDeviceDataFromDiskWithEmptyDeviceId": {
			provisioningFileContent: `{"assignedHub":"test-iot.azure-devices.net","deviceId":""}`,
		},
		"TestDeviceDataFromDiskWithInvalidProvisioningData": {
			provisioningFileContent: `{"}`,
		},
		"TestDeviceDataFromDiskWithEmptyProvisioningData": {
			provisioningFileContent: `{}`,
		},
	}

	provisioningService := config.NewProvisioningService(nil)

	for testName, testValue := range testData {
		t.Run(testName, func(t *testing.T) {
			var writer bytes.Buffer
			writer.WriteString(testValue.provisioningFileContent)
			provisioningService.Init(nil, &writer)

			connSettings := &config.AzureConnectionSettings{}
			_, err := provisioningService.GetDeviceData("", connSettings)
			assert.Error(t, err)
		})
	}
}

func TestDeviceDataFromDiskWithCorruptedProvisioningData(t *testing.T) {
	provisioningService := config.NewProvisioningService(nil)
	errorFileWriter := test.CreateErrorReadWriterCloser()
	provisioningService.Init(nil, errorFileWriter)

	connSettings := &config.AzureConnectionSettings{}
	_, err := provisioningService.GetDeviceData("", connSettings)
	assert.Error(t, err)
}

func mockRequest(body io.ReadCloser, statusCode int, header http.Header) *http.Response {
	return &http.Response{
		StatusCode: statusCode,
		Body:       body,
		Header:     header,
	}
}

func bodyFromStr(body string) io.ReadCloser {
	return ioutil.NopCloser(bytes.NewReader([]byte(body)))
}
