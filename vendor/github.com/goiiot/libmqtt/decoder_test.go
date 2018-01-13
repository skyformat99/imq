/*
 * Copyright GoIIoT (https://github.com/goiiot)
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package libmqtt

import (
	"bytes"
	"testing"
)

func TestDecodeRemainLength(t *testing.T) {
	buffer := &bytes.Buffer{}
	buffer.Write([]byte{0x04})
	length := decodeRemainLength(buffer)
	if length != 0x04 {
		t.Error(length)
	}
	buffer.Reset()
}

func TestDecodeOnePacket(t *testing.T) {
	// MQTT packet should work
	targetBytes := testConnWillMsgBytes
	buf := &bytes.Buffer{}

	if _, err := buf.Write(targetBytes); err != nil {
		t.Error(err)
	} else {
		pkt, err := DecodeOnePacket(V311, buf)
		if err != nil {
			t.Error(err)
		}
		buf.Reset()
		switch pkt.(type) {
		case *ConnPacket:
			EncodeOnePacket(V311, pkt, buf)
			pktBytes := buf.Bytes()
			if bytes.Compare(pktBytes, targetBytes) != 0 {
				t.Error(pktBytes)
			}
		default:
			t.Error(pkt)
		}
	}

	// malformed MQTT packets should fail
	buf.Reset()
	malformedConnBytes := []byte{
		0x10,                 // fixed header: conn:0
		38,                   // remaining length: 38
		0, 4, 77, 81, 84, 84, // Protocol Name: "MQTT"
		4,     // Protocol Level 3.1.1
		0xF6,  // connect flags: 11110110
		0, 10, // keepalive: 10s
		0, 4, 108, 111, 115, 116, // will topic: "lost"
		0, 5, 112, 101, 97, 99, 101, // will msg: "peace"
		// omit username field 0, 4, 117, 115, 101, 114, // Username: "user"
		0, 4, 112, 97, 115, 115, // Password: "pass"
		// another conn packet preventing EOF
		0x10,                 // fixed header: conn:0
		38,                   // remaining length: 38
		0, 4, 77, 81, 84, 84, // Protocol Name: "MQTT"
		4,     // Protocol Level 3.1.1
		0xF6,  // connect flags: 11110110
		0, 10, // keepalive: 10s
		0, 4, 108, 111, 115, 116, // will topic: "lost"
		0, 5, 112, 101, 97, 99, 101, // will msg: "peace"
		0, 4, 117, 115, 101, 114, // Username: "user"
		0, 4, 112, 97, 115, 115, // Password: "pass"
	}
	if _, err := buf.Write(malformedConnBytes); err != nil {
		t.Error(err)
	} else {
		if _, err := DecodeOnePacket(V311, buf); err == nil {
			t.Error("decoded conn packet, should not happen")
		}
	}
}

func BenchmarkDecodeOnePacket(b *testing.B) {
	b.StopTimer()
	buf := &bytes.Buffer{}
	for i := 0; i < b.N; i++ {
		buf.Write(testConnWillMsgBytes)
	}

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		_, err := DecodeOnePacket(V311, buf)
		if err != nil {
			b.Error(err)
		}
	}
}
