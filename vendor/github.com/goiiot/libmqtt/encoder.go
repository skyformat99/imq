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
	"errors"
)

var (
	// ErrUnsupportedVersion unsupported mqtt ProtoVersion
	ErrUnsupportedVersion = errors.New("trying encode unsupported mqtt ProtoVersion ")
	// ErrEncodeBadPacket happens when trying to encode none MQTT packet
	ErrEncodeBadPacket = errors.New("trying encode none MQTT packet ")
	// ErrEncodeLargePacket happens when mqtt packet is too large according to mqtt spec
	ErrEncodeLargePacket = errors.New("mqtt packet too large")
)

// Encode MQTT packet to bytes according to protocol ProtoVersion
func Encode(packet Packet, w BufferedWriter) error {
	switch packet.Version() {
	case V311:
		return encodeV311Packet(packet, w)
	case V5:
		return encodeV5Packet(packet, w)
	default:
		return ErrUnsupportedVersion
	}
}

// encode MQTT v3.1.1 packet to writer
func encodeV311Packet(pkt Packet, w BufferedWriter) error {
	if pkt == nil || w == nil {
		return nil
	}

	switch pkt.(type) {
	case *ConnPacket:
		c := pkt.(*ConnPacket)
		w.WriteByte(byte(CtrlConn << 4))
		payload := c.payload()
		writeVarInt(10+len(payload), w)
		w.Write(mqtt)
		w.WriteByte(byte(V311))
		w.WriteByte(c.flags())
		w.WriteByte(byte(c.Keepalive >> 8))
		w.WriteByte(byte(c.Keepalive))
		_, err := w.Write(payload)
		return err
	case *ConnAckPacket:
		c := pkt.(*ConnAckPacket)
		w.WriteByte(byte(CtrlConnAck << 4))
		w.WriteByte(0x02)
		w.WriteByte(boolToByte(c.Present))
		return w.WriteByte(c.Code)
	case *PublishPacket:
		p := pkt.(*PublishPacket)
		w.WriteByte(byte(CtrlPublish<<4) | boolToByte(p.IsDup)<<3 | boolToByte(p.IsRetain) | p.Qos<<1)
		payload := p.payload()
		writeVarInt(len(payload), w)
		_, err := w.Write(payload)
		return err
	case *PubAckPacket:
		p := pkt.(*PubAckPacket)
		w.WriteByte(byte(CtrlPubAck << 4))
		w.WriteByte(0x02)
		w.WriteByte(byte(p.PacketID >> 8))
		return w.WriteByte(byte(p.PacketID))
	case *PubRecvPacket:
		p := pkt.(*PubRecvPacket)
		w.WriteByte(byte(CtrlPubRecv << 4))
		w.WriteByte(0x02)
		w.WriteByte(byte(p.PacketID >> 8))
		return w.WriteByte(byte(p.PacketID))
	case *PubRelPacket:
		p := pkt.(*PubRelPacket)
		w.WriteByte(byte(CtrlPubRel<<4 | 0x02))
		w.WriteByte(0x02)
		w.WriteByte(byte(p.PacketID >> 8))
		return w.WriteByte(byte(p.PacketID))
	case *PubCompPacket:
		p := pkt.(*PubCompPacket)
		w.WriteByte(byte(CtrlPubComp << 4))
		w.WriteByte(0x02)
		w.WriteByte(byte(p.PacketID >> 8))
		return w.WriteByte(byte(p.PacketID))
	case *SubscribePacket:
		s := pkt.(*SubscribePacket)
		w.WriteByte(byte(CtrlSubscribe<<4 | 0x02))
		payload := s.payload()
		writeVarInt(2+len(payload), w)
		w.WriteByte(byte(s.PacketID >> 8))
		w.WriteByte(byte(s.PacketID))
		_, err := w.Write(payload)
		return err
	case *SubAckPacket:
		s := pkt.(*SubAckPacket)
		w.WriteByte(byte(CtrlSubAck << 4))
		payload := s.payload()
		writeVarInt(2+len(payload), w)
		w.WriteByte(byte(s.PacketID >> 8))
		w.WriteByte(byte(s.PacketID))
		_, err := w.Write(payload)
		return err
	case *UnSubPacket:
		s := pkt.(*UnSubPacket)
		w.WriteByte(byte(CtrlUnSub<<4 | 0x02))
		payload := s.payload()
		writeVarInt(2+len(payload), w)
		w.WriteByte(byte(s.PacketID >> 8))
		w.WriteByte(byte(s.PacketID))
		_, err := w.Write(payload)
		return err
	case *UnSubAckPacket:
		s := pkt.(*UnSubAckPacket)
		w.WriteByte(byte(CtrlUnSubAck << 4))
		w.WriteByte(0x02)
		w.WriteByte(byte(s.PacketID >> 8))
		return w.WriteByte(byte(s.PacketID))
	case *pingReqPacket:
		w.WriteByte(byte(CtrlPingReq << 4))
		return w.WriteByte(0x00)
	case *pingRespPacket:
		w.WriteByte(byte(CtrlPingResp << 4))
		return w.WriteByte(0x00)
	case *DisConnPacket:
		w.WriteByte(byte(CtrlDisConn << 4))
		return w.WriteByte(0x00)
	}

	return ErrEncodeBadPacket
}

