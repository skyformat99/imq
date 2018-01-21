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
	"io"
)

var (
	// ErrDecodeBadPacket is the error happened when trying to decode a none MQTT packet
	ErrDecodeBadPacket = errors.New("decoded none MQTT packet ")

	// ErrDecodeNoneV311Packet is the error happened when
	// trying to decode mqtt 3.1.1 packet but got other mqtt packet ProtoVersion
	ErrDecodeNoneV311Packet = errors.New("decoded none MQTT v3.1.1 packet ")

	// ErrDecodeNoneV5Packet is the error happened when
	// trying to decode mqtt 5 packet but got other mqtt packet ProtoVersion
	ErrDecodeNoneV5Packet = errors.New("decoded none MQTT v5 packet ")
)

// Decode will decode one mqtt packet
func Decode(version ProtoVersion, reader BufferedReader) (Packet, error) {
	switch version {
	case V311:
		return decodeV311Packet(reader)
	case V5:
		return decodeV5Packet(reader)
	default:
		return nil, ErrUnsupportedVersion
	}
}

// decode mqtt v3.1.1 packets
func decodeV311Packet(r BufferedReader) (Packet, error) {
	header, err := r.ReadByte()
	if err != nil {
		return nil, err
	}

	bytesToRead, _ := getRemainLength(r)
	if bytesToRead == 0 {
		switch CtrlType(header >> 4) {
		case CtrlPingReq:
			return PingReqPacket, nil
		case CtrlPingResp:
			return PingRespPacket, nil
		case CtrlDisConn:
			return &DisConnPacket{}, nil
		default:
			return nil, ErrDecodeBadPacket
		}
	} else if bytesToRead < 2 {
		return nil, ErrDecodeBadPacket
	}

	body := make([]byte, bytesToRead)
	if _, err = io.ReadFull(r, body[:]); err != nil {
		return nil, err
	}

	switch CtrlType(header >> 4) {
	case CtrlConn:
		protocol, body, err := getString(body)
		if err != nil {
			return nil, err
		}

		if body[0] != byte(V311) {
			return nil, ErrDecodeNoneV311Packet
		}

		if len(body) < 4 {
			return nil, ErrDecodeBadPacket
		}
		hasUsername := body[1]&0x80 == 0x80
		hasPassword := body[1]&0x40 == 0x40
		pkt := &ConnPacket{
			ProtoName:    protocol,
			CleanSession: body[1]&0x02 == 0x02,
			IsWill:       body[1]&0x04 == 0x04,
			WillQos:      body[1] & 0x18 >> 3,
			WillRetain:   body[1]&0x20 == 0x20,
			Keepalive:    getUint16(body[2:4]),
		}
		pkt.ProtoVersion = ProtoVersion(body[0])

		if pkt.ClientID, body, err = getString(body[4:]); err != nil {
			return nil, err
		}

		if pkt.IsWill {
			pkt.WillTopic, body, err = getString(body)
			pkt.WillMessage, body, err = getBinaryData(body)
		}

		if hasUsername {
			pkt.Username, body, err = getString(body)
		}

		if hasPassword {
			pkt.Password, _, err = getString(body)
		}

		if err != nil {
			return nil, err
		}

		return pkt, nil
	case CtrlConnAck:
		return &ConnAckPacket{Present: body[0]&0x01 == 0x01, Code: body[1]}, nil
	case CtrlPublish:
		topicName, body, err := getString(body)
		if err != nil {
			return nil, err
		}

		if len(body) < 2 {
			return nil, ErrDecodeBadPacket
		}

		pub := &PublishPacket{
			IsDup:     header&0x08 == 0x08,
			Qos:       header & 0x06 >> 1,
			IsRetain:  header&0x01 == 1,
			TopicName: topicName,
		}

		if pub.Qos > Qos0 {
			pub.PacketID = getUint16(body)
			body = body[2:]
		}

		pub.Payload = body
		return pub, nil
	case CtrlPubAck:
		return &PubAckPacket{PacketID: getUint16(body)}, nil
	case CtrlPubRecv:
		return &PubRecvPacket{PacketID: getUint16(body)}, nil
	case CtrlPubRel:
		return &PubRelPacket{PacketID: getUint16(body)}, nil
	case CtrlPubComp:
		return &PubCompPacket{PacketID: getUint16(body)}, nil
	case CtrlSubscribe:
		pkt := &SubscribePacket{
			PacketID: getUint16(body),
			Topics:   make([]*Topic, 0),
		}

		body = body[2:]
		for len(body) > 0 {
			var name string
			if name, body, err = getString(body); err != nil {
				return nil, err
			}

			if len(body) < 1 {
				return nil, ErrDecodeBadPacket
			}

			pkt.Topics = append(pkt.Topics, &Topic{Name: name, Qos: body[0]})
			body = body[1:]
		}
		return pkt, nil
	case CtrlSubAck:
		pkt := &SubAckPacket{
			PacketID: getUint16(body),
			Codes:    make([]byte, 0),
		}

		body = body[2:]
		for i := range body {
			pkt.Codes = append(pkt.Codes, body[i])
		}
		return pkt, nil
	case CtrlUnSub:
		pkt := &UnSubPacket{
			PacketID:   getUint16(body),
			TopicNames: make([]string, 0),
		}

		body = body[2:]
		for len(body) > 0 {
			var name string
			name, body, err = getString(body)
			if err != nil {
				return nil, err
			}
			pkt.TopicNames = append(pkt.TopicNames, name)
		}
		return pkt, nil
	case CtrlUnSubAck:
		return &UnSubAckPacket{PacketID: getUint16(body)}, nil
	}

	return nil, ErrDecodeBadPacket
}

