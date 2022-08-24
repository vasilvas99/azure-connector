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
	"encoding/base64"
	"io"
	"log"
	"strings"
	"testing"
	"time"

	"github.com/pkg/errors"

	conn "github.com/eclipse-kanto/suite-connector/config"
	"github.com/eclipse-kanto/suite-connector/logger"

	"github.com/eclipse-kanto/azure-connector/config"
	mock "github.com/eclipse-kanto/azure-connector/config/internal/mock"
	test "github.com/eclipse-kanto/azure-connector/config/internal/testing"

	gomock "github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseMalformedConnectionString(t *testing.T) {
	connectionString := "HostName=dummy-hub.azure-devices.net;DeviceId;cGFzc3dvcmQ="
	settings := &config.AzureSettings{ConnectionString: connectionString}
	logger := logger.NewLogger(log.New(io.Discard, "", log.Ldate), logger.INFO)
	_, err := config.CreateAzureConnectionSettings(settings, logger)
	require.Error(t, err)
}

func TestCreateTokenConnectionSettings(t *testing.T) {
	connectionString := "HostName=dummy-hub.azure-devices.net;DeviceId=dummy-device;SharedAccessKey=cGFzc3dvcmQ="
	settings := &config.AzureSettings{ConnectionString: connectionString}
	logger := logger.NewLogger(log.New(io.Discard, "", log.Ldate), logger.INFO)
	connSettings, err := config.CreateAzureConnectionSettings(settings, logger)

	require.NoError(t, err)
	decodedSharedAccessKey, _ := base64.StdEncoding.DecodeString("cGFzc3dvcmQ=")
	assert.Equal(t, "dummy-device", connSettings.DeviceID)
	assert.Equal(t, "dummy-hub.azure-devices.net", connSettings.HostName)
	assert.Equal(t, "dummy-hub", connSettings.HubName)
	assert.Equal(t, "", connSettings.SharedAccessKey.SharedAccessKeyName)
	assert.Equal(t, decodedSharedAccessKey, connSettings.SharedAccessKey.SharedAccessKeyDecoded)
	assert.Equal(t, time.Hour, connSettings.TokenValidity)
	assert.Equal(t, "", connSettings.DeviceCert)
	assert.Equal(t, "", connSettings.DeviceKey)
}

func TestMalformedSharedAccessKey(t *testing.T) {
	connectionString := "HostName=dummy-hub.azure-devices.net;DeviceId=dummy-device;SharedAccessKey=x7HrdC+URzEneFam9ZKa0Ke7="
	settings := &config.AzureSettings{ConnectionString: connectionString}
	logger := logger.NewLogger(log.New(io.Discard, "", log.Ldate), logger.INFO)
	_, err := config.CreateAzureConnectionSettings(settings, logger)
	require.Error(t, err)
}

func TestSASTokenValidityPeriod(t *testing.T) {
	connectionString := "HostName=dummy-hub.azure-devices.net;DeviceId=dummy-device;SharedAccessKey=cGFzc3dvcmQ="
	settings := &config.AzureSettings{ConnectionString: connectionString, SASTokenValidity: "2h"}
	logger := logger.NewLogger(log.New(io.Discard, "", log.Ldate), logger.INFO)
	connSettings, err := config.CreateAzureConnectionSettings(settings, logger)

	require.NoError(t, err)
	assert.Equal(t, 2*time.Hour, connSettings.TokenValidity)
}

func TestFallbackDefaultSASTokenValidityPeriod(t *testing.T) {
	connectionString := "HostName=dummy-hub.azure-devices.net;DeviceId=dummy-device;SharedAccessKey=cGFzc3dvcmQ="
	settings := &config.AzureSettings{ConnectionString: connectionString, SASTokenValidity: "invalid-sas-token-validity"}
	logger := logger.NewLogger(log.New(io.Discard, "", log.Ldate), logger.INFO)
	connSettings, err := config.CreateAzureConnectionSettings(settings, logger)

	require.NoError(t, err)
	assert.Equal(t, time.Hour, connSettings.TokenValidity)
}

