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

import (
	"testing"
)

func TestPublishPacket_Bytes(t *testing.T) {
	for i, p := range testPubMsgs {
		testV311Bytes(p, testPubMsgBytes[i], t)
	}
}

func TestPubAckPacket_Bytes(t *testing.T) {
	testV311Bytes(testPubAckMsg, testPubAckMsgBytes, t)
}

func TestPubRecvPacket_Bytes(t *testing.T) {
	testV311Bytes(testPubRecvMsg, testPubRecvMsgBytes, t)
}

func TestPubRelPacket_Bytes(t *testing.T) {
	testV311Bytes(testPubRelMsg, testPubRelMsgBytes, t)
}

func TestPubCompPacket_Bytes(t *testing.T) {
	testV311Bytes(testPubCompMsg, testPubCompMsgBytes, t)
}
