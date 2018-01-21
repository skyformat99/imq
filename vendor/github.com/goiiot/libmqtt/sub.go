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

import "bytes"

// SubscribePacket is sent from the Client to the Server
// to create one or more Subscriptions.
//
// Each Subscription registers a Client's interest in one or more TopicNames.
// The Server sends PublishPackets to the Client in order to forward
// Application Messages that were published to TopicNames that match these Subscriptions.
// The SubscribePacket also specifies (for each Subscription)
// the maximum QoS with which the Server can send Application Messages to the Client
type SubscribePacket struct {
	basePacket
	PacketID uint16
	Topics   []*Topic
	Props    *SubscribeProps
}

// Type of SubscribePacket is CtrlSubscribe
func (s *SubscribePacket) Type() CtrlType {
	return CtrlSubscribe
}

func (s *SubscribePacket) Bytes() []byte {
	return s.bytes(s)
}

func (s *SubscribePacket) payload() []byte {
	var result []byte
	if s.Topics != nil {
		for _, t := range s.Topics {
			result = append(result, encodeDataWithLen([]byte(t.Name))...)
			result = append(result, t.Qos)
		}
	}
	return result
}

// SubscribeProps properties for SubscribePacket
type SubscribeProps struct {
	// SubID identifier of the subscription
	SubID uint32
	// UserProps User defined Properties
	UserProps UserProperties
}

func (s *SubscribeProps) props() []byte {
	if s == nil {
		return nil
	}

	result := make([]byte, 0)

	if s.SubID != 0 {
		buf := &bytes.Buffer{}
		writeVarInt(int(s.SubID), buf)
		result = append(result, propKeySubID)
		result = append(result, buf.Bytes()...)
	}

	if s.UserProps != nil {
		s.UserProps.encodeTo(result)
	}
	return result
}

func (s *SubscribeProps) setProps(props map[byte][]byte) {
	if s == nil || props == nil {
		return
	}

	if v, ok := props[propKeySubID]; ok {
		id, _ := getRemainLength(bytes.NewReader(v))
		s.SubID = uint32(id)
	}

	if v, ok := props[propKeyUserProps]; ok {
		s.UserProps = getUserProps(v)
	}
}

// SubAckPacket is sent by the Server to the Client
// to confirm receipt and processing of a SubscribePacket.
//
// SubAckPacket contains a list of return codes,
// that specify the maximum QoS level that was granted in
// each Subscription that was requested by the SubscribePacket.
type SubAckPacket struct {
	basePacket
	PacketID uint16
	Codes    []byte
	Props    *SubAckProps
}

// Type of SubAckPacket is CtrlSubAck
func (s *SubAckPacket) Type() CtrlType {
	return CtrlSubAck
}

func (s *SubAckPacket) Bytes() []byte {
	return s.bytes(s)
}

func (s *SubAckPacket) payload() []byte {
	return s.Codes
}

// SubAckProps properties for SubAckPacket
type SubAckProps struct {
	// Human readable string designed for diagnostics
	Reason string

	// UserProps User defined Properties
	UserProps UserProperties
}

func (p *SubAckProps) props() []byte {
	if p == nil {
		return nil
	}

	result := make([]byte, 0)
	if p.Reason != "" {
		result = append(result, propKeyReasonString)
		result = append(result, encodeDataWithLen([]byte(p.Reason))...)
	}

	if p.UserProps != nil {
		p.UserProps.encodeTo(result)
	}
	return result
}

func (p *SubAckProps) setProps(props map[byte][]byte) {
	if p == nil || props == nil {
		return
	}

	if v, ok := props[propKeyReasonString]; ok {
		p.Reason, _, _ = getString(v)
	}

	if v, ok := props[propKeyUserProps]; ok {
		p.UserProps = getUserProps(v)
	}
}

// UnSubPacket is sent by the Client to the Server,
// to unsubscribe from topics.
type UnSubPacket struct {
	basePacket
	PacketID   uint16
	TopicNames []string
	Props      *UnSubProps
}

// Type of UnSubPacket is CtrlUnSub
func (s *UnSubPacket) Type() CtrlType {
	return CtrlUnSub
}
func (s *UnSubPacket) Bytes() []byte {
	return s.bytes(s)
}

func (s *UnSubPacket) payload() []byte {
	result := make([]byte, 0)
	if s.TopicNames != nil {
		for _, t := range s.TopicNames {
			result = append(result, encodeDataWithLen([]byte(t))...)
		}
	}
	return result
}

// UnSubProps properties for UnSubPacket
type UnSubProps struct {
	// UserProps User defined Properties
	UserProps UserProperties
}

func (p *UnSubProps) props() []byte {
	if p == nil {
		return nil
	}
	result := make([]byte, 0)
	if p.UserProps != nil {
		p.UserProps.encodeTo(result)
	}
	return result
}

func (p *UnSubProps) setProps(props map[byte][]byte) {
	if p == nil || props == nil {
		return
	}

	if v, ok := props[propKeyUserProps]; ok {
		p.UserProps = getUserProps(v)
	}
}

// UnSubAckPacket is sent by the Server to the Client to confirm
// receipt of an UnSubPacket
type UnSubAckPacket struct {
	basePacket
	PacketID uint16
	Props    *UnSubAckProps
}

// Type of UnSubAckPacket is CtrlUnSubAck
func (s *UnSubAckPacket) Type() CtrlType {
	return CtrlUnSubAck
}

func (s *UnSubAckPacket) Bytes() []byte {
	return s.bytes(s)
}

// UnSubAckProps properties for UnSubAckPacket
type UnSubAckProps struct {
	// Human readable string designed for diagnostics
	Reason string

	// UserProps User defined Properties
	UserProps UserProperties
}

func (p *UnSubAckProps) props() []byte {
	if p == nil {
		return nil
	}
	result := make([]byte, 0)
	if p.Reason != "" {
		result = append(result, propKeyReasonString)
		result = append(result, encodeDataWithLen([]byte(p.Reason))...)
	}

	if p.UserProps != nil {
		p.UserProps.encodeTo(result)
	}
	return result
}

func (p *UnSubAckProps) setProps(props map[byte][]byte) {
	if p == nil || props == nil {
		return
	}

	if v, ok := props[propKeyReasonString]; ok {
		p.Reason, _, _ = getString(v)
	}

	if v, ok := props[propKeyUserProps]; ok {
		p.UserProps = getUserProps(v)
	}
}
