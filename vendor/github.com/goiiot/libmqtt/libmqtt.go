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
)

// UserProperties contains user defined properties
type UserProperties map[string][]string

func (u UserProperties) encodeTo(result []byte) {
	for k, v := range u {
		for _, val := range v {
			result = append(result, propKeyUserProps)
			result = append(result, encodeDataWithLen([]byte(k))...)
			result = append(result, encodeDataWithLen([]byte(val))...)
		}
	}
}

// Packet is MQTT control packet
type Packet interface {
	// Type return the packet type
	Type() CtrlType

	Bytes() []byte

	Version() ProtoVersion
}

type basePacket struct {
	ProtoVersion ProtoVersion
}

func (basePacket) bytes(p Packet) []byte {
	if p == nil {
		return nil
	}

	buf := &bytes.Buffer{}
	if Encode(p, buf) != nil {
		return nil
	}

	return buf.Bytes()
}

func (b basePacket) Version() ProtoVersion {
	if b.ProtoVersion != 0 {
		return b.ProtoVersion
	}

	return V311
}

// Topic for both topic name and topic qos
type Topic struct {
	Name string
	Qos  QosLevel
}

func (t *Topic) String() string {
	return t.Name
}

const (
	maxMsgSize = 0xffffff7f
)

// CtrlType is MQTT Control packet type
type CtrlType byte

const (
	CtrlConn      CtrlType = 1  // CtrlConn Connect
	CtrlConnAck   CtrlType = 2  // CtrlConnAck Connect Ack
	CtrlPublish   CtrlType = 3  // CtrlPublish Publish
	CtrlPubAck    CtrlType = 4  // CtrlPubAck Publish Ack
	CtrlPubRecv   CtrlType = 5  // CtrlPubRecv Publish Received
	CtrlPubRel    CtrlType = 6  // CtrlPubRel Publish Release
	CtrlPubComp   CtrlType = 7  // CtrlPubComp Publish Complete
	CtrlSubscribe CtrlType = 8  // CtrlSubscribe Subscribe
	CtrlSubAck    CtrlType = 9  // CtrlSubAck Subscribe Ack
	CtrlUnSub     CtrlType = 10 // CtrlUnSub UnSubscribe
	CtrlUnSubAck  CtrlType = 11 // CtrlUnSubAck UnSubscribe Ack
	CtrlPingReq   CtrlType = 12 // CtrlPingReq Ping Request
	CtrlPingResp  CtrlType = 13 // CtrlPingResp Ping Response
	CtrlDisConn   CtrlType = 14 // CtrlDisConn Disconnect
	CtrlAuth      CtrlType = 15 // CtrlAuth Authentication (since MQTT 5.0)
)

// ProtoVersion MQTT Protocol ProtoVersion
type ProtoVersion byte

const (
	V311 ProtoVersion = 4 // V311 means MQTT 3.1.1
	V5   ProtoVersion = 5 // V5 means MQTT 5
)

// QosLevel is either 0, 1, 2
type QosLevel = byte

const (
	Qos0 QosLevel = 0x00 // Qos0 = 0
	Qos1 QosLevel = 0x01 // Qos1 = 1
	Qos2 QosLevel = 0x02 // Qos2 = 2
)

var (
	mqtt = []byte{0x00, 0x04, 'M', 'Q', 'T', 'T'}
)

const (
	SubOkMaxQos0 = 0    // SubOkMaxQos0 QoS 0 is used by server
	SubOkMaxQos1 = 1    // SubOkMaxQos1 QoS 1 is used by server
	SubOkMaxQos2 = 2    // SubOkMaxQos2 QoS 2 is used by server
	SubFail      = 0x80 // SubFail means that subscription is not successful
)

// reason code

