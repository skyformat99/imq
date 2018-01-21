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

// PublishPacket is sent from a Client to a Server or from Server to a Client
// to transport an Application Message.
type PublishPacket struct {
	basePacket
	IsDup     bool
	Qos       QosLevel
	IsRetain  bool
	TopicName string
	Payload   []byte
	PacketID  uint16
	Props     *PublishProps
}

// Type of PublishPacket is CtrlPublish
func (p *PublishPacket) Type() CtrlType {
	return CtrlPublish
}

func (p *PublishPacket) Bytes() []byte {
	return p.bytes(p)
}

func (p *PublishPacket) payload() []byte {
	data := encodeDataWithLen([]byte(p.TopicName))
	if p.Qos > Qos0 {
		data = append(data, byte(p.PacketID>>8), byte(p.PacketID))
	}
	return append(data, p.Payload...)
}

// PublishProps properties for PublishPacket
type PublishProps struct {
	// PayloadFormat Indicator
	// 0, Indicates that the Payload is unspecified bytes, which is equivalent to not sending a Payload Format Indicator
	// 1, Indicates that the Payload is UTF-8 Encoded Character Data. The UTF-8 data in the Payload
	PayloadFormat byte // required in server

	// MessageExpiryInterval
	// Lifetime of the Application Message in seconds
	// If absent, the Application Message does not expire
	MessageExpiryInterval uint32

	// A Topic Alias is an integer value that is used to identify the Topic
	// instead of using the Topic Name.
	//
	// This reduces the size of the PUBLISH packet, and is useful when the
	// Topic Names are long and the same Topic Names are used repetitively
	// within a Network Connection
	TopicAlias uint16

	// RespTopic Used as the Topic Name for a response message
	RespTopic string

	// CorrelationData used by the sender of the Request Message to identify which request the Response Message is for when it is received
	CorrelationData []byte

	// User defined Properties
	UserProps UserProperties

	// SubIDs the identifier of the subscription (always no 0)
	//
	// Multiple Subscription Identifiers will be included if the publication
	// is the result of a match to more than one subscription, in this case
	// their order is not significant
	SubIDs []int

	// ContentType describe the content of the Application Message
	ContentType string
}

func (p *PublishProps) props() []byte {
	if p == nil {
		return nil
	}

	result := make([]byte, 2)
	result[0] = propKeyPayloadFormatIndicator
	result[1] = p.PayloadFormat

	if p.MessageExpiryInterval != 0 {
		data := []byte{propKeyMessageExpiryInterval, 0, 0, 0, 0}
		putUint32(data[1:], p.MessageExpiryInterval)
		result = append(result, data...)
	}

	if p.TopicAlias != 0 {
		data := []byte{propKeyTopicAlias, 0, 0, 0, 0}
		putUint16(data[1:], p.TopicAlias)
		result = append(result, data...)
	}

	if p.RespTopic != "" {
		result = append(result, propKeyRespTopic)
		result = append(result, encodeDataWithLen([]byte(p.RespTopic))...)
	}

	if p.CorrelationData != nil {
		result = append(result, propKeyCorrelationData)
		result = append(result, encodeDataWithLen(p.CorrelationData)...)
	}

	if p.UserProps != nil {
		p.UserProps.encodeTo(result)
	}

	if p.SubIDs != nil {
		buf := &bytes.Buffer{}
		for _, v := range p.SubIDs {
			result = append(result, propKeySubID)
			writeVarInt(v, buf)
			result = append(result, buf.Bytes()...)
			buf.Reset()
		}
	}

	if p.ContentType != "" {
		result = append(result, propKeyContentType)
		result = append(result, encodeDataWithLen([]byte(p.ContentType))...)
	}

	return result
}