func TestTokenConnectionSettingsMissingRequiredProperties(t *testing.T) {
	var testData = []struct {
		connString      string
		missingProperty string
	}{
		{
			"HostName=dummy-hub.azure-devices.net;SharedAccessKey=cGFzc3dvcmQ=", "DeviceId",
		},
		{
			"DeviceId=dummy-device;SharedAccessKey=cGFzc3dvcmQ=", "HostName",
		},
	}
	for _, testValues := range testData {
		t.Run(testValues.missingProperty, func(t *testing.T) {
			settings := &config.AzureSettings{ConnectionString: testValues.connString}
			logger := logger.NewLogger(log.New(io.Discard, "", log.Ldate), logger.INFO)
			_, err := config.CreateAzureConnectionSettings(settings, logger)
			require.Error(t, err)
		})
	}
}

func TestTokenConnectionSettingsInvalidHostName(t *testing.T) {
	var testData = []struct {
		connString string
		name       string
	}{
		{
			"HostName=.azure-devices.net;DeviceId=dummy-device;SharedAccessKey=cGFzc3dvcmQ=",
			"EmptyHostName",
		},
		{
			"HostName=malformed-host-name;DeviceId=dummy-device;SharedAccessKey=cGFzc3dvcmQ=",
			"MalformedHostName",
		},
		{
			"HostName=dummy-hub.azure-devices;DeviceId=dummy-device;SharedAccessKey=cGFzc3dvcmQ=",
			"MalformedHostNameDomain",
		},
	}
	for _, testValues := range testData {
		t.Run(testValues.name, func(t *testing.T) {
			settings := &config.AzureSettings{ConnectionString: testValues.connString}
			logger := logger.NewLogger(log.New(io.Discard, "", log.Ldate), logger.INFO)
			_, err := config.CreateAzureConnectionSettings(settings, logger)
			require.Error(t, err)
		})
	}
}

func TestCertificateConnectionSettingsMissingRequiredProperties(t *testing.T) {
	var testData = []struct {
		connString      string
		missingProperty string
	}{
		{
			"HostName=dummy-hub.azure-devices.net", "DeviceId",
		},
		{
			"DeviceId=dummy-device", "HostName",
		},
	}
	for _, testValues := range testData {
		t.Run(testValues.missingProperty, func(t *testing.T) {
			settings := &config.AzureSettings{
				ConnectionString: testValues.connString,
				TLSSettings: conn.TLSSettings{
					Cert: "device-certificate.crt",
					Key:  "device-certificate.key",
				},
			}
			logger := logger.NewLogger(log.New(io.Discard, "", log.Ldate), logger.INFO)
			_, err := config.CreateAzureConnectionSettings(settings, logger)
			require.Error(t, err)
		})
	}
}

func TestCreateCertificateConnectionSettings(t *testing.T) {
	connStringProperties := map[string]string{"HostName": "dummy-hub.azure-devices.net", "DeviceId": "dummy-device"}
	certFileReader := test.CreateDeviceCertificateReader()
	keyFileReader := test.CreateCertificateKeyReader()
	connSettings, err := config.CreateAzureCertificateConnectionSettings(connStringProperties, certFileReader, keyFileReader)

	require.NoError(t, err)
	assert.Equal(t, "dummy-device", connSettings.DeviceID)
	assert.Equal(t, "dummy-hub.azure-devices.net", connSettings.HostName)
	assert.Equal(t, "dummy-hub", connSettings.HubName)
	assert.Nil(t, connSettings.SharedAccessKey)
	assert.Equal(t, 0*time.Second, connSettings.TokenValidity)
	assert.Equal(t, test.DeviceCertificate(), connSettings.DeviceCert)
	assert.Equal(t, test.CertificateKey(), connSettings.DeviceKey)
}

