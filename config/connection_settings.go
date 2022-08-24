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

import (
	"bufio"
	"encoding/base64"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/pkg/errors"

	"github.com/eclipse-kanto/suite-connector/logger"
	conn "github.com/eclipse-kanto/suite-connector/util"

	"github.com/eclipse-kanto/azure-connector/util"
)

const (
	provisioningJSONConfig         = "provisioning.json" // TODO: why with the same name as default flag value?
	hostNameSuffix                 = ".azure-devices.net"
	propertyKeyHostName            = "HostName"
	propertyKeyDeviceID            = "DeviceId"
	propertyKeySharedAccessKey     = "SharedAccessKey"
	propertyKeySharedAccessKeyName = "SharedAccessKeyName"
)

// SharedAccessKey contains the shared access key for generating SAS token for device authentication.
type SharedAccessKey struct {
	SharedAccessKeyName    string
	SharedAccessKeyDecoded []byte
}

// AzureConnectionSettings contains the configuration data for establishing connection to the Azure IoT Hub.
type AzureConnectionSettings struct {
	HubName       string
	HostName      string
	DeviceID      string
	DeviceCert    string
	DeviceKey     string
	TokenValidity time.Duration

	*SharedAccessKey
}

// CreateAzureConnectionSettings creates the configuration data for establishing connection to the Azure IoT Hub.
func CreateAzureConnectionSettings(settings *AzureSettings, log logger.Logger) (*AzureConnectionSettings, error) {
	connProps, err := parseConnectionString(settings.ConnectionString)
	if err != nil {
		return nil, err
	}
	if len(connProps[propertyKeySharedAccessKey]) > 0 {
		return CreateAzureSASTokenConnectionSettings(connProps, settings, log)
	}

	if !util.DeviceCertificatesArePresent(settings.Cert, settings.Key) {
		return nil, util.GenerateCertKeyError("connectionString", settings.Cert, settings.Key)
	}

	hasDeviceID := len(connProps[propertyKeyDeviceID]) > 0
	hasHostName := len(connProps[propertyKeyHostName]) > 0

	if hasDeviceID && !hasHostName {
		return nil, errors.New("the HostName is required")
	}

	if !hasDeviceID && hasHostName {
		return nil, errors.New("the DeviceId is required")
	}

	certFileReader, keyFileReader, err := createDeviceCertReaders(settings)
	if err != nil {
		return nil, err
	}
	if hasDeviceID && hasHostName {
		return CreateAzureCertificateConnectionSettings(connProps, certFileReader, keyFileReader)
	}

	useProvisioningClient := !conn.FileExists(provisioningJSONConfig)
	provisioningFile, err := os.OpenFile(provisioningJSONConfig, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return nil, err
	}

	connSettings, err := CreateAzureProvisioningConnectionSettings(
		settings,
		NewProvisioningService(log),
		provisioningFile,
		useProvisioningClient,
		certFileReader,
		keyFileReader)

	if err != nil {
		util.DeleteFileIfEmpty(provisioningFile)
		return nil, err
	}
	return connSettings, nil
}

// CreateAzureCertificateConnectionSettings creates the configuration data for establishing connection to the Azure IoT Hub via X.509 certificate.
func CreateAzureCertificateConnectionSettings(
	connStringProperties map[string]string,
	certFileReader io.Reader,
	keyFileReader io.Reader,
) (*AzureConnectionSettings, error) {
	var err error
	connSettings := &AzureConnectionSettings{}
	if err = attachCertificateInfo(connSettings, certFileReader, keyFileReader); err != nil {
		return nil, err
	}

	connSettings.HostName = connStringProperties[propertyKeyHostName]
	connSettings.DeviceID = connStringProperties[propertyKeyDeviceID]

	connSettings.HubName, err = extractAzureHubName(connStringProperties[propertyKeyHostName])
	if err != nil {
		return nil, err
	}

	return connSettings, nil
}

