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

// ConnPacket is the first packet sent by Client to Server
type ConnPacket struct {
	protoName    string
	protoLevel   ProtocolVersion
	Username     string
	Password     string
	ClientID     string
	CleanSession bool
	IsWill       bool
	WillQos      QosLevel
	WillRetain   bool
	Keepalive    uint16
	WillTopic    string
	WillMessage  []byte
}

// Type ConnPacket's type is CtrlConn
func (c *ConnPacket) Type() CtrlType {
	return CtrlConn
}

func (c *ConnPacket) flags() byte {
	var flag byte
	if c.ClientID == "" {
		c.CleanSession = true
	}

	if c.CleanSession {
		flag |= 0x02
	}

	if c.IsWill {
		flag |= 0x04
		flag |= c.WillQos << 3

		if c.WillRetain {
			flag |= 0x20
		}
	}

	if c.Password != "" {
		flag |= 0x40
	}

	if c.Username != "" {
		flag |= 0x80
	}

	return flag
}

func (c *ConnPacket) payload() []byte {
	// client id
	result := encodeDataWithLen([]byte(c.ClientID))

	// will topic and message
	if c.IsWill {
		result = append(result, encodeDataWithLen([]byte(c.WillTopic))...)
		result = append(result, encodeDataWithLen(c.WillMessage)...)
	}

	if c.Username != "" {
		result = append(result, encodeDataWithLen([]byte(c.Username))...)
	}

	if c.Password != "" {
		result = append(result, encodeDataWithLen([]byte(c.Password))...)
	}

	return result
}

// ConnAckPacket is the packet sent by the Server in response to a ConnPacket
// received from a Client.
//
// The first packet sent from the Server to the Client MUST be a ConnAckPacket
type ConnAckPacket struct {
	Present bool
	Code    ConnAckCode
}

// Type ConnAckPacket's type is CtrlConnAck
func (c *ConnAckPacket) Type() CtrlType {
	return CtrlConnAck
}

var (
	// DisConnPacket is the final instance of disConnPacket
	DisConnPacket = &disConnPacket{}
)

// disConnPacket is the final Control Packet sent from the Client to the Server.
// It indicates that the Client is disconnecting cleanly.
type disConnPacket struct {
}

func (s *disConnPacket) Type() CtrlType {
	return CtrlDisConn
}