func (p *PublishProps) setProps(props map[byte][]byte) {
	if p == nil || props == nil {
		return
	}

	if v, ok := props[propKeyPayloadFormatIndicator]; ok && len(v) == 1 {
		p.PayloadFormat = v[0]
	}

	if v, ok := props[propKeyMessageExpiryInterval]; ok {
		p.MessageExpiryInterval = getUint32(v)
	}

	if v, ok := props[propKeyTopicAlias]; ok {
		p.TopicAlias = getUint16(v)
	}

	if v, ok := props[propKeyRespTopic]; ok {
		p.RespTopic, _, _ = getString(v)
	}

	if v, ok := props[propKeyCorrelationData]; ok {
		p.CorrelationData, _, _ = getBinaryData(v)
	}

	if v, ok := props[propKeyUserProps]; ok {
		p.UserProps = getUserProps(v)
	}

	if v, ok := props[propKeySubID]; ok {
		p.SubIDs = make([]int, 0)
		for i := 0; i < len(v); {
			d, n := getRemainLength(bytes.NewReader(v[i:]))
			p.SubIDs = append(p.SubIDs, d)
			i += n
		}
	}

	if v, ok := props[propKeyContentType]; ok {
		p.ContentType, _, _ = getString(v)
	}

}

// PubAckPacket is the response to a PublishPacket with QoS level 1.
type PubAckPacket struct {
	basePacket
	PacketID uint16
	Code     byte
	Props    *PubAckProps
}

// Type of PubAckPacket is CtrlPubAck
func (p *PubAckPacket) Type() CtrlType {
	return CtrlPubAck
}

func (p *PubAckPacket) Bytes() []byte {
	return p.bytes(p)
}

// PubAckProps properties for PubAckPacket
type PubAckProps struct {
	// Human readable string designed for diagnostics
	Reason string

	// UserProps User defined Properties
	UserProps UserProperties
}

func (p *PubAckProps) props() []byte {
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

func (p *PubAckProps) setProps(props map[byte][]byte) {
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

// PubRecvPacket is the response to a PublishPacket with QoS 2.
// It is the second packet of the QoS 2 protocol exchange.
type PubRecvPacket struct {
	basePacket
	PacketID uint16
	Code     byte
	Props    *PubRecvProps
}

// Type of PubRecvPacket is CtrlPubRecv
func (p *PubRecvPacket) Type() CtrlType {
	return CtrlPubRecv
}

func (p *PubRecvPacket) Bytes() []byte {
	return p.bytes(p)
}

// PubRecvProps properties for PubRecvPacket
type PubRecvProps struct {
	// Human readable string designed for diagnostics
	Reason string

	// UserProps User defined Properties
	UserProps UserProperties
}

func (p *PubRecvProps) props() []byte {
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

func (p *PubRecvProps) setProps(props map[byte][]byte) {
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

// PubRelPacket is the response to a PubRecvPacket.
// It is the third packet of the QoS 2 protocol exchange.
type PubRelPacket struct {
	basePacket
	PacketID uint16
	Code     byte
	Props    *PubRelProps
}

// Type of PubRelPacket is CtrlPubRel
func (p *PubRelPacket) Type() CtrlType {
	return CtrlPubRel
}

func (p *PubRelPacket) Bytes() []byte {
	return p.bytes(p)
}

// PubRelProps properties for PubRelPacket
type PubRelProps struct {
	// Human readable string designed for diagnostics
	Reason string

	// UserProps User defined Properties
	UserProps UserProperties
}

func (p *PubRelProps) props() []byte {
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

func (p *PubRelProps) setProps(props map[byte][]byte) {
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

// PubCompPacket is the response to a PubRelPacket.
// It is the fourth and final packet of the QoS 892 2 protocol exchange. 893
type PubCompPacket struct {
	basePacket
	PacketID uint16
	Code     byte
	Props    *PubCompProps
}

// Type of PubCompPacket is CtrlPubComp
func (p *PubCompPacket) Type() CtrlType {
	return CtrlPubComp
}

func (p *PubCompPacket) Bytes() []byte {
	return p.bytes(p)
}

// PubCompProps properties for PubCompPacket
type PubCompProps struct {
	// Human readable string designed for diagnostics
	Reason string

	// UserProps User defined Properties
	UserProps UserProperties
}

func (p *PubCompProps) props() []byte {
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

func (p *PubCompProps) setProps(props map[byte][]byte) {
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
