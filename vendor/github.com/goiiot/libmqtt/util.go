/*
 * Copyright Go-IIoT (https://github.com/goiiot)
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
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"sync"
)

// BufferedWriter buffered writer, e.g. bufio.Writer, bytes.Buffer
type BufferedWriter interface {
	io.Writer
	io.ByteWriter
}

// BufferedReader buffered reader, e.g. bufio.Reader, bytes.Buffer
type BufferedReader interface {
	io.Reader
	io.ByteReader
}

func boolToByte(flag bool) byte {
	if flag {
		return 1
	}
	return 0
}

func recvKey(packetID uint16) string {
	return fmt.Sprintf("%s%d", "R", packetID)
}

func sendKey(packetID uint16) string {
	return fmt.Sprintf("%s%d", "S", packetID)
}

type idGenerator struct {
	usedIds *sync.Map
}

func newIDGenerator() *idGenerator {
	return &idGenerator{
		usedIds: &sync.Map{},
	}
}

func (g *idGenerator) next(extra interface{}) uint16 {
	var i uint16
	for i = 1; i < math.MaxUint16; i++ {
		if _, ok := g.usedIds.Load(i); !ok {
			g.usedIds.Store(i, extra)
			return i
		}
	}
	return 1
}

func (g *idGenerator) free(id uint16) {
	g.usedIds.Delete(id)
}

func (g *idGenerator) getExtra(id uint16) (interface{}, bool) {
	return g.usedIds.Load(id)
}

func putUint16(d []byte, v uint16) {
	binary.BigEndian.PutUint16(d[:], v)
}

func putUint32(d []byte, v uint32) {
	binary.BigEndian.PutUint32(d[:], v)
}

func encodeDataWithLen(data []byte) []byte {
	l := len(data)
	result := []byte{byte(l >> 8), byte(l)}
	return append(result, data...)
}

func writeVarInt(n int, w BufferedWriter) {
	if n < 0 || n > maxMsgSize {
		return
	}

	if n == 0 {
		w.WriteByte(0)
		return
	}

	for n > 0 {
		encodedByte := byte(n % 128)
		n /= 128
		if n > 0 {
			encodedByte |= 128
		}
		w.WriteByte(encodedByte)
	}
}

func getString(data []byte) (string, []byte, error) {
	b, next, err := getBinaryData(data)
	if err == nil {
		return string(b), next, err
	}

	return "", next, err
}

func getBinaryData(data []byte) ([]byte, []byte, error) {
	if len(data) < 2 {
		return nil, nil, ErrDecodeBadPacket
	}

	length := int(getUint16(data))
	if length+2 > len(data) {
		// out of bounds
		return nil, nil, ErrDecodeBadPacket
	}
	return data[2 : length+2], data[length+2:], nil
}

func getRemainLength(r io.ByteReader) (int, int) {
	var length, m uint32
	for m < 27 {
		b, err := r.ReadByte()
		if err != nil {
			return 0, 0
		}
		length |= uint32(b&127) << m
		if (b & 128) == 0 {
			break
		}
		m += 7
	}

	return int(length), int(m/7 + 1)
}

func getUint16(d []byte) uint16 {
	return binary.BigEndian.Uint16(d)
}

func getUint32(d []byte) uint32 {
	return binary.BigEndian.Uint32(d)
}

func getRawProps(data []byte) (map[byte][]byte, []byte) {
	propsLen, byteLen := getRemainLength(bytes.NewReader(data))
	propsBytes := data[1 : propsLen+byteLen]
	next := data[1+propsLen:]

	props := make(map[byte][]byte)
	for i := 0; i < propsLen; {
		var p []byte
		switch propsBytes[0] {
		case propKeyPayloadFormatIndicator:
			p = propsBytes[1:2]
		case propKeyMessageExpiryInterval:
			p = propsBytes[1:5]
		case propKeyContentType:
			p = propsBytes[1 : 3+getUint16(propsBytes[1:3])]
		case propKeyRespTopic:
			p = propsBytes[1 : 3+getUint16(propsBytes[1:3])]
		case propKeyCorrelationData:
			p = propsBytes[1 : 3+getUint16(propsBytes[1:3])]
		case propKeySubID:
			_, byteLen := getRemainLength(bytes.NewReader(propsBytes[1:]))
			p = propsBytes[1 : 1+byteLen]
		case propKeySessionExpiryInterval:
			p = propsBytes[1:5]
		case propKeyAssignedClientID:
			p = propsBytes[1 : 3+getUint16(propsBytes[1:3])]
		case propKeyServerKeepalive:
			p = propsBytes[1:3]
		case propKeyAuthMethod:
			p = propsBytes[1 : 3+getUint16(propsBytes[1:3])]
		case propKeyAuthData:
			p = propsBytes[1 : 3+getUint16(propsBytes[1:3])]
		case propKeyReqProblemInfo:
			p = propsBytes[1:2]
		case propKeyWillDelayInterval:
			p = propsBytes[1:5]
		case propKeyReqRespInfo:
			p = propsBytes[1:2]
		case propKeyRespInfo:
			p = propsBytes[1:2]
		case propKeyServerRef:
			p = propsBytes[1 : 3+getUint16(propsBytes[1:3])]
		case propKeyReasonString:
			p = propsBytes[1 : 3+getUint16(propsBytes[1:3])]
		case propKeyMaxRecv:
			p = propsBytes[1:3]
		case propKeyMaxTopicAlias:
			p = propsBytes[1:3]
		case propKeyTopicAlias:
			p = propsBytes[1:3]
		case propKeyMaxQos:
			p = propsBytes[1:2]
		case propKeyRetainAvail:
			p = propsBytes[1:2]
		case propKeyUserProps:
			keyEnd := 3 + getUint16(propsBytes[2:4])
			valEnd := keyEnd + getUint16(propsBytes[keyEnd:keyEnd+2])
			p = append(propsBytes[1:keyEnd], propsBytes[keyEnd:valEnd]...)
		case propKeyMaxPacketSize:
			p = propsBytes[1:5]
		case propKeyWildcardSubAvail:
			p = propsBytes[1:2]
		case propKeySubIDAvail:
			p = propsBytes[1:2]
		case propKeySharedSubAvail:
			p = propsBytes[1:2]
		}
		props[propsBytes[0]] = p
		propsBytes = propsBytes[1+len(p):]
		i += 1 + len(p)
	}

	return props, next
}

func getUserProps(data []byte) UserProperties {
	props := make(UserProperties)
	for str, next, _ := getString(data); next != nil; {
		var val string
		val, next, _ = getString(next)

		if _, ok := props[str]; ok {
			props[str] = append(props[str], val)
		} else {
			props[str] = []string{val}
		}
	}
	return props
}