// encode MQTT v5 packet to writer
func encodeV5Packet(pkt Packet, w BufferedWriter) error {
	if pkt == nil || w == nil {
		return nil
	}

	switch pkt.(type) {
	case *ConnPacket:
		c := pkt.(*ConnPacket)
		w.WriteByte(byte(CtrlConn << 4))

		props := c.Props.props()
		propLen := len(props)
		payload := c.payload()

		writeVarInt(10+len(payload)+propLen, w)
		w.Write(mqtt)
		w.WriteByte(byte(V5))
		w.WriteByte(c.flags())
		w.WriteByte(byte(c.Keepalive >> 8))
		w.WriteByte(byte(c.Keepalive))

		writeVarInt(propLen, w)
		w.Write(props)

		_, err := w.Write(payload)
		return err
	case *ConnAckPacket:
		c := pkt.(*ConnAckPacket)
		w.WriteByte(byte(CtrlConnAck << 4))

		props := c.Props.props()
		propLen := len(props)

		writeVarInt(2+propLen, w)
		w.WriteByte(boolToByte(c.Present))
		w.WriteByte(c.Code)

		writeVarInt(propLen, w)
		_, err := w.Write(props)

		return err
	case *PublishPacket:
		p := pkt.(*PublishPacket)
		w.WriteByte(byte(CtrlPublish<<4) | boolToByte(p.IsDup)<<3 | boolToByte(p.IsRetain) | p.Qos<<1)

		props := p.Props.props()
		propLen := len(props)
		payload := p.payload()

		writeVarInt(propLen+len(payload), w)

		writeVarInt(propLen, w)
		w.Write(props)

		_, err := w.Write(payload)
		return err
	case *PubAckPacket:
		p := pkt.(*PubAckPacket)
		w.WriteByte(byte(CtrlPubAck << 4))

		props := p.Props.props()
		propLen := len(props)
		writeVarInt(propLen+2, w)

		w.WriteByte(byte(p.PacketID >> 8))
		w.WriteByte(byte(p.PacketID))

		writeVarInt(propLen, w)
		_, err := w.Write(props)

		return err
	case *PubRecvPacket:
		p := pkt.(*PubRecvPacket)
		w.WriteByte(byte(CtrlPubRecv << 4))

		props := p.Props.props()
		propLen := len(props)
		writeVarInt(propLen+2, w)

		w.WriteByte(byte(p.PacketID >> 8))
		w.WriteByte(byte(p.PacketID))

		writeVarInt(propLen, w)
		_, err := w.Write(props)

		return err
	case *PubRelPacket:
		p := pkt.(*PubRelPacket)
		w.WriteByte(byte(CtrlPubRel<<4 | 0x02))

		props := p.Props.props()
		propLen := len(props)
		writeVarInt(propLen+2, w)

		w.WriteByte(byte(p.PacketID >> 8))
		w.WriteByte(byte(p.PacketID))

		writeVarInt(propLen, w)
		_, err := w.Write(props)

		return err
	case *PubCompPacket:
		p := pkt.(*PubCompPacket)
		w.WriteByte(byte(CtrlPubComp << 4))

		props := p.Props.props()
		propLen := len(props)
		writeVarInt(propLen+2, w)

		w.WriteByte(byte(p.PacketID >> 8))
		w.WriteByte(byte(p.PacketID))

		writeVarInt(propLen, w)
		_, err := w.Write(props)

		return err
	case *SubscribePacket:
		s := pkt.(*SubscribePacket)
		w.WriteByte(byte(CtrlSubscribe<<4 | 0x02))

		props := s.Props.props()
		payload := s.payload()
		propLen := len(props)

		writeVarInt(2+len(payload)+propLen, w)

		w.WriteByte(byte(s.PacketID >> 8))
		w.WriteByte(byte(s.PacketID))

		writeVarInt(propLen, w)
		w.Write(props)

		_, err := w.Write(payload)
		return err
	case *SubAckPacket:
		s := pkt.(*SubAckPacket)
		w.WriteByte(byte(CtrlSubAck << 4))

		props := s.Props.props()
		payload := s.payload()
		propLen := len(props)

		writeVarInt(2+len(payload)+propLen, w)

		w.WriteByte(byte(s.PacketID >> 8))
		w.WriteByte(byte(s.PacketID))

		writeVarInt(propLen, w)
		w.Write(props)

		_, err := w.Write(payload)
		return err
	case *UnSubPacket:
		s := pkt.(*UnSubPacket)
		w.WriteByte(byte(CtrlUnSub<<4 | 0x02))
		props := s.Props.props()
		payload := s.payload()
		propLen := len(props)

		writeVarInt(2+len(payload)+propLen, w)

		w.WriteByte(byte(s.PacketID >> 8))
		w.WriteByte(byte(s.PacketID))

		writeVarInt(propLen, w)
		w.Write(props)

		_, err := w.Write(payload)
		return err
	case *UnSubAckPacket:
		s := pkt.(*UnSubAckPacket)
		w.WriteByte(byte(CtrlUnSubAck << 4))
		w.WriteByte(0x02)
		w.WriteByte(byte(s.PacketID >> 8))
		w.WriteByte(byte(s.PacketID))

		props := s.Props.props()
		writeVarInt(len(props), w)
		_, err := w.Write(props)
		return err
	case *pingReqPacket:
		w.WriteByte(byte(CtrlPingReq << 4))
		return w.WriteByte(0x00)
	case *pingRespPacket:
		w.WriteByte(byte(CtrlPingResp << 4))
		return w.WriteByte(0x00)
	case *DisConnPacket:
		d := pkt.(*DisConnPacket)
		w.WriteByte(byte(CtrlDisConn << 4))
		props := d.Props.props()

		writeVarInt(len(props)+1, w)
		w.WriteByte(d.Code)
		writeVarInt(len(props), w)
		_, err := w.Write(props)
		return err
	case *AuthPacket:
		a := pkt.(*AuthPacket)
		w.WriteByte(byte(CtrlAuth << 4))
		props := a.Props.props()
		writeVarInt(1+len(props), w)
		w.WriteByte(a.Code)
		writeVarInt(len(props), w)
		_, err := w.Write(props)
		return err
	}

	return ErrEncodeBadPacket
}
