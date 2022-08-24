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
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net/url"
	"strconv"
	"time"

	"github.com/pkg/errors"
)

// Now represents the current time.
var Now = time.Now

const (
	// SASTokenValidityFactor represents the default factor of the SAS token validity period.
	// The token will be refreshed after SASTokenValidityFactor * defaultSASTokenValidity seconds.
	SASTokenValidityFactor  = 0.9
	defaultSASTokenValidity = time.Hour
)

// SharedAccessSignature represents the SAS access signature for generating SAS token for device authentication.
type SharedAccessSignature struct {
	Sr  string
	Sig string
	Se  time.Time
	Skn string
}

// GenerateSASToken generates the SAS token for device authentication.
func GenerateSASToken(connSettings *AzureConnectionSettings) *SharedAccessSignature {
	return newSharedAccessSignature(connSettings.HostName,
		connSettings.SharedAccessKeyName,
		connSettings.SharedAccessKey.SharedAccessKeyDecoded,
		Now().Add(connSettings.TokenValidity))
}

func sasTokenToString(sas *SharedAccessSignature) string {
	s := "SharedAccessSignature " +
		"sr=" + url.QueryEscape(sas.Sr) +
		"&sig=" + url.QueryEscape(sas.Sig) +
		"&se=" + url.QueryEscape(strconv.FormatInt(sas.Se.Unix(), 10))
	if sas.Skn != "" {
		s += "&skn=" + url.QueryEscape(sas.Skn)
	}
	return s
}

func newSharedAccessSignature(resource, policy string, decodedKey []byte, expiry time.Time) *SharedAccessSignature {
	sig := messageKeySignature(resource, decodedKey, expiry)
	return &SharedAccessSignature{
		Sr:  resource,
		Sig: sig,
		Se:  expiry,
		Skn: policy,
	}
}

func messageKeySignature(sr string, decodedKey []byte, se time.Time) string {
	h := hmac.New(sha256.New, decodedKey)
	fmt.Fprintf(h, "%s\n%d", url.QueryEscape(sr), se.Unix())
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

// ParseSASTokenValidity is a utility function that parses the string representation of the SAS token validity period
// to time.Duration value.
func ParseSASTokenValidity(sasTokenValidity string) (time.Duration, error) {
	if len(sasTokenValidity) == 0 {
		return -1, errors.New("empty SAS token validity value")
	}

	periodStr := sasTokenValidity[:len(sasTokenValidity)-1]
	period, err := strconv.Atoi(periodStr)
	if err != nil {
		return -1, errors.Wrapf(err, "invalid SAS token validity '%s'", sasTokenValidity)
	}

	if period < 1 {
		return -1, errors.Errorf("invalid SAS token validity '%s'", sasTokenValidity)
	}

	periodType := sasTokenValidity[len(sasTokenValidity)-1:]
	switch periodType {
	case "m":
		return time.Duration(period) * time.Minute, nil
	case "h":
		return time.Duration(period) * time.Hour, nil
	case "d":
		return 24 * time.Duration(period) * time.Hour, nil
	default:
		return -1, errors.Errorf("invalid SAS token validity '%s'", sasTokenValidity)
	}
}