const (
	CodeSuccess                             = 0   // Packet: ConnAck, PubAck, PubRecv, PubRel, PubComp, UnSubAck, Auth
	CodeNormalDisconn                       = 0   // Packet: DisConn
	CodeGrantedQos0                         = 0   // Packet: SubAck
	CodeGrantedQos1                         = 1   // Packet: SubAck
	CodeGrantedQos2                         = 2   // Packet: SubAck
	CodeDisconnWithWill                     = 4   // Packet: DisConn
	CodeNoMatchingSubscribers               = 16  // Packet: PubAck, PubRecv
	CodeNoSubscriptionExisted               = 17  // Packet: UnSubAck
	CodeContinueAuth                        = 24  // Packet: Auth
	CodeReAuth                              = 25  // Packet: Auth
	CodeUnspecifiedError                    = 128 // Packet: ConnAck, PubAck, PubRecv, SubAck, UnSubAck, DisConn
	CodeMalformedPacket                     = 129 // Packet: ConnAck, DisConn
	CodeProtoError                          = 130 // Packet: ConnAck, DisConn
	CodeImplementationSpecificError         = 131 // Packet: ConnAck, PubAck, PubRecv, SubAck, UnSubAck, DisConn
	CodeUnsupportedProtoVersion             = 132 // Packet: ConnAck
	CodeClientIdNotValid                    = 133 // Packet: ConnAck
	CodeBadUserPass                         = 134 // Packet: ConnAck
	CodeNotAuthorized                       = 135 // Packet: ConnAck, PubAck, PubRecv, SubAck, UnSubAck, DisConn
	CodeServerUnavail                       = 136 // Packet: ConnAck
	CodeServerBusy                          = 137 // Packet: ConnAck, DisConn
	CodeBanned                              = 138 // Packet: ConnAck
	CodeServerShuttingDown                  = 139 // Packet: DisConn
	CodeBadAuthenticationMethod             = 140 // Packet: ConnAck, DisConn
	CodeKeepaliveTimeout                    = 141 // Packet: DisConn
	CodeSessionTakenOver                    = 142 // Packet: DisConn
	CodeTopicFilterInvalid                  = 143 // Packet: SubAck, UnSubAck, DisConn
	CodeTopicNameInvalid                    = 144 // Packet: ConnAck, PubAck, PubRecv, DisConn
	CodePacketIdentifierInUse               = 145 // Packet: PubAck, PubRecv, PubAck, UnSubAck
	CodePacketIdentifierNotFound            = 146 // Packet: PubRel, PubComp
	CodeReceiveMaxExceeded                  = 147 // Packet: DisConn
	CodeTopicAliasInvalid                   = 148 // Packet: DisConn
	CodePacketTooLarge                      = 149 // Packet: ConnAck, DisConn
	CodeMessageRateTooHigh                  = 150 // Packet: DisConn
	CodeQuotaExceeded                       = 151 // Packet: ConnAck, PubAck, PubRec, SubAck, DisConn
	CodeAdministrativeAction                = 152 // Packet: DisConn
	CodePayloadFormatInvalid                = 153 // Packet: ConnAck, PubAck, PubRecv, DisConn
	CodeRetainNotSupported                  = 154 // Packet: ConnAck, DisConn
	CodeQosNoSupported                      = 155 // Packet: ConnAck, DisConn
	CodeUseAnotherServer                    = 156 // Packet: ConnAck, DisConn
	CodeServerMoved                         = 157 // Packet: ConnAck, DisConn
	CodeSharedSubscriptionNotSupported      = 158 // Packet: SubAck, DisConn
	CodeConnectionRateExceeded              = 159 // Packet: ConnAck, DisConn
	CodeMaxConnectTime                      = 160 // Packet: DisConn
	CodeSubscriptionIdentifiersNotSupported = 161 // Packet: SubAck, DisConn
	CodeWildcardSubscriptionNotSupported    = 162 // Packet: SubAck, DisConn
)

// property identifiers