func TestCertificateConnectionSettingsInvalidHostName(t *testing.T) {
	var testData = []struct {
		hostName string
		name     string
	}{
		{
			".azure-devices.net", "EmptyHostName",
		},
		{
			"malformed-host-name", "MalformedHostName",
		},
		{
			"dummy-hub.azure-devices", "MalformedHostNameDomain",
		},
	}
	for _, testValues := range testData {
		t.Run(testValues.name, func(t *testing.T) {
			connStringProperties := map[string]string{"HostName": testValues.hostName, "DeviceId": "dummy-device"}
			certFileReader := test.CreateDeviceCertificateReader()
			keyFileReader := test.CreateCertificateKeyReader()
			_, err := config.CreateAzureCertificateConnectionSettings(connStringProperties, certFileReader, keyFileReader)
			require.Error(t, err)
		})
	}
}

func TestCertificateConnectionSettingsMissingCertParams(t *testing.T) {
	var testData = []struct {
		certPath    string
		certKeyPath string
		name        string
	}{
		{
			"device-certificate.crt", "", "MissingCertificateKey",
		},
		{
			"", "device-certificate.key", "MissingCertificate",
		},
		{
			"", "", "MissingCertificateKeyPair",
		},
	}
	connectionString := "HostName=dummy-device.azure-devices.net;DeviceId=dummy-device"
	for _, testValues := range testData {
		t.Run(testValues.name, func(t *testing.T) {
			settings := &config.AzureSettings{
				ConnectionString: connectionString,
				TLSSettings: conn.TLSSettings{
					Cert: testValues.certPath,
					Key:  testValues.certKeyPath,
				},
			}
			logger := logger.NewLogger(log.New(io.Discard, "", log.Ldate), logger.INFO)
			_, err := config.CreateAzureConnectionSettings(settings, logger)
			require.Error(t, err)
		})
	}
}

func TestCertificateConnectionPropertiesInvalidCertAndKeyPaths(t *testing.T) {
	var testData = []struct {
		certPath    string
		certKeyPath string
		name        string
	}{
		{
			"invalid-certificate-path.crt", "device-certificate.key", "CertificatePath",
		},
		{
			"device-certificate.crt", "invalid-certificate-key-path.crt", "CertificateKeyPath",
		},
	}
	connectionString := "HostName=dummy-device.azure-devices.net;DeviceId=dummy-device"
	for _, testValues := range testData {
		t.Run(testValues.name, func(t *testing.T) {
			settings := &config.AzureSettings{
				ConnectionString: connectionString,
				TLSSettings: conn.TLSSettings{
					Cert: testValues.certPath,
					Key:  testValues.certKeyPath,
				},
			}
			logger := logger.NewLogger(log.New(io.Discard, "", log.Ldate), logger.INFO)
			_, err := config.CreateAzureConnectionSettings(settings, logger)
			require.Error(t, err)
		})
	}
}

func TestCertificateConnectionSettingsNoCertificate(t *testing.T) {
	connStringProperties := map[string]string{"HostName": "dummy-device.azure-devices.net", "DeviceId": "dummy-device"}
	certFileReader := strings.NewReader("")
	keyFileReader := test.CreateCertificateKeyReader()
	_, err := config.CreateAzureCertificateConnectionSettings(connStringProperties, certFileReader, keyFileReader)
	require.Error(t, err)
}

func TestCertificateConnectionSettingsMalformedCertificate(t *testing.T) {
	connStringProperties := map[string]string{"HostName": "dummy-device.azure-devices.net", "DeviceId": "dummy-device"}
	certFileReader := test.CreateMalformedDeviceCertificateReader()
	keyFileReader := test.CreateCertificateKeyReader()
	_, err := config.CreateAzureCertificateConnectionSettings(connStringProperties, certFileReader, keyFileReader)
	require.Error(t, err)
}

