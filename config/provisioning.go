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

package config

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/pkg/errors"

	"github.com/eclipse-kanto/suite-connector/logger"
)

const (
	azureDPSRegisterRequestURL      = "https://global.azure-devices-provisioning.net/%s/registrations/%s/register?api-version=2021-06-01"
	azureDPSGetDeviceInfoRequestURL = "https://global.azure-devices-provisioning.net/%s/registrations/%s/operations/%s?api-version=2021-06-01"

	contentTypeHeaderKey       = "Content-Type"
	applicationJSONHeaderValue = "application/json"
	retryAfterHeaderKey        = "Retry-After"
)

// ProvisioningService abstracts the access to the provisioning services (PSS & Azure DPS).
type ProvisioningService interface {
	Init(client ProvisioningHTTPClient, provisioningFile io.ReadWriter)
	GetDeviceData(idScope string, connSettings *AzureConnectionSettings) (*AzureDeviceData, error)
}

type defProvisioningService struct {
	client           ProvisioningHTTPClient
	provisioningFile io.ReadWriter
	logger           logger.Logger
}

type defHTTPClient struct {
	client *http.Client
}

// ProvisioningHTTPClient is used as a wrapper of the HTTP client for accessing the provisioning services (PSS & Azure DPS).
type ProvisioningHTTPClient interface {
	Get(url string) (*http.Response, error)
	Post(url string, contentType string, body io.Reader) (*http.Response, error)
	Do(req *http.Request) (*http.Response, error)
}

// IDScopeProvider is used as ID scope provider
type IDScopeProvider func(connSettings *AzureConnectionSettings) (string, error)

func (p *defHTTPClient) Get(url string) (*http.Response, error) {
	return p.client.Get(url)
}

func (p *defHTTPClient) Post(url string, contentType string, body io.Reader) (*http.Response, error) {
	return p.client.Post(url, contentType, body)
}

func (p *defHTTPClient) Do(req *http.Request) (*http.Response, error) {
	return p.client.Do(req)
}

// NewProvisioningService is a creator method for instantiating a provisioning service instance.
func NewProvisioningService(logger logger.Logger) ProvisioningService {
	return &defProvisioningService{
		logger: logger,
	}
}

// NewHTTPClient is a creator method for instantiating a provisioning HTTP client.
func NewHTTPClient(client *http.Client) ProvisioningHTTPClient {
	return &defHTTPClient{
		client: client,
	}
}

func (p *defProvisioningService) Init(client ProvisioningHTTPClient, provisioningFile io.ReadWriter) {
	p.client = client
	p.provisioningFile = provisioningFile
}

func (p *defProvisioningService) GetDeviceData(idScope string, connSettings *AzureConnectionSettings) (*AzureDeviceData, error) {
	deviceData, err := p.getDeviceDataFromDisk()
	if err != nil {
		return nil, err
	}
	if deviceData != nil {
		return deviceData, err
	}

	if p.client == nil {
		return nil, errors.New("error HTTP client not initialized")
	}

	azureDeviceInfo, err := getDeviceInfoRequest(idScope, p.client, connSettings)
	if err != nil {
		return nil, err
	}
	if err = p.persistDeviceInfoToDisk(azureDeviceInfo); err != nil {
		p.logger.Warn("Error occurred while writing file to disk", err, nil)
	}

	return extractDeviceDataFromResponse(azureDeviceInfo)
}

func (p *defProvisioningService) getDeviceDataFromDisk() (*AzureDeviceData, error) {
	fileContents, err := ioutil.ReadAll(p.provisioningFile)
	if err != nil {
		return nil, err
	}
	if len(fileContents) == 0 {
		return nil, nil
	}
	deviceData := &AzureDeviceData{}

	err = json.Unmarshal(fileContents, deviceData)
	if err != nil {
		return nil, errors.Wrap(err, "error on unmarshalling provisioning file")
	}
	if *deviceData == (AzureDeviceData{}) {
		return nil, errors.New("provisioning.json is empty")
	}

	err = deviceData.validate()
	if err != nil {
		return nil, err
	}

	return deviceData, nil
}

func (p *defProvisioningService) persistDeviceInfoToDisk(deviceInfo *AzureDpsDeviceInfoResponse) error {
	persistInfo, err := extractDeviceDataFromResponse(deviceInfo)
	if err != nil {
		return err
	}

	file, err := json.Marshal(persistInfo)
	if err != nil {
		return errors.Wrap(err, "error on marshalling deviceInfo")
	}

	_, err = p.provisioningFile.Write(file)
	if err != nil {
		return errors.Wrap(err, "error on persisting provisioning data")
	}

	return nil
}

