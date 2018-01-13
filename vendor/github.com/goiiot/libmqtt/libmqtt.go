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

// Packet is MQTT control packet
type Packet interface {
	// Type return the packet type
	Type() CtrlType
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
type CtrlType = byte

const (
	// CtrlConn Connect
	CtrlConn CtrlType = 1
	// CtrlConnAck Connect Ack
	CtrlConnAck CtrlType = 2
	// CtrlPublish Publish
	CtrlPublish CtrlType = 3
	// CtrlPubAck Publish Ack
	CtrlPubAck CtrlType = 4
	// CtrlPubRecv Publish Received
	CtrlPubRecv CtrlType = 5
	// CtrlPubRel Publish Release
	CtrlPubRel CtrlType = 6
	// CtrlPubComp Publish Complete
	CtrlPubComp CtrlType = 7
	// CtrlSubscribe Subscribe
	CtrlSubscribe CtrlType = 8
	// CtrlSubAck Subscribe Ack
	CtrlSubAck CtrlType = 9
	// CtrlUnSub UnSubscribe
	CtrlUnSub CtrlType = 10
	// CtrlUnSubAck UnSubscribe Ack
	CtrlUnSubAck CtrlType = 11
	// CtrlPingReq Ping Request
	CtrlPingReq CtrlType = 12
	// CtrlPingResp Ping Response
	CtrlPingResp CtrlType = 13
	// CtrlDisConn Disconnect
	CtrlDisConn CtrlType = 14
	// CtrlAuth Authentication (since MQTT 5.0)
	CtrlAuth CtrlType = 15
)

// ProtocolVersion MQTT Protocol version
type ProtocolVersion = byte

const (
	// V311 means MQTT 3.1.1
	V311 ProtocolVersion = 4
	// V5 means MQTT 5
	V5 ProtocolVersion = 5
)

// QosLevel is either 0, 1, 2
type QosLevel = byte

const (
	// Qos0 = 0
	Qos0 QosLevel = 0x00
	// Qos1 = 1
	Qos1 QosLevel = 0x01
	// Qos2 = 2
	Qos2 QosLevel = 0x02
)

var (
	mqtt = []byte("MQTT")
)

// ConnAckCode is connection response code from server
type ConnAckCode = byte

const (
	// ConnAccepted client accepted by server
	ConnAccepted ConnAckCode = 0
	// ConnBadProtocol Protocol not supported
	ConnBadProtocol ConnAckCode = 1
	// ConnIDRejected Connection Id not valid
	ConnIDRejected ConnAckCode = 2
	// ConnServerUnavailable Server error
	ConnServerUnavailable ConnAckCode = 3
	// ConnBadIdentity Identity failed
	ConnBadIdentity ConnAckCode = 4
	// ConnAuthFail Auth failed
	ConnAuthFail ConnAckCode = 5
)

// SubAckCode is returned by server in SubAckPacket
type SubAckCode = byte

const (
	// SubOkMaxQos0 QoS 0 is used by server
	SubOkMaxQos0 SubAckCode = 0
	// SubOkMaxQos1 QoS 1 is used by server
	SubOkMaxQos1 SubAckCode = 1
	// SubOkMaxQos2 QoS 2 is used by server
	SubOkMaxQos2 SubAckCode = 2
	// SubFail means that subscription is not successful
	SubFail SubAckCode = 0x80
)

// property identifiers

const (
	// PayloadFormatIndicator is
	//
	// Property type: byte
	// Packet: Will, Publish
	payloadFormatIndicator = 1

	// MessageExpiryInterval is
	//
	// Property type: 4 bytes int
	// Packet: Will, Publish
	messageExpiryInterval = 2

	// ContentType is
	//
	// Property type: utf-8 encoded string
	// Packet: Will, Publish
	contentType = 3

	// ResponseTopic is
	//
	// Property type: utf-8 encoded string
	// Packet: Will, Publish
	responseTopic = 8

	// CorrelationData is
	//
	// Property type: binary data
	// Packet: Will, Publish
	correlationData = 9

	// SubscriptionIdentifier is
	//
	// Property type: variable bytes int
	// Packet: Publish, Subscribe
	subscriptionIdentifier = 11

	// SessionExpiryInterval is
	//
	// Property type: 4 bytes int
	// Packet: Connect, ConnAck, DisConn
	sessionExpiryInterval = 17

	// AssignedClientIdentifier is
	//
	// Property type: utf-8 encoded string
	// Packet: ConnAck
	assignedClientIdentifier = 18

	// ServerKeepAlive is
	//
	// Property type: int (2 bytes)
	// Packet: ConnAck
	serverKeepAlive = 19

	// AuthenticationMethod is
	//
	// Property type: utf-8
	// Packet: Connect, ConnAck, Auth
	authenticationMethod = 21

	// AuthenticationData is
	//
	// Property type: binary data
	// Packet: Connect, ConnAck, Auth
	authenticationData = 22

	// RequestProblemInfo is
	//
	// Property type: byte
	// Packet: Connect
	requestProblemInfo = 23

	// WillDelayInterval is
	//
	// Property type: int (4 bytes)
	// Packet: Will
	willDelayInterval = 24

	// RequestResponseInfo is
	//
	// Property type: byte
	// Packet: Connect
	requestResponseInfo = 25

	// ResponseInfo is
	//
	// Property type: utf-8
	// Packet: ConnAck
	responseInfo = 26

	// ServerReference is
	//
	// Property type: utf-8 encoded string
	// Packet: ConnAck, DisConn
	serverReference = 28

	// ReasonString is
	//
	// Property type: utf-8
	// Packet: ConnAck, PubAck, PubRecv, PubRel,
	// 		   PubComp, SubAck, UnSubAck, DisConn,
	// 		   Auth
	reasonString = 31

	// ReceiveMax is
	//
	// Property type: int (2 bytes)
	// Packet: Connect, ConnAck
	receiveMax = 33

	// TopicAliasMax is
	//
	// Property type: int (2 bytes)
	// Packet: Connect, ConnAck
	topicAliasMax = 34

	// TopicAlias is
	//
	// Property type: int (2 bytes)
	// Packet: Publish
	topicAlias = 35

	// MaxQos is
	//
	// Property type: byte
	// Packet: ConnAck
	maxQos = 36

	// RetainAvail is
	//
	// Property type: byte
	// Packet: ConnAck
	retainAvail = 37

	// UserProperty is
	//
	// Property type: utf-8 string pair
	// Packet: Connect, ConnAck, Publish, Will,
	// 		   PubAck, PubRecv, PubRel, PubComp,
	// 		   Subscribe, SubAck, UnSub, UnSubAck,
	// 		   DisConn, Auth
	userProperty = 38

	// MaxPacketSize is
	//
	// Property type: int (4 bytes)
	// Packet: Connect, ConnAck
	maxPacketSize = 39

	// WildcardSubscriptionAvail is
	//
	// Property type: byte
	// Packet: ConnAck
	wildcardSubscriptionAvail = 40

	// SubscriptionIdentifierAvailable is
	//
	// Property type: byte
	// Packet: ConnAck
	subscriptionIdentifierAvailable = 41

	// SharedSubscriptionAvailable is
	//
	// Property type: byte
	// Packet: ConnAck
	sharedSubscriptionAvailable = 42
)

// reason code

const (
	// Packet: ConnAck, PubAck, PubRecv, PubRel,
	// 		   PubComp, UnSubAck, Auth
	success = 0

	// Packet: DisConn
	normalDisconnection = 0

	// Packet: SubAck
	grantedQos0 = 0

	// Packet: SubAck
	grantedQos1 = 1

	// Packet: SubAck
	grantedQos2 = 2

	// Packet: DisConn
	disconnectWithWillMessage = 4

	// Packet: PubAck, PubRecv
	noMatchingSubscribers = 16

	// Packet: UnSubAck
	noSubscriptionExisted = 17

	// Packet: Auth
	continueAuthentication = 24

	// Packet: Auth
	reAuthenticatie = 25

	// Packet: ConnAck, PubAck, PubRecv, SubAck,
	// 		   UnSubAck, DisConn
	UnspecifiedError = 128

	// Packet: ConnAck, DisConn
	malformedPacket = 129

	// Packet: ConnAck, DisConn
	protocolError = 130

	// Packet: ConnAck, PubAck, PubRecv, SubAck, UnSubAck, DisConn
	implementationSpecificError = 131

	// Packet: ConnAck
	unsupportedProtocolVersion = 132

	// Packet: ConnAck
	clientIdentifierNotValid = 133

	// Packet: ConnAck
	badUserNameOrPassword = 134

	// Packet: ConnAck, PubAck, PubRecv, SubAck, UnSubAck, DisConn
	notAuthorized = 135

	// Packet: ConnAck
	serverUnavailable = 136

	// Packet: ConnAck, DisConn
	serverBusy = 137

	// Packet: ConnAck
	banned = 138

	// Packet: DisConn
	serverShuttingDown = 139

	// Packet: ConnAck, DisConn
	badAuthenticationMethod = 140

	// Packet: DisConn
	keepaliveTimeout = 141

	// Packet: DisConn
	sessionTakenOver = 142

	// Packet: SubAck, UnSubAck, DisConn
	topicFilterInvalid = 143

	// Packet: ConnAck, PubAck, PubRecv, DisConn
	topicNameInvalid = 144

	// Packet: PubAck, PubRecv, PubAck, UnSubAck
	//
	// For Packet identifier in use code, the response to this is
	// either to try to fix the state, or to reset the Session
	// state by connecting using Clean Start set to 1, or to
	// decide if the Client or Server implementations are defective.
	packetIdentifierInUse = 145

	// Packet: PubRel, PubComp
	packetIdentifierNotFound = 146

	// Packet: DisConn
	receiveMaxExceeded = 147

	// Packet: DisConn
	topicAliasInvalid = 148

	// Packet: ConnAck, DisConn
	packetTooLarge = 149

	// Packet: DisConn
	messageRateTooHigh = 150

	// Packet: ConnAck, PubAck, PubRec, SubAck, DisConn
	quotaExceeded = 151

	// Packet: DisConn
	administrativeAction = 152

	// Packet: ConnAck, PubAck, PubRecv, DisConn
	payloadFormatInvalid = 153

	// Packet: ConnAck, DisConn
	retainNotSupported = 154

	// Packet: ConnAck, DisConn
	qosNoSupported = 155

	// Packet: ConnAck, DisConn
	useAnotherServer = 156

	// Packet: ConnAck, DisConn
	serverMoved = 157

	// Packet: SubAck, DisConn
	sharedSubscriptionNotSupported = 158

	// Packet: ConnAck, DisConn
	connectionRateExceeded = 159

	// Packet: DisConn
	maxConnectTime = 160

	// Packet: SubAck, DisConn
	subscriptionIdentifiersNotSupported = 161

	// Packet: SubAck, DisConn
	wildcardSubscriptionNotSupported = 162
)
