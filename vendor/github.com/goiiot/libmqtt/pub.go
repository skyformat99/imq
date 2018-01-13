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

// PublishPacket is sent from a Client to a Server or from Server to a Client
// to transport an Application Message.
type PublishPacket struct {
	IsDup     bool
	Qos       QosLevel
	IsRetain  bool
	TopicName string
	Payload   []byte
	PacketID  uint16
}

// Type PublishPacket's type is CtrlPublish
func (p *PublishPacket) Type() CtrlType {
	return CtrlPublish
}

func (p *PublishPacket) payload() []byte {
	data := encodeDataWithLen([]byte(p.TopicName))
	if p.Qos > Qos0 {
		data = append(data, byte(p.PacketID>>8), byte(p.PacketID))
	}
	return append(data, p.Payload...)
}

// PubAckPacket is the response to a PublishPacket with QoS level 1.
type PubAckPacket struct {
	PacketID uint16
}

// Type PubAckPacket's type is CtrlPubAck
func (p *PubAckPacket) Type() CtrlType {
	return CtrlPubAck
}

// PubRecvPacket is the response to a PublishPacket with QoS 2.
// It is the second packet of the QoS 2 protocol exchange.
type PubRecvPacket struct {
	PacketID uint16
}

// Type PubRecvPacket's type is CtrlPubRecv
func (p *PubRecvPacket) Type() CtrlType {
	return CtrlPubRecv
}

// PubRelPacket is the response to a PubRecvPacket.
// It is the third packet of the QoS 2 protocol exchange.
type PubRelPacket struct {
	PacketID uint16
}

// Type PubRelPacket's type is CtrlPubRel
func (p *PubRelPacket) Type() CtrlType {
	return CtrlPubRel
}

// PubCompPacket is the response to a PubRelPacket.
// It is the fourth and final packet of the QoS 892 2 protocol exchange. 893
type PubCompPacket struct {
	PacketID uint16
}

// Type PubCompPacket's type is CtrlPubComp
func (p *PubCompPacket) Type() CtrlType {
	return CtrlPubComp
}