// CreateAzureProvisioningConnectionSettings creates the configuration data for establishing connection to Azure device that requires device provisioning.
func CreateAzureProvisioningConnectionSettings(
	settings *AzureSettings,
	provisioningService ProvisioningService,
	provisioningFile io.ReadWriter,
	useProvisioningClient bool,
	certFileReader io.Reader,
	keyFileReader io.Reader,
) (*AzureConnectionSettings, error) {
	connSettings := &AzureConnectionSettings{}
	err := attachCertificateInfo(connSettings, certFileReader, keyFileReader)
	if err != nil {
		return nil, err
	}

	var client ProvisioningHTTPClient
	if useProvisioningClient {
		client, err = initDeviceProvisioningClient(connSettings)
		if err != nil {
			return nil, err
		}
	}
	provisioningService.Init(client, provisioningFile)

	azureDeviceData, err := provisioningService.GetDeviceData(settings.IDScope, connSettings)
	if err != nil {
		return nil, err
	}

	connSettings.HubName, err = extractAzureHubName(azureDeviceData.AssignedHub)
	if err != nil {
		return nil, err
	}

	//  TODO:
	// op1: require AssignedHub and DeviceID if there is no idScope support
	// op2: add config per idScopeRequestURL and verificationCodeRequestURL
	connSettings.HostName = azureDeviceData.AssignedHub
	connSettings.DeviceID = azureDeviceData.DeviceID
	return connSettings, nil
}

// CreateAzureSASTokenConnectionSettings creates the configuration data for establishing connection to the Azure IoT Hub via SAS token.
func CreateAzureSASTokenConnectionSettings(
	connStringProperties map[string]string,
	settings *AzureSettings,
	logger logger.Logger,
) (*AzureConnectionSettings, error) {
	var err error
	connSettings := &AzureConnectionSettings{}
	if value, ok := connStringProperties[propertyKeyHostName]; ok {
		connSettings.HostName = value
	} else {
		return nil, errors.New("the HostName is required")
	}

	if value, ok := connStringProperties[propertyKeyDeviceID]; ok {
		connSettings.DeviceID = value
	} else {
		return nil, errors.New("the DeviceId is required")
	}

	connSettings.HubName, err = extractAzureHubName(connStringProperties[propertyKeyHostName])
	if err != nil {
		return nil, err
	}

	var sharedAccessKeyDecoded []byte
	sharedAccessKey := connStringProperties[propertyKeySharedAccessKey]
	if sharedAccessKeyDecoded, err = base64.StdEncoding.DecodeString(sharedAccessKey); err != nil {
		return nil, errors.New("the SharedAccessKey is not base64 encoded")
	}
	connSettings.SharedAccessKey = &SharedAccessKey{
		SharedAccessKeyName:    connStringProperties[propertyKeySharedAccessKeyName],
		SharedAccessKeyDecoded: sharedAccessKeyDecoded,
	}

	if tokenValidity, err := ParseSASTokenValidity(settings.SASTokenValidity); err != nil {
		logger.Warn("The default SAS token validity period will be set.", err, nil)
		connSettings.TokenValidity = defaultSASTokenValidity
	} else {
		connSettings.TokenValidity = tokenValidity
	}

	return connSettings, nil
}

func createDeviceCertReaders(settings *AzureSettings) (*bufio.Reader, *bufio.Reader, error) {
	certFile, err := os.Open(settings.Cert)
	if err != nil {
		return nil, nil, err
	}
	certFileReader := bufio.NewReader(certFile)
	certKeyFile, err := os.Open(settings.Key)
	if err != nil {
		return nil, nil, err
	}
	certKeyFileReader := bufio.NewReader(certKeyFile)
	return certFileReader, certKeyFileReader, nil
}

func attachCertificateInfo(connSettings *AzureConnectionSettings, certFileReader io.Reader, keyFileReader io.Reader) error {
	deviceCertRaw, err := ioutil.ReadAll(certFileReader)
	if err != nil {
		return errors.Wrap(err, "error occurred while reading public certificate file")
	}
	deviceKeyRaw, err := ioutil.ReadAll(keyFileReader)
	if err != nil {
		return errors.Wrap(err, "error occurred while reading certificate key file")
	}

	connSettings.DeviceCert = string(deviceCertRaw)
	connSettings.DeviceKey = string(deviceKeyRaw)

	if len(connSettings.DeviceID) == 0 {
		connSettings.DeviceID, err = util.ReadDeviceID(connSettings.DeviceCert)
		if err != nil {
			return err
		}
	}

	return nil
}

func parseConnectionString(connectionString string) (map[string]string, error) {
	properties := map[string]string{}
	for _, s := range strings.Split(connectionString, ";") {
		if len(s) == 0 {
			continue
		}
		kv := strings.SplitN(s, "=", 2)
		if len(kv) != 2 {
			return nil, errors.New("malformed connection string")
		}
		properties[kv[0]] = kv[1]
	}
	return properties, nil
}

func extractAzureHubName(hostName string) (string, error) {
	index := strings.Index(hostName, hostNameSuffix)
	if index == -1 {
		return "", errors.New("invalid HostName")
	}

	hubName := hostName[:index]
	if len(hubName) == 0 {
		return "", errors.New("the HubName cannot be empty")
	}

	return hubName, nil
}
