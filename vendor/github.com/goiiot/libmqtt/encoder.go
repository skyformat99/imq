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

import "errors"

var (
	errUnsupportedVersion = errors.New("trying encode unsupported mqtt version ")
	errEncodeBadPacket    = errors.New("trying encode none MQTT packet ")
)

// encode MQTT packet to bytes according to protocol version
func EncodeOnePacket(version ProtocolVersion, packet Packet, w BufferWriter) error {
	switch version {
	case V311:
		return encodeV311Packet(packet, w)
	case V5:
		return encodeV5Packet(packet, w)
	default:
		return errUnsupportedVersion
	}
}

// encode MQTT v3.1.1 packet to writer
func encodeV311Packet(pkt Packet, w BufferWriter) error {
	if pkt == nil || w == nil {
		return nil
	}

	switch pkt.(type) {
	case *ConnPacket:
		c := pkt.(*ConnPacket)
		w.WriteByte(CtrlConn << 4)
		payload := c.payload()
		writeRemainLength(10+len(payload), w)
		w.WriteByte(0x00)
		w.WriteByte(0x04)
		w.Write(mqtt)
		w.WriteByte(V311)
		w.WriteByte(c.flags())
		w.WriteByte(byte(c.Keepalive >> 8))
		w.WriteByte(byte(c.Keepalive))
		_, err := w.Write(payload)
		return err
	case *ConnAckPacket:
		c := pkt.(*ConnAckPacket)
		w.WriteByte(CtrlConnAck << 4)
		w.WriteByte(0x02)
		w.WriteByte(boolToByte(c.Present))
		return w.WriteByte(c.Code)
	case *PublishPacket:
		p := pkt.(*PublishPacket)
		w.WriteByte(CtrlPublish<<4 | boolToByte(p.IsDup)<<3 | boolToByte(p.IsRetain) | p.Qos<<1)
		payload := p.payload()
		writeRemainLength(len(payload), w)
		_, err := w.Write(payload)
		return err
	case *PubAckPacket:
		p := pkt.(*PubAckPacket)
		w.WriteByte(CtrlPubAck << 4)
		w.WriteByte(0x02)
		w.WriteByte(byte(p.PacketID >> 8))
		return w.WriteByte(byte(p.PacketID))
	case *PubRecvPacket:
		p := pkt.(*PubRecvPacket)
		w.WriteByte(CtrlPubRecv << 4)
		w.WriteByte(0x02)
		w.WriteByte(byte(p.PacketID >> 8))
		return w.WriteByte(byte(p.PacketID))
	case *PubRelPacket:
		p := pkt.(*PubRelPacket)
		w.WriteByte(CtrlPubRel<<4 | 0x02)
		w.WriteByte(0x02)
		w.WriteByte(byte(p.PacketID >> 8))
		return w.WriteByte(byte(p.PacketID))
	case *PubCompPacket:
		p := pkt.(*PubCompPacket)
		w.WriteByte(CtrlPubComp << 4)
		w.WriteByte(0x02)
		w.WriteByte(byte(p.PacketID >> 8))
		return w.WriteByte(byte(p.PacketID))
	case *SubscribePacket:
		s := pkt.(*SubscribePacket)
		w.WriteByte(CtrlSubscribe<<4 | 0x02)
		payload := s.payload()
		writeRemainLength(2+len(payload), w)
		w.WriteByte(byte(s.PacketID >> 8))
		w.WriteByte(byte(s.PacketID))
		_, err := w.Write(payload)
		return err
	case *SubAckPacket:
		s := pkt.(*SubAckPacket)
		w.WriteByte(CtrlSubAck << 4)
		payload := s.payload()
		writeRemainLength(2+len(payload), w)
		w.WriteByte(byte(s.PacketID >> 8))
		w.WriteByte(byte(s.PacketID))
		_, err := w.Write(payload)
		return err
	case *UnSubPacket:
		s := pkt.(*UnSubPacket)
		w.WriteByte(CtrlUnSub<<4 | 0x02)
		payload := s.payload()
		writeRemainLength(2+len(payload), w)
		w.WriteByte(byte(s.PacketID >> 8))
		w.WriteByte(byte(s.PacketID))
		_, err := w.Write(payload)
		return err
	case *UnSubAckPacket:
		s := pkt.(*UnSubAckPacket)
		w.WriteByte(CtrlUnSubAck << 4)
		w.WriteByte(0x02)
		w.WriteByte(byte(s.PacketID >> 8))
		return w.WriteByte(byte(s.PacketID))
	case *pingReqPacket:
		w.WriteByte(CtrlPingReq << 4)
		return w.WriteByte(0x00)
	case *pingRespPacket:
		w.WriteByte(CtrlPingResp << 4)
		return w.WriteByte(0x00)
	case *disConnPacket:
		w.WriteByte(CtrlDisConn << 4)
		return w.WriteByte(0x00)
	}

	return errEncodeBadPacket
}

