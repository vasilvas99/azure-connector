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

package util_test

import (
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/eclipse-kanto/azure-connector/util"
	"github.com/stretchr/testify/assert"
)

const (
	certKey = `-----BEGIN RSA PRIVATE KEY-----
MIIEowIBAAKCAQEAxwOEtbC1X+zSg8c9gJsDu8QzK/mr0b54tz2st9chiMdxdw9p
-----END RSA PRIVATE KEY-----`

	cert = `-----BEGIN CERTIFICATE-----
MIIEWzCCAkMCFEhFajH0fSSgRaqg8mEnozLvxQHFMA0GCSqGSIb3DQEBCwUAMGwx
CzAJBgNVBAYTAkJHMQ4wDAYDVQQIDAVTb2ZpYTEOMAwGA1UEBwwFU29maWExETAP
BgNVBAoMCEJvc2NoLklPMQwwCgYDVQQLDANQQVAxHDAaBgNVBAMME1NEVl9Cb3Nj
aF9UZXN0X0NlcnQwHhcNMjExMDA2MTgzMjUzWhcNMjcwNzA3MTgzMjUzWjBoMQsw
CQYDVQQGEwJCRzEOMAwGA1UECAwFU29maWExDjAMBgNVBAcMBVNvZmlhMREwDwYD
VQQKDAhCb3NjaC5JTzEMMAoGA1UECwwDUEFQMRgwFgYDVQQDDA9wYXAtZGVtby1k
ZXZpY2UwggEiMA0GCSqGSIb3DQEBAQUAA4IBDwAwggEKAoIBAQDHA4S1sLVf7NKD
xz2AmwO7xDMr+avRvni3Pay31yGIx3F3D2lRCoSNk6lpoxlMSWFcasR6CEODylA/
mmcSgfQyKGVOOuRir7asSOlyahpeMI0cAfcf6FNX03s55gS4Tmjn7TYpdwTMSWYK
ie4Q0XnyO7uJhK6UcPL9s2nYV564geMwnbBGqMPOBnLjjSWDG1Cik3PHcLmsmg0a
2cYYmzglr4qRd23n/gLV19t6XoNV7zTAwD2jMuvnHASR9ITO2rywur9yeQmzx5XH
Pg3Dgi49NuRKHeuTEoI1N5YuzcnEZNA6UIH3yaP0bfZj6mFymlzETkmjPi+yzb6j
/gH9D+51AgMBAAEwDQYJKoZIhvcNAQELBQADggIBAIIrIluWwIeB/PJun2hAxW+C
X5wVcU1eSpn5IN25jRyAvRTpTVlxFpUPXA3W5sb+TWTI1HsRToHHAseBO1c9PnlB
/zQCgWllpkNUNXvyyhDSSInRdlFpg4Ju0A5SidJZjaMwfEA9xUGxt5iiwHBG+uHA
4LvPEJLhXJv7Jyr1oYaI6WbByLF6/7GWbna0ckRZ53mSCojuG/ldWObLKpAw0mrD
/QQPvfA89EG/r6iSeRvdK+mdYMexEmCIJPYz4p3pRcvLIBL6mVE8kFFOmcL5Kpit
BR11Y/+99HiQRFXZTqfQsE2P40wvkKbg4q6cK4D4cWIv+DjhIrsUbsrIdX1VsoQW
C5Lnd84U/simYPGKjLJIILKNLHKdvgOZnZIaMPvFvmG4Cge/icawwyREuGzSXAEK
HDCTv/fMj+xj7vNw61j0q7qdQfwJPoHkYlbkIijl1k3J49VPtmQSDiMKgg2gPVoF
qy13v40DS89APVXL90iyfxTguKGn0Ocbd89WnC7A/JjXBODKNz3EOeKmnpxBbPx0
cUb7TPRTNUEhX+wuLtnRWOGbYaWuK1Pg/6mEswLH3v0sMVcKgiovDs1PkyJ1l8ag
Fa05mI5Z1RfwhcN5b06z3GhgJYM1BoimyLLlt0+0LEYZ7/6Xt2ImueYk5bsRcGSQ
XuqPEyIaaxYwg0+et0CB
-----END CERTIFICATE-----`
)

func TestContainsString(t *testing.T) {
	arrays := []string{"array.value.name.1", "array.value.name.2", "array.value.name.3"}

	targets := []string{"", "array", "array.value", "array.value.name", "array.value.name.", "array.value.name.11"}

	for _, target := range targets {
		res := util.ContainsString(arrays, target)
		assert.False(t, res)
	}

	assert.True(t, util.ContainsString(arrays, "array.value.name.3"))
}

func TestDeleteFileIfEmpty(t *testing.T) {
	fileName := "test_file_for_delete.out"
	file, err := os.Create(fileName)
	assert.NoError(t, err)
	util.DeleteFileIfEmpty(file)

	_, err = os.Stat(file.Name())
	assert.True(t, os.IsNotExist(err))

	// try to delete non existing file
	util.DeleteFileIfEmpty(file)
}

func TestDeviceCertificatesArePresent(t *testing.T) {
	assert.False(t, util.DeviceCertificatesArePresent("", ""))
	assert.False(t, util.DeviceCertificatesArePresent("cert", ""))
	assert.False(t, util.DeviceCertificatesArePresent("", "key"))

	assert.True(t, util.DeviceCertificatesArePresent("cert", "key"))
}

func TestReadDeviceID(t *testing.T) {
	err := readDeviceId(t, cert)
	assert.NoError(t, err)

	err = readDeviceId(t, certKey)
	assert.Error(t, err)

	err = readDeviceId(t, `-----BEGIN CERTIFICATE----------END CERTIFICATE-----`)
	assert.Error(t, err)
}

func readDeviceId(t *testing.T, certStr string) error {
	certFileReader := strings.NewReader(certStr)
	deviceCertRaw, err := ioutil.ReadAll(certFileReader)
	assert.NoError(t, err)

	_, err = util.ReadDeviceID(string(deviceCertRaw))
	return err
}

func TestGenerateCertKeyError(t *testing.T) {
	assert.Error(t, util.GenerateCertKeyError("connectionString", "", ""))
	assert.Error(t, util.GenerateCertKeyError("connectionString", "cert", ""))
	assert.Error(t, util.GenerateCertKeyError("connectionString", "", "key"))
}