const (
	// PayloadFormatIndicator is
	//
	// Property type: byte
	// Packet: Will, Publish
	propKeyPayloadFormatIndicator = 1

	// MessageExpiryInterval is
	//
	// Property type: 4 bytes int
	// Packet: Will, Publish
	propKeyMessageExpiryInterval = 2

	// ContentType is
	//
	// Property type: utf-8 encoded string
	// Packet: Will, Publish
	propKeyContentType = 3

	// ResponseTopic is
	//
	// Property type: utf-8 encoded string
	// Packet: Will, Publish
	propKeyRespTopic = 8

	// CorrelationData is
	//
	// Property type: binary data
	// Packet: Will, Publish
	propKeyCorrelationData = 9

	// SubscriptionIdentifier is
	//
	// Property type: variable bytes int
	// Packet: Publish, Subscribe
	propKeySubID = 11

	// SessionExpiryInterval is
	//
	// Property type: 4 bytes int
	// Packet: Connect, ConnAck, DisConn
	propKeySessionExpiryInterval = 17

	// AssignedClientIdentifier is
	//
	// Property type: utf-8 encoded string
	// Packet: ConnAck
	propKeyAssignedClientID = 18

	// ServerKeepAlive is
	//
	// Property type: int (2 bytes)
	// Packet: ConnAck
	propKeyServerKeepalive = 19

	// AuthenticationMethod is
	//
	// Property type: utf-8
	// Packet: Connect, ConnAck, Auth
	propKeyAuthMethod = 21

	// AuthenticationData is
	//
	// Property type: binary data
	// Packet: Connect, ConnAck, Auth
	propKeyAuthData = 22

	// RequestProblemInfo is
	//
	// Property type: byte
	// Packet: Connect
	propKeyReqProblemInfo = 23

	// WillDelayInterval is
	//
	// Property type: int (4 bytes)
	// Packet: Will
	propKeyWillDelayInterval = 24

	// RequestResponseInfo is
	//
	// Property type: byte
	// Packet: Connect
	propKeyReqRespInfo = 25

	// ResponseInfo is
	//
	// Property type: utf-8
	// Packet: ConnAck
	propKeyRespInfo = 26

	// ServerReference is
	//
	// Property type: utf-8 encoded string
	// Packet: ConnAck, DisConn
	propKeyServerRef = 28

	// ReasonString is
	//
	// Property type: utf-8
	// Packet: ConnAck, PubAck, PubRecv, PubRel,
	// 		   PubComp, SubAck, UnSubAck, DisConn,
	// 		   Auth
	propKeyReasonString = 31

	// ReceiveMax is
	//
	// Property type: int (2 bytes)
	// Packet: Connect, ConnAck
	propKeyMaxRecv = 33

	// MaxTopicAlias is
	//
	// Property type: int (2 bytes)
	// Packet: Connect, ConnAck
	propKeyMaxTopicAlias = 34

	// TopicAlias is
	//
	// Property type: int (2 bytes)
	// Packet: Publish
	propKeyTopicAlias = 35

	// MaxQos is
	//
	// Property type: byte
	// Packet: ConnAck
	propKeyMaxQos = 36

	// RetainAvail is
	//
	// Property type: byte
	// Packet: ConnAck
	propKeyRetainAvail = 37

	// UserProperties is
	//
	// Property type: utf-8 string pair
	// Packet: Connect, ConnAck, Publish, Will,
	// 		   PubAck, PubRecv, PubRel, PubComp,
	// 		   Subscribe, SubAck, UnSub, UnSubAck,
	// 		   DisConn, Auth
	propKeyUserProps = 38

	// MaxPacketSize is
	//
	// Property type: int (4 bytes)
	// Packet: Connect, ConnAck
	propKeyMaxPacketSize = 39

	// WildcardSubscriptionAvail is
	//
	// Property type: byte
	// Packet: ConnAck
	propKeyWildcardSubAvail = 40

	// SubscriptionIdentifierAvailable is
	//
	// Property type: byte
	// Packet: ConnAck
	propKeySubIDAvail = 41

	// SharedSubscriptionAvailable is
	//
	// Property type: byte
	// Packet: ConnAck
	propKeySharedSubAvail = 42
)