func TestCertificateConnectionSettingsNoCertificateKey(t *testing.T) {
	connStringProperties := map[string]string{"HostName": "dummy-device.azure-devices.net", "DeviceId": "dummy-device"}
	certFileReader := test.CreateDeviceCertificateReader()
	keyFileReader := strings.NewReader("")
	connSettings, err := config.CreateAzureCertificateConnectionSettings(connStringProperties, certFileReader, keyFileReader)

	require.NoError(t, err)
	assert.Equal(t, test.DeviceCertificate(), connSettings.DeviceCert)
	assert.Equal(t, "", connSettings.DeviceKey)
}

func TestCertificateConnectionSettingsMalformedCertificateKey(t *testing.T) {
	connStringProperties := map[string]string{"HostName": "dummy-device.azure-devices.net", "DeviceId": "dummy-device"}
	certFileReader := test.CreateDeviceCertificateReader()
	keyFileReader := test.CreateMalformedCertificateKeyReader()
	connSettings, err := config.CreateAzureCertificateConnectionSettings(connStringProperties, certFileReader, keyFileReader)

	require.NoError(t, err)
	assert.Equal(t, test.DeviceCertificate(), connSettings.DeviceCert)
	assert.Equal(t, test.MalformedCertificateKey(), connSettings.DeviceKey)
}

func TestCreateProvisioningConnectionSettings(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()

	provisioningService := mockProvisioningService(t, controller, createGetDeviceData(), nil, false, 1, 1)
	certFileReader := test.CreateDeviceCertificateReader()
	keyFileReader := test.CreateCertificateKeyReader()
	connSettings, err := config.CreateAzureProvisioningConnectionSettings(&config.AzureSettings{}, provisioningService, nil, true, certFileReader, keyFileReader)

	require.NoError(t, err)
	assert.Equal(t, "dummy-device", connSettings.DeviceID)
	assert.Equal(t, "dummy-hub.azure-devices.net", connSettings.HostName)
	assert.Equal(t, "dummy-hub", connSettings.HubName)
	assert.Nil(t, connSettings.SharedAccessKey)
	assert.Equal(t, 0*time.Second, connSettings.TokenValidity)
	assert.Equal(t, test.DeviceCertificate(), connSettings.DeviceCert)
	assert.Equal(t, test.CertificateKey(), connSettings.DeviceKey)
}

func TestCreateProvisioningConnectionSettingsExistingProvisioningFile(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()

	provisioningService := mockProvisioningService(t, controller, createGetDeviceData(), nil, true, 1, 1)
	certFileReader := test.CreateDeviceCertificateReader()
	keyFileReader := test.CreateCertificateKeyReader()
	connSettings, err := config.CreateAzureProvisioningConnectionSettings(&config.AzureSettings{}, provisioningService, nil, false, certFileReader, keyFileReader)

	require.NoError(t, err)
	assert.Equal(t, "dummy-device", connSettings.DeviceID)
	assert.Equal(t, "dummy-hub.azure-devices.net", connSettings.HostName)
	assert.Equal(t, "dummy-hub", connSettings.HubName)
	assert.Nil(t, connSettings.SharedAccessKey)
	assert.Equal(t, 0*time.Second, connSettings.TokenValidity)
	assert.Equal(t, test.DeviceCertificate(), connSettings.DeviceCert)
	assert.Equal(t, test.CertificateKey(), connSettings.DeviceKey)
	assert.Equal(t, 0*time.Second, connSettings.TokenValidity)
}

func TestProvisioningConnectionSettingsErrorGetDeviceData(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()

	provisioningService := mockProvisioningService(t, controller, nil, errors.New("cannot access DPS."), false, 1, 1)
	certFileReader := test.CreateDeviceCertificateReader()
	keyFileReader := test.CreateCertificateKeyReader()
	_, err := config.CreateAzureProvisioningConnectionSettings(&config.AzureSettings{}, provisioningService, nil, true, certFileReader, keyFileReader)

	require.Error(t, err)
}