// decode mqtt v5 packets
func decodeV5Packet(r BufferedReader) (Packet, error) {
	header, err := r.ReadByte()
	if err != nil {
		return nil, err
	}

	bytesToRead, _ := getRemainLength(r)
	if bytesToRead == 0 {
		switch CtrlType(header >> 4) {
		case CtrlPingReq:
			return PingReqPacket, nil
		case CtrlPingResp:
			return PingRespPacket, nil
		default:
			return nil, ErrDecodeBadPacket
		}
	} else if bytesToRead < 2 {
		return nil, ErrDecodeBadPacket
	}

	body := make([]byte, bytesToRead)
	if _, err = io.ReadFull(r, body[:]); err != nil {
		return nil, err
	}

	switch CtrlType(header >> 4) {
	case CtrlConn:
		protocol, next, err := getString(body)
		if err != nil {
			return nil, err
		}

		if next[0] != byte(V5) {
			return nil, ErrDecodeNoneV5Packet
		}

		if len(next) < 5 {
			return nil, ErrDecodeBadPacket
		}

		hasUsername := next[1]&0x80 == 0x80
		hasPassword := next[1]&0x40 == 0x40
		pkt := &ConnPacket{
			ProtoName:    protocol,
			CleanSession: next[1]&0x02 == 0x02,
			IsWill:       next[1]&0x04 == 0x04,
			WillQos:      next[1] & 0x18 >> 3,
			WillRetain:   next[1]&0x20 == 0x20,
			Keepalive:    getUint16(next[2:4]),
			Props:        &ConnProps{},
		}
		pkt.ProtoVersion = ProtoVersion(body[0])

		// read properties
		var props map[byte][]byte
		props, next = getRawProps(next[4:])
		pkt.Props.setProps(props)

		if pkt.ClientID, next, err = getString(next); err != nil {
			return nil, err
		}

		if pkt.IsWill {
			pkt.WillTopic, next, err = getString(next)
			pkt.WillMessage, next, err = getBinaryData(next)
		}

		if hasUsername {
			pkt.Username, next, err = getString(next)
		}

		if hasPassword {
			pkt.Password, _, err = getString(next)
		}

		if err != nil {
			return nil, err
		}

		return pkt, nil
	case CtrlConnAck:
		pkt := &ConnAckPacket{
			Present: body[0]&0x01 == 0x01,
			Code:    body[1],
			Props:   &ConnAckProps{},
		}

		props, _ := getRawProps(body[2:])
		pkt.Props.setProps(props)

		return pkt, nil
	case CtrlPublish:
		topicName, next, err := getString(body)
		if err != nil {
			return nil, err
		}

		if len(next) < 2 {
			return nil, ErrDecodeBadPacket
		}

		pub := &PublishPacket{
			IsDup:     header&0x08 == 0x08,
			Qos:       header & 0x06 >> 1,
			IsRetain:  header&0x01 == 1,
			TopicName: topicName,
			Props:     &PublishProps{},
		}

		if pub.Qos > Qos0 {
			pub.PacketID = getUint16(next)
			next = next[2:]
		}

		var props map[byte][]byte
		props, next = getRawProps(body[3:])
		pub.Props.setProps(props)

		pub.Payload = next
		return pub, nil
	case CtrlPubAck:
		pkt := &PubAckPacket{
			PacketID: getUint16(body),
			Code:     body[2],
			Props:    &PubAckProps{},
		}

		props, _ := getRawProps(body[3:])
		pkt.Props.setProps(props)

		return pkt, nil
	case CtrlPubRecv:
		pkt := &PubRecvPacket{
			PacketID: getUint16(body),
			Code:     body[2],
			Props:    &PubRecvProps{},
		}
		props, _ := getRawProps(body[3:])
		pkt.Props.setProps(props)

		return pkt, nil
	case CtrlPubRel:
		pkt := &PubRelPacket{
			PacketID: getUint16(body),
			Code:     body[2],
			Props:    &PubRelProps{},
		}
		props, _ := getRawProps(body[3:])
		pkt.Props.setProps(props)

		return pkt, nil
	case CtrlPubComp:
		pkt := &PubCompPacket{
			PacketID: getUint16(body),
			Code:     body[2],
			Props:    &PubCompProps{},
		}

		props, _ := getRawProps(body[3:])
		pkt.Props.setProps(props)

		return pkt, nil
	case CtrlSubscribe:
		pkt := &SubscribePacket{
			PacketID: getUint16(body),
			Props:    &SubscribeProps{},
		}

		props, next := getRawProps(body[2:])
		pkt.Props.setProps(props)

		topics := make([]*Topic, 0)
		for len(next) > 0 {
			var name string
			if name, next, err = getString(next); err != nil {
				return nil, err
			}

			if len(next) < 1 {
				return nil, ErrDecodeBadPacket
			}

			topics = append(topics, &Topic{Name: name, Qos: next[0]})
			next = next[1:]
		}
		pkt.Topics = topics
		return pkt, nil
	case CtrlSubAck:
		pkt := &SubAckPacket{
			PacketID: getUint16(body),
			Props:    &SubAckProps{},
		}

		props, next := getRawProps(body[2:])
		pkt.Props.setProps(props)

		pkt.Codes = make([]byte, 0)
		for i := 0; i < len(next); i++ {
			pkt.Codes = append(pkt.Codes, next[i])
		}
		return pkt, nil
	case CtrlUnSub:
		pkt := &UnSubPacket{
			PacketID: getUint16(body),
			Props:    &UnSubProps{},
		}

		props, next := getRawProps(body[2:])
		pkt.Props.setProps(props)

		pkt.TopicNames = make([]string, 0)
		for len(next) > 0 {
			var name string
			name, next, err = getString(next)
			if err != nil {
				return nil, err
			}
			pkt.TopicNames = append(pkt.TopicNames, name)
		}
		return pkt, nil
	case CtrlUnSubAck:
		pkt := &UnSubAckPacket{
			PacketID: getUint16(body),
			Props:    &UnSubAckProps{},
		}

		props, _ := getRawProps(body[2:])
		pkt.Props.setProps(props)

		return pkt, nil
	case CtrlDisConn:
		pkt := &DisConnPacket{
			Code:  body[0],
			Props: &DisConnProps{},
		}

		props, _ := getRawProps(body[1:])
		pkt.Props.setProps(props)

		return pkt, nil
	case CtrlAuth:
		pkt := &AuthPacket{
			Code:  body[0],
			Props: &AuthProps{},
		}

		props, _ := getRawProps(body[1:])
		pkt.Props.setProps(props)

		return pkt, nil
	default:
		return nil, ErrDecodeBadPacket
	}
}