func extractDeviceDataFromResponse(deviceInfo *AzureDpsDeviceInfoResponse) (*AzureDeviceData, error) {
	if deviceInfo == nil || *deviceInfo == (AzureDpsDeviceInfoResponse{}) ||
		deviceInfo.RegistrationState == (AzureDpsRegistrationState{}) {
		return nil, errors.New("error on mapping device data: deviceInfo cannot be empty")
	}

	persistInfo := &AzureDeviceData{
		AssignedHub: deviceInfo.RegistrationState.AssignedHub,
		DeviceID:    deviceInfo.RegistrationState.DeviceID,
	}

	err := persistInfo.validate()
	if err != nil {
		return nil, err
	}
	return persistInfo, nil
}

func getDeviceInfoRequest(
	idScope string, client ProvisioningHTTPClient, connSettings *AzureConnectionSettings,
) (*AzureDpsDeviceInfoResponse, error) {
	if len(idScope) == 0 {
		return nil, errors.New("idScope cannot be empty")
	}

	azureRegistrationInfo, retryPeriod, err := registerAzureDeviceInDPS(idScope, connSettings, client)
	if err != nil {
		return nil, err
	}
	if azureRegistrationInfo.RegistrationState != (AzureDpsRegistrationState{}) {
		return azureRegistrationInfo, nil
	}

	retrySeconds, err := strconv.Atoi(retryPeriod)
	if err != nil {
		return nil, errors.Wrap(err, "error on retry period creation")
	}
	time.Sleep(time.Duration(retrySeconds) * time.Second)

	url := fmt.Sprintf(azureDPSGetDeviceInfoRequestURL, idScope, connSettings.DeviceID, azureRegistrationInfo.OperationID)

	res, err := client.Get(url)
	if resErr := parseResponseError(err, http.StatusOK, res); resErr != nil {
		return nil, errors.Wrap(resErr, "error on getting device info from AzureDPS")
	}

	resBody, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, errors.Wrap(err, "error on reading deviceInfo response body")
	}

	azureDeviceInfo := &AzureDpsDeviceInfoResponse{}
	err = json.Unmarshal(resBody, azureDeviceInfo)
	if err != nil {
		return nil, errors.Wrap(err, "error on unmarshalling GET device info from AzureDPS response body")
	}

	return azureDeviceInfo, nil
}

func registerAzureDeviceInDPS(idScope string, connSettings *AzureConnectionSettings, client ProvisioningHTTPClient) (*AzureDpsDeviceInfoResponse, string, error) {
	azureDpsReq := &AzureDpsRegisterDeviceRequest{
		RegistrationID: connSettings.DeviceID,
	}

	jsonBody, err := json.Marshal(azureDpsReq)
	if err != nil {
		return nil, "", errors.Wrap(err, "error on marshalling register to AzureDPS request body")
	}

	requestBody := bytes.NewBuffer(jsonBody)
	url := fmt.Sprintf(azureDPSRegisterRequestURL, idScope, connSettings.DeviceID)
	request, _ := http.NewRequest("PUT", url, requestBody)
	request.Header.Set(contentTypeHeaderKey, applicationJSONHeaderValue)
	res, err := client.Do(request)
	if resErr := parseResponseError(err, http.StatusAccepted, res); resErr != nil {
		return nil, "", errors.Wrap(resErr, "error on registering device to AzureDPS")
	}

	resBody, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, "", errors.Wrap(err, "error on reading register Azure device response body")
	}

	retryPeriod := res.Header.Get(retryAfterHeaderKey)
	registerAzureDevice := &AzureDpsDeviceInfoResponse{}
	err = json.Unmarshal(resBody, registerAzureDevice)
	if err != nil {
		return nil, "", errors.Wrap(err, "error on unmarshalling register to AzureDPS response body")
	}
	return registerAzureDevice, retryPeriod, nil
}

func parseResponseError(err error, expectedStatusCode int, response *http.Response) error {
	if err != nil {
		return err
	}
	if response == nil {
		return errors.New("response cannot be empty")
	}

	if response.StatusCode != expectedStatusCode {
		resBody, _ := ioutil.ReadAll(response.Body)
		resError := &ResponseError{}
		err = json.Unmarshal(resBody, resError)
		if err != nil {
			return errors.New("cannot unmarshal response error")
		}

		return errors.New(fmt.Sprintf("expected StatusCode %d, but got %d, message: %s%s",
			expectedStatusCode, response.StatusCode, resError.Message, resError.Detail))
	}
	return nil
}

func initDeviceProvisioningClient(connSettings *AzureConnectionSettings) (ProvisioningHTTPClient, error) {
	tlsConfig, err := createProvisioningTLSConfig(connSettings)
	if err != nil {
		return nil, err
	}

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsConfig,
		},
	}

	return NewHTTPClient(client), nil
}

func createProvisioningTLSConfig(connSettings *AzureConnectionSettings) (*tls.Config, error) {
	certificatePair, err := tls.X509KeyPair([]byte(connSettings.DeviceCert), []byte(connSettings.DeviceKey))
	if err != nil {
		return nil, errors.Wrap(err, "error on loading X509 Key Pair")
	}

	return &tls.Config{
		ClientCAs:          nil,
		InsecureSkipVerify: false,
		Certificates:       []tls.Certificate{certificatePair},
	}, nil
}