// encode MQTT v5 packet to writer
func encodeV5Packet(pkt Packet, w BufferWriter) error {
	if pkt == nil || w == nil {
		return nil
	}

	switch pkt.Type() {
	case CtrlConn:
		c := pkt.(*ConnPacket)
		w.WriteByte(CtrlConn << 4)
		payload := c.payload()
		writeRemainLength(10+len(payload), w)
		w.WriteByte(0x00)
		w.WriteByte(0x04)
		w.Write(mqtt)
		w.WriteByte(V311)
		w.WriteByte(c.flags())
		w.WriteByte(byte(c.Keepalive >> 8))
		w.WriteByte(byte(c.Keepalive))
		_, err := w.Write(payload)
		return err
	case CtrlConnAck:
		c := pkt.(*ConnAckPacket)
		w.WriteByte(CtrlConnAck << 4)
		w.WriteByte(0x02)
		w.WriteByte(boolToByte(c.Present))
		return w.WriteByte(c.Code)
	case CtrlPublish:
		p := pkt.(*PublishPacket)
		w.WriteByte(CtrlPublish<<4 | boolToByte(p.IsDup)<<3 | boolToByte(p.IsRetain) | p.Qos<<1)
		payload := p.payload()
		writeRemainLength(len(payload), w)
		_, err := w.Write(payload)
		return err
	case CtrlPubAck:
		p := pkt.(*PubAckPacket)
		w.WriteByte(CtrlPubAck << 4)
		w.WriteByte(0x02)
		w.WriteByte(byte(p.PacketID >> 8))
		return w.WriteByte(byte(p.PacketID))
	case CtrlPubRecv:
		p := pkt.(*PubRecvPacket)
		w.WriteByte(CtrlPubRecv << 4)
		w.WriteByte(0x02)
		w.WriteByte(byte(p.PacketID >> 8))
		return w.WriteByte(byte(p.PacketID))
	case CtrlPubRel:
		p := pkt.(*PubRelPacket)
		w.WriteByte(CtrlPubRel<<4 | 0x02)
		w.WriteByte(0x02)
		w.WriteByte(byte(p.PacketID >> 8))
		return w.WriteByte(byte(p.PacketID))
	case CtrlPubComp:
		p := pkt.(*PubCompPacket)
		w.WriteByte(CtrlPubComp << 4)
		w.WriteByte(0x02)
		w.WriteByte(byte(p.PacketID >> 8))
		return w.WriteByte(byte(p.PacketID))
	case CtrlSubscribe:
		s := pkt.(*SubscribePacket)
		w.WriteByte(CtrlSubscribe<<4 | 0x02)
		payload := s.payload()
		writeRemainLength(2+len(payload), w)
		w.WriteByte(byte(s.PacketID >> 8))
		w.WriteByte(byte(s.PacketID))
		_, err := w.Write(payload)
		return err
	case CtrlSubAck:
		s := pkt.(*SubAckPacket)
		w.WriteByte(CtrlSubAck << 4)
		payload := s.payload()
		writeRemainLength(2+len(payload), w)
		w.WriteByte(byte(s.PacketID >> 8))
		w.WriteByte(byte(s.PacketID))
		_, err := w.Write(payload)
		return err
	case CtrlUnSub:
		s := pkt.(*UnSubPacket)
		w.WriteByte(CtrlUnSub<<4 | 0x02)
		payload := s.payload()
		writeRemainLength(2+len(payload), w)
		w.WriteByte(byte(s.PacketID >> 8))
		w.WriteByte(byte(s.PacketID))
		_, err := w.Write(payload)
		return err
	case CtrlUnSubAck:
		s := pkt.(*UnSubAckPacket)
		w.WriteByte(CtrlUnSubAck << 4)
		w.WriteByte(0x02)
		w.WriteByte(byte(s.PacketID >> 8))
		return w.WriteByte(byte(s.PacketID))
	case CtrlPingReq:
		w.WriteByte(CtrlPingReq << 4)
		return w.WriteByte(0x00)
	case CtrlPingResp:
		w.WriteByte(CtrlPingResp << 4)
		return w.WriteByte(0x00)
	case CtrlDisConn:
		w.WriteByte(CtrlDisConn << 4)
		return w.WriteByte(0x00)
	case CtrlAuth:
	}

	return errEncodeBadPacket
}

func encodeDataWithLen(data []byte) []byte {
	l := len(data)
	result := []byte{byte(l >> 8), byte(l)}
	return append(result, data...)
}

func writeRemainLength(n int, w BufferWriter) {
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
