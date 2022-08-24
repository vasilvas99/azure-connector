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

package protobuf_test

import (
	"encoding/base64"
	"testing"

	"github.com/eclipse-kanto/azure-connector/routing/message/config"
	"github.com/eclipse-kanto/azure-connector/routing/message/protobuf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var testData = []struct {
	messageSubType string
	jsonPayload    string
	jsonString     string
	encodedPayload string
}{
	{
		"dummy-message",
		`{
			"value" : "dummy_value"
		}`,
		`{"value":"dummy_value"}`,
		"CgtkdW1teV92YWx1ZQ==",
	},
	{
		"dummy-message-missing",
		`{
			"value" : "dummy_value"
		}`,
		`{"value":"dummy_value"}`,
		"CgtkdW1teV92YWx1ZQ==",
	},
	{
		"simple-message",
		`{
			"name" : "dummy_name",
			"value" : "dummy_value"
		}`,
		`{"name":"dummy_name","value":"dummy_value"}`,
		"CgpkdW1teV9uYW1lEgtkdW1teV92YWx1ZQ==",
	},
	{
		"composite-message",
		`{
			"type" : "dummy_type",
			"composite" : {
				"type" : "composite_type",
				"composite" : {
					"name" : "dummy_name",
					"value" : "dummy_value"
				}
			}
		}`,
		`{"type":"dummy_type","composite":{"type":"composite_type","composite":{"name":"dummy_name","value":"dummy_value"}}}`,
		"CgpkdW1teV90eXBlEisKDmNvbXBvc2l0ZV90eXBlEhkKCmR1bW15X25hbWUSC2R1bW15X3ZhbHVl",
	},
	{
		"composite-repeated-message",
		`{
			"type" : "dummy_type",
			"composite" : [
				{
					"name" : "dummy_name_1",
					"value" : "dummy_value_1"
				},
				{
					"name" : "dummy_name_2",
					"value" : "dummy_value_2"
				}
			]
		}`,
		`{"type":"dummy_type","composite":[{"name":"dummy_name_1","value":"dummy_value_1"},{"name":"dummy_name_2","value":"dummy_value_2"}]}`,
		"CgpkdW1teV90eXBlEh0KDGR1bW15X25hbWVfMRINZHVtbXlfdmFsdWVfMRIdCgxkdW1teV9uYW1lXzISDWR1bW15X3ZhbHVlXzI=",
	},
	{
		"nested-message",
		`{
			"type" : "dummy_type",
			"nested" : {
				"name" : "dummy_name",
				"value" : "dummy_value"
			}
		}`,
		`{"type":"dummy_type","nested":{"name":"dummy_name","value":"dummy_value"}}`,
		"CgpkdW1teV90eXBlEhkKCmR1bW15X25hbWUSC2R1bW15X3ZhbHVl",
	},
	{
		"nested-repeated-message",
		`{
			"nested" : [
				{
					"name" : "dummy_name_1",
					"value" : "dummy_value_1"
				},
				{
					"name" : "dummy_name_2",
					"value" : "dummy_value_2"
				}
			]
		}`,
		`{"nested":[{"name":"dummy_name_1","value":"dummy_value_1"},{"name":"dummy_name_2","value":"dummy_value_2"}]}`,
		"Ch0KDGR1bW15X25hbWVfMRINZHVtbXlfdmFsdWVfMQodCgxkdW1teV9uYW1lXzISDWR1bW15X3ZhbHVlXzI=",
	},
	{
		"deep-nested-message",
		`{
			"type" : "dummy_type",
			"nested" : {
				"type" : "nested_type",
				"nested" : {
					"name" : "dummy_name",
					"value" : "dummy_value"
				}
			}
		}`,
		`{"type":"dummy_type","nested":{"type":"nested_type","nested":{"name":"dummy_name","value":"dummy_value"}}}`,
		"CgpkdW1teV90eXBlEigKC25lc3RlZF90eXBlEhkKCmR1bW15X25hbWUSC2R1bW15X3ZhbHVl",
	},
	{
		"multiple-nested-message",
		`{
			"type" : "dummy_type",
			"first_nested" : {
				"name" : "dummy_nested_1",
				"value" : "dummy_value_1"
			},
			"second_nested" : {
				"name" : "dummy_name_2",
				"value" : "dummy_value_2"
			}
		}`,
		`{"type":"dummy_type","firstNested":{"name":"dummy_nested_1","value":"dummy_value_1"},"secondNested":{"name":"dummy_name_2","value":"dummy_value_2"}}`,
		"CgpkdW1teV90eXBlEh8KDmR1bW15X25lc3RlZF8xEg1kdW1teV92YWx1ZV8xGh0KDGR1bW15X25hbWVfMhINZHVtbXlfdmFsdWVfMg==",
	},
	{
		"multiple-types-message",
		`{
			"value_string" : "dummy_string",
			"rep_value_string" : ["dummy_string_1", "dummy_string_2"],
			"value_int" : 1,
			"rep_value_int" : [1, 2, 3],
			"value_double" : 5.0,
			"rep_value_double" : [7.2, 7.1, 5.0],
			"value_bool" : true,
			"rep_value_bool" : [true, false]
		}`,
		`{"valueString":"dummy_string","repValueString":["dummy_string_1","dummy_string_2"],"valueInt":1,"repValueInt":[1,2,3],"valueDouble":5,"repValueDouble":[7.2,7.1,5],"valueBool":true,"repValueBool":[true,false]}`,
		"CgxkdW1teV9zdHJpbmcSDmR1bW15X3N0cmluZ18xEg5kdW1teV9zdHJpbmdfMhgBIgMBAgMpAAAAAAAAFEAyGM3MzMzMzBxAZmZmZmZmHEAAAAAAAAAUQDgBQgIBAA==",
	},
}

