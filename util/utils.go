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

package util

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"

	"github.com/pkg/errors"
)

// ContainsString returns true if a slice contains a given string.
func ContainsString(arr []string, target string) bool {
	for _, a := range arr {
		if a == target {
			return true
		}
	}
	return false
}

// DeleteFileIfEmpty deletes a file if empty.
func DeleteFileIfEmpty(file *os.File) {
	info, err := os.Stat(file.Name())
	if info == nil || os.IsNotExist(err) {
		return
	}
	if info.Size() == 0 {
		file.Close()
		os.Remove(file.Name())
	}
}

// DeviceCertificatesArePresent check if certificates are used
func DeviceCertificatesArePresent(cert string, key string) bool {
	return cert != "" && key != ""
}

// ReadDeviceID reads device id from PEM encoded certificate
func ReadDeviceID(deviceCert string) (string, error) {
	block, _ := pem.Decode([]byte(deviceCert))
	if block == nil {
		return "", errors.New("empty device certificate file content")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil || cert == nil {
		return "", errors.Wrap(err, "error on parsing the device certificate")
	}

	return cert.Subject.CommonName, nil
}

// GenerateCertKeyError generates certificare error
func GenerateCertKeyError(messagePart string, cert string, key string) error {
	var filesMissing string

	if len(cert) == 0 {
		if len(key) == 0 {
			filesMissing = "cert/key pair"
		} else {
			filesMissing = "client cert file"
		}
	} else {
		if len(key) == 0 {
			filesMissing = "private key file"
		}
	}

	return errors.New(fmt.Sprintf("missing %s and %s", messagePart, filesMissing))
}