func TestProvisioningConnectionSettingsInvalidCertAndKey(t *testing.T) {
	var testData = []struct {
		certReader    io.Reader
		certKeyReader io.Reader
		name          string
	}{
		{
			test.CreateDeviceCertificateReader(), test.CreateMismatchCertificateKeyReader(), "MismatchCertificateKeyPair",
		},
		{
			strings.NewReader(""), test.CreateCertificateKeyReader(), "NoDeviceCertificate",
		},
		{
			test.CreateErrorCertificateReader(), test.CreateCertificateKeyReader(), "ErrorReadCertificate",
		},
		{
			test.CreateDeviceCertificateReader(), strings.NewReader(""), "NoCertificateKey",
		},
		{
			test.CreateDeviceCertificateReader(), test.CreateMalformedCertificateKeyReader(), "MalformedCertificateKey",
		},
		{
			test.CreateDeviceCertificateReader(), test.CreateErrorCertificateKeyReader(), "ErrorReadCertificateKey",
		},
	}
	for _, testValues := range testData {
		t.Run(testValues.name, func(t *testing.T) {
			controller := gomock.NewController(t)
			provisioningService := mockProvisioningService(t, controller, createGetDeviceData(), nil, false, 0, 0)
			_, err := config.CreateAzureProvisioningConnectionSettings(&config.AzureSettings{}, provisioningService, nil, true, testValues.certReader, testValues.certKeyReader)
			controller.Finish()
			require.Error(t, err)
		})
	}
}

func TestProvisioningConnectionSettingsNoCertificates(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()

	provisioningService := mockProvisioningService(t, controller, createGetDeviceData(), nil, false, 0, 0)
	certFileReader := strings.NewReader("")
	keyFileReader := test.CreateCertificateKeyReader()
	_, err := config.CreateAzureProvisioningConnectionSettings(&config.AzureSettings{IDScope: "dummyIdScope"}, provisioningService, nil, true, certFileReader, keyFileReader)

	require.Error(t, err)
}

func TestProvisioningConnectionSettingsMalformedHostName(t *testing.T) {
	var testData = []struct {
		hostName string
		name     string
	}{
		{
			".azure-devices.net", "EmptyHostName",
		},
		{
			"malformed-host-name", "MalformedHostName",
		},
		{
			"dummy-hub.azure-devices", "MalformedHostNameDomain",
		},
	}
	for _, testValues := range testData {
		t.Run(testValues.name, func(t *testing.T) {
			controller := gomock.NewController(t)
			defer controller.Finish()
			deviceData := &config.AzureDeviceData{
				AssignedHub: testValues.hostName,
				DeviceID:    "dummy-device",
			}
			provisioningService := mockProvisioningService(t, controller, deviceData, nil, false, 1, 1)
			certFileReader := test.CreateDeviceCertificateReader()
			keyFileReader := test.CreateCertificateKeyReader()
			_, err := config.CreateAzureProvisioningConnectionSettings(&config.AzureSettings{}, provisioningService, nil, true, certFileReader, keyFileReader)
			require.Error(t, err)
		})
	}
}

func mockProvisioningService(t *testing.T, controller *gomock.Controller, deviceData *config.AzureDeviceData, deviceDataError error, hasProvisioningFile bool, timesGetDataCalled, timesInitCalled int) *mock.MockProvisioningService {
	provisioningService := mock.NewMockProvisioningService(controller)
	provisioningService.EXPECT().GetDeviceData(gomock.Any(), gomock.Any()).Return(deviceData, deviceDataError).Times(timesGetDataCalled)
	provisioningService.EXPECT().Init(gomock.Any(), gomock.Any()).Do(func(client config.ProvisioningHTTPClient, provisioningFile io.ReadWriter) {
		if hasProvisioningFile != (client == nil) {
			t.Fail()
		}
	}).Times(timesInitCalled)
	return provisioningService
}

func createGetDeviceData() *config.AzureDeviceData {
	return &config.AzureDeviceData{
		AssignedHub: "dummy-hub.azure-devices.net",
		DeviceID:    "dummy-device",
	}
}