func TestMarshalPayloadEmptyProtoMessage(t *testing.T) {
	marshaller := createProtobufMarshaller(t)
	json := `{
		"name" : "dummy_name",
		"value" : "dummy_value
	}`
	_, err := marshaller.Marshal(1, "empty-messages", []byte(json))
	require.Error(t, err)
}

func TestMarshalPayloadInvalidProtoMessage(t *testing.T) {
	marshaller := createProtobufMarshaller(t)
	json := `{
		"name" : "dummy_name",
		"value" : "dummy_value
	}`
	_, err := marshaller.Marshal(1, "invalid-message", []byte(json))
	require.Error(t, err)
}

func TestMarshalPayloadForUnsupportedMessage(t *testing.T) {
	marshaller := createProtobufMarshaller(t)
	json := `{
		"value" : "dummy_value"
	}`
	_, err := marshaller.Marshal(1, "dummy-message-unsupported", []byte(json))
	require.Error(t, err)
}

func TestMarshalPayloadForUnsupportedMessageFromMultiMessageProtoFile(t *testing.T) {
	marshaller := createProtobufMarshaller(t)
	json := `{
		"name" : "dummy_name",
		"value" : "dummy_value"
	}`
	_, err := marshaller.Marshal(1, "multiple-messages-unsupported", []byte(json))
	require.Error(t, err)
}

func TestMarshalPayloadForMissingMessageFromMultiMessageProtoFile(t *testing.T) {
	marshaller := createProtobufMarshaller(t)
	json := `{
		"name" : "dummy_name",
		"value" : "dummy_value"
	}`
	_, err := marshaller.Marshal(1, "multiple-messages-missing", []byte(json))
	require.Error(t, err)
}

func TestMarshalPayloadForMessageSubTypeFromMissingProtoFile(t *testing.T) {
	marshaller := createProtobufMarshaller(t)
	json := `{
		"name" : "dummy_name",
		"value" : "dummy_value"
	}`
	_, err := marshaller.Marshal(1, "missing-messages-file", []byte(json))
	require.Error(t, err)
}

func TestMarshalPayloadForUnsupportedMessageType(t *testing.T) {
	marshaller := createProtobufMarshaller(t)
	json := `{
		"name" : "dummy_name",
		"value" : "dummy_value"
	}`
	_, err := marshaller.Marshal(2, "unsupported-mapping", []byte(json))
	require.Error(t, err)
}

func TestMarshalPayloadForUnsupportedMessageSubType(t *testing.T) {
	marshaller := createProtobufMarshaller(t)
	json := `{
		"name" : "dummy_name",
		"value" : "dummy_value"
	}`
	_, err := marshaller.Marshal(1, "unsupported-mapping", []byte(json))
	require.Error(t, err)
}

func TestMarshalMalformedPayload(t *testing.T) {
	marshaller := createProtobufMarshaller(t)
	json := `{
		"name" : "dummy_name",
		"value" : "dummy_value,
	}`
	_, err := marshaller.Marshal(1, "simple-message", []byte(json))
	require.Error(t, err)
}

func TestMarshalInvalidPayload(t *testing.T) {
	marshaller := createProtobufMarshaller(t)
	json := `{
		"name" : "dummy_name",
		"value" : {
			"x":"y"
		}
	}`
	_, err := marshaller.Marshal(1, "simple-message", []byte(json))
	require.Error(t, err)
}

func TestMarshalPayloadForCachedMessageSubType(t *testing.T) {
	marshaller := createProtobufMarshaller(t)
	json := `{
		"value" : "dummy_value"
	}`
	protobufPayload, err := marshaller.Marshal(1, "dummy-message", []byte(json))
	require.NoError(t, err)
	encodedPayload := base64.StdEncoding.EncodeToString(protobufPayload)
	assert.Equal(t, "CgtkdW1teV92YWx1ZQ==", encodedPayload)

	protobufPayload, err = marshaller.Marshal(1, "dummy-message", []byte(json))
	require.NoError(t, err)
	encodedPayload = base64.StdEncoding.EncodeToString(protobufPayload)
	assert.Equal(t, "CgtkdW1teV92YWx1ZQ==", encodedPayload)
}

