// Copyright 2020 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package json

// Allow "encoding/json" import.
import (
	"bytes"
	"encoding/binary"
	"encoding/json" //nolint:depguard
	"io"

	goccyJson "github.com/goccy/go-json"
)

// but xorm still uses unmaintained "github.com/json-iterator/go"

// Encoder represents an encoder for json
type Encoder interface {
	Encode(v any) error
}

// Decoder represents a decoder for json
type Decoder interface {
	Decode(v any) error
}

// Interface represents an interface to handle json data
type Interface interface {
	Marshal(v any) ([]byte, error)
	Unmarshal(data []byte, v any) error
	NewEncoder(writer io.Writer) Encoder
	NewDecoder(reader io.Reader) Decoder
	Indent(dst *bytes.Buffer, src []byte, prefix, indent string) error
}

// Marshal converts object as bytes
func Marshal(v any) ([]byte, error) {
	return goccyJson.Marshal(v)
}

// Unmarshal decodes object from bytes
func Unmarshal(data []byte, v any) error {
	return goccyJson.Unmarshal(data, v)
}

// NewEncoder creates an encoder to write objects to writer
func NewEncoder(writer io.Writer) Encoder {
	return goccyJson.NewEncoder(writer)
}

// NewDecoder creates a decoder to read objects from reader
func NewDecoder(reader io.Reader) Decoder {
	return goccyJson.NewDecoder(reader)
}

// Indent appends to dst an indented form of the JSON-encoded src.
func Indent(dst *bytes.Buffer, src []byte, prefix, indent string) error {
	return goccyJson.Indent(dst, src, prefix, indent)
}

// MarshalIndent copied from encoding/json
func MarshalIndent(v any, prefix, indent string) ([]byte, error) {
	b, err := Marshal(v)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	err = Indent(&buf, b, prefix, indent)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// Valid proxy to json.Valid
func Valid(data []byte) bool {
	return goccyJson.Valid(data)
}

// UnmarshalHandleDoubleEncode - due to a bug in xorm (see https://gitea.com/xorm/xorm/pulls/1957) - it's
// possible that a Blob may be double encoded or gain an unwanted prefix of 0xff 0xfe.
func UnmarshalHandleDoubleEncode(bs []byte, v any) error {
	err := json.Unmarshal(bs, v)
	if err != nil {
		ok := true
		rs := []byte{}
		temp := make([]byte, 2)
		for _, rn := range string(bs) {
			if rn > 0xffff {
				ok = false
				break
			}
			binary.LittleEndian.PutUint16(temp, uint16(rn))
			rs = append(rs, temp...)
		}
		if ok {
			if len(rs) > 1 && rs[0] == 0xff && rs[1] == 0xfe {
				rs = rs[2:]
			}
			err = json.Unmarshal(rs, v)
		}
	}
	if err != nil && len(bs) > 2 && bs[0] == 0xff && bs[1] == 0xfe {
		err = json.Unmarshal(bs[2:], v)
	}
	return err
}
