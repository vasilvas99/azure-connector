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

package testing

import (
	"bytes"
	"errors"
	"io"
)

type errorReadWriterCloser struct{}

func (errorReadWriterCloser) Read(p []byte) (n int, err error) {
	return 0, errors.New("cannot read file contents")
}

func (errorReadWriterCloser) Write(p []byte) (n int, err error) {
	return 0, errors.New("cannot write file contents")
}

func (errorReadWriterCloser) Close() error {
	return nil
}

// CreateErrorReadWriterCloser is a creator method for an io.ReadWriteCloser implementation that fails with error.
func CreateErrorReadWriterCloser() io.ReadWriteCloser {
	return errorReadWriterCloser{}
}

type errorWriter struct {
	Buf bytes.Buffer
}

func (ew errorWriter) Read(p []byte) (n int, err error) {
	return ew.Buf.Read(p)
}

func (errorWriter) Write(p []byte) (n int, err error) {
	return 0, errors.New("cannot write file contents")
}

// CreateErrorWriter is a creator method for an io.ReadWriter implementation that fails with error.
func CreateErrorWriter() io.ReadWriter {
	return errorWriter{}
}