func TestMarshalValidPayloadAndMessageSubType(t *testing.T) {
	marshaller := createProtobufMarshaller(t)
	for _, testValues := range testData {
		t.Run(testValues.messageSubType, func(t *testing.T) {
			protobufPayload, err := marshaller.Marshal(1, testValues.messageSubType, []byte(testValues.jsonPayload))
			require.NoError(t, err)
			encodedPayload := base64.StdEncoding.EncodeToString(protobufPayload)
			assert.Equal(t, testValues.encodedPayload, encodedPayload)
		})
	}
}

func TestUnmarshalPayloadEmptyProtoMessage(t *testing.T) {
	marshaller := createProtobufMarshaller(t)
	_, err := marshaller.Unmarshal("empty-messages", "CghpbmZsdXhkYg==")
	require.Error(t, err)
}

func TestUnmarshalPayloadInvalidProtoMessage(t *testing.T) {
	marshaller := createProtobufMarshaller(t)
	_, err := marshaller.Unmarshal("invalid-message", "CgpkdW1teV9uYW1lEgtkdW1teV92YWx1ZQ==")
	require.Error(t, err)
}

func TestUnmarshalPayloadForUnsupportedMessage(t *testing.T) {
	marshaller := createProtobufMarshaller(t)
	_, err := marshaller.Unmarshal("dummy-message-unsupported", "CghpbmZsdXhkYg==")
	require.Error(t, err)
}

func TestUnmarshalPayloadForUnsupportedMessageFromMultiMessageProtoFile(t *testing.T) {
	marshaller := createProtobufMarshaller(t)
	_, err := marshaller.Unmarshal("multiple-messages-unsupported", "CghpbmZsdXhkYg==")
	require.Error(t, err)
}

func TestUnmarshalPayloadForMissingMessageFromMultiMessageProtoFile(t *testing.T) {
	marshaller := createProtobufMarshaller(t)
	_, err := marshaller.Unmarshal("multiple-messages-missing", "CghpbmZsdXhkYg==")
	require.Error(t, err)
}

func TestUnmarshalPayloadForMessageSubTypeFromMissingProtoFile(t *testing.T) {
	marshaller := createProtobufMarshaller(t)
	_, err := marshaller.Unmarshal("missing-messages-file", "CghpbmZsdXhkYg==")
	require.Error(t, err)
}

func TestUnmarshalPayloadForUnsupportedMessageSubType(t *testing.T) {
	marshaller := createProtobufMarshaller(t)
	_, err := marshaller.Unmarshal("unsupported-mapping", "CghpbmZsdXhkYg==")
	require.Error(t, err)
}

func TestUnmarshalInvalidPayload(t *testing.T) {
	marshaller := createProtobufMarshaller(t)
	_, err := marshaller.Unmarshal("simple-message", "aZhp4M7m5pi=")
	require.Error(t, err)
}

func TestUnmarshalMalformedEncodedPayload(t *testing.T) {
	marshaller := createProtobufMarshaller(t)
	_, err := marshaller.Unmarshal("simple-message", "aZhp4M7m5pi===")
	require.Error(t, err)
}

func TestUnmarshalPayloadForCachedMessageSubType(t *testing.T) {
	marshaller := createProtobufMarshaller(t)
	jsonString := `{"value":"dummy_value"}`
	jsonPayload, err := marshaller.Unmarshal("dummy-message", "CgtkdW1teV92YWx1ZQ==")
	require.NoError(t, err)
	assert.Equal(t, jsonString, string(jsonPayload))

	jsonPayload, err = marshaller.Unmarshal("dummy-message", "CgtkdW1teV92YWx1ZQ==")
	require.NoError(t, err)
	assert.Equal(t, jsonString, string(jsonPayload))
}

func TestUnmarshalValidPayloadAndMessageSubType(t *testing.T) {
	marshaller := createProtobufMarshaller(t)
	for _, testValues := range testData {
		t.Run(testValues.messageSubType, func(t *testing.T) {
			jsonPayload, err := marshaller.Unmarshal(testValues.messageSubType, testValues.encodedPayload)
			require.NoError(t, err)
			assert.Equal(t, testValues.jsonString, string(jsonPayload))
		})
	}
}

func createProtobufMarshaller(t *testing.T) protobuf.Marshaller {
	mapperConfig, err := config.LoadMessageMapperConfig("testdata/message-mappings.json")
	require.NoError(t, err)
	return protobuf.NewProtobufJSONMarshaller(mapperConfig)
}
