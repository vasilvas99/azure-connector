#!/bin/sh

# Copyright (c) 2023 Contributors to the Eclipse Foundation
#
# See the NOTICE file(s) distributed with this work for additional
# information regarding copyright ownership.
#
# This program and the accompanying materials are made available under the
# terms of the Apache License 2.0 which is available at
# https://www.apache.org/licenses/LICENSE-2.0
#
# SPDX-License-Identifier: Apache-2.0
#

# adapted from https://github.com/eclipse-leda/leda-contrib-cloud-connector/blob/main/resources/cloudconnector_start.sh
# to fit: https://github.com/eclipse-kanto/azure-connector

AZURE_CONNECTOR_BIN="/app/azure-connector"
ARGUMENTS=""

# Check if cloudconnector is already started
[ -n "`pidof ${AZURE_CONNECTOR_BIN}`" ] && { echo "Cloud Connector already started, exiting!"; exit 1; }

# Configuration file in json format with flags values, configure with parameter -configFile.
[ -n "${CLOUD_CONNECTOR_CONFIG+x}" ] && ARGUMENTS="$ARGUMENTS -configFile=$CLOUD_CONNECTOR_CONFIG"

# Connection string for the Azure IoT Hub connectivity, configure with parameter -connectionString.
[ -n "${CONNECTION_STRING+x}" ] && ARGUMENTS="$ARGUMENTS -connectionString=$CONNECTION_STRING"

# ID Scope for the Azure DPS authentication, configure with parameter -idScope.
[ -n "${ID_SCOPE+x}" ] && ARGUMENTS="$ARGUMENTS -idScope=$ID_SCOPE"

# User-specified tenant id, configure with parameter -tenantId (default "defaultTenant").
[ -n "${TENANT_ID+x}" ] && ARGUMENTS="$ARGUMENTS -tenantId=$TENANT_ID"

# Address of the local MQTT broker, configure with parameter -localAddress (default "tcp://localhost:1883").
[ -n "${LOCAL_ADDRESS+x}" ] && ARGUMENTS="$ARGUMENTS -localAddress=$LOCAL_ADDRESS"

# Username for authentication to the local client, configure with parameter -localUsername.
[ -n "${LOCAL_USERNAME+x}" ] && ARGUMENTS="$ARGUMENTS -localUsername=$LOCAL_USERNAME"

# Password for authentication to the local client, configure with parameter -localPassword.
[ -n "${LOCAL_PASSWORD+x}" ] && ARGUMENTS="$ARGUMENTS -localPassword=$LOCAL_PASSWORD"

# Path to Hub certificate, configure with parameter -caCert (default "iothub.crt").
[ -n "${CA_CERT_PATH+x}" ] && ARGUMENTS="$ARGUMENTS -caCert=$CA_CERT_PATH"

# Log file location, configure with parameter -logFile (default "logs/log.txt")
[ -n "${LOG_FILE+x}" ] && ARGUMENTS="$ARGUMENTS -logFile=$LOG_FILE"

# Log level, configure with parameter -logLevel.
# Possible values: ERROR, WARN, INFO, DEBUG, TRACE (default "INFO").
[ -n "${LOG_LEVEL+x}" ] && ARGUMENTS="$ARGUMENTS -logLevel=$LOG_LEVEL"

# Log file size in MB before it gets rotated, configure with parameter -logFileSize (default 2).
[ -n "${LOG_FILE_SIZE+x}" ] && ARGUMENTS="$ARGUMENTS -logFileSize=$LOG_FILE_SIZE"

# Log file max rotations count, configure with parameter -logFileCount (default 5).
[ -n "${LOG_FILE_COUNT+x}" ] && ARGUMENTS="$ARGUMENTS -logFileCount=$LOG_FILE_COUNT"

# Log file rotations max age in days, configure with parameter -logFileMaxAge (default 28).
[ -n "${LOG_FILE_MAX_AGE+x}" ] && ARGUMENTS="$ARGUMENTS -logFileMaxAge=$LOG_FILE_MAX_AGE"

# A PEM encoded certificate file for cloud access
[ -n "${CERT_FILE+x}" ] && ARGUMENTS="$ARGUMENTS -cert=$CERT_FILE"

# A PEM encoded unencrypted private key file for cloud access
[ -n "${KEY_FILE+x}" ] && ARGUMENTS="$ARGUMENTS -key=$KEY_FILE"

# A PEM encoded local broker CA certificates file
[ -n "${LOCAL_CA_CERT_FILE+x}" ] && ARGUMENTS="$ARGUMENTS -localCACert=$LOCAL_CA_CERT_FILE"

#  A PEM encoded certificate file for local broker
[ -n "${LOCAL_CERT_FILE+x}" ] && ARGUMENTS="$ARGUMENTS -localCert=$LOCAL_CERT_FILE"

#  A PEM encoded unencrypted private key file for local broker
[ -n "${LOCAL_KEY_FILE+x}" ] && ARGUMENTS="$ARGUMENTS -localKey=$LOCAL_KEY_FILE"

#  The validity period for the generated SAS token for device authentication. Should be a positive integer number followed by a unit suffix, such as '300m', '1h', etc. Valid time units are 'm' (minutes), 'h' (hours), 'd' (days) (default "1h")
[ -n "${SAS_TOKEN_VALIDITY+x}" ] && ARGUMENTS="$ARGUMENTS -sasTokenValidity=$SAS_TOKEN_VALIDITY"

#  Path to the device file or the unix socket to access the TPM2
[ -n "${TPM_DEVICE+x}" ] && ARGUMENTS="$ARGUMENTS -tpmDevice=$TPM_DEVICE"

#  TPM2 storage root key handle
[ -n "${TPM_HANDLE+x}" ] && ARGUMENTS="$ARGUMENTS -tpmHandle=$TPM_HANDLE"

#  Private part of TPM2 key file
[ -n "${TPM_KEY+x}" ] && ARGUMENTS="$ARGUMENTS -tpmKey=$TPM_KEY"

#  Public part of TPM2 key file
[ -n "${TPM_KEY_PUB+x}" ] && ARGUMENTS="$ARGUMENTS -tpmKeyPub=$TPM_KEY_PUB"

echo "Starting with arguments: $AZURE_CONNECTOR_BIN $ARGUMENTS"
$AZURE_CONNECTOR_BIN $ARGUMENTS;
