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

var (
	// PingReqPacket is the final instance of pingReqPacket
	PingReqPacket = &pingReqPacket{}
	// PingRespPacket is the final instance of pingRespPacket
	PingRespPacket = &pingRespPacket{}
)

// pingReqPacket is sent from a Client to the Server.
//
// It can be used to:
// 		1. Indicate to the Server that the Client is alive in the absence of any other Control Packets being sent from the Client to the Server.
// 		2. Request that the Server responds to confirm that it is alive.
// 		3. Exercise the network to indicate that the Network Connection is active.
//
// This Packet is used in Keep Alive processing
type pingReqPacket struct {
	basePacket
}

func (p *pingReqPacket) Type() CtrlType {
	return CtrlPingReq
}

func (p *pingReqPacket) Bytes() []byte {
	return p.bytes(p)
}

// pingRespPacket is sent by the Server to the Client in response to
// a pingReqPacket. It indicates that the Server is alive.
type pingRespPacket struct {
	basePacket
}

func (p *pingRespPacket) Type() CtrlType {
	return CtrlPingResp
}

func (p *pingRespPacket) Bytes() []byte {
	return p.bytes(p)
}
