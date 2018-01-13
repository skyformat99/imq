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

package mqtt

import (
	"context"
	"net"
	"sync"

	"github.com/goiiot/libmqtt"
	"go.uber.org/zap"
)

var (
	connections = &sync.Map{}
)

func tcpWorker(conn net.Conn) {
	switch conf.version {
	case libmqtt.V311:
	case libmqtt.V5:

	}
	pkt, err := libmqtt.DecodeOnePacket(libmqtt.V5, conn)
	if err != nil {
		log.Error("connection error", zap.Error(err))
		return
	}

	connPkt, ok := pkt.(*libmqtt.ConnPacket)
	if !ok {
		log.Error("expect connect packet")
		conn.Close()
		return
	}

	ctx, cancel := context.WithCancel(context.TODO())
	c := &connImpl{
		conn:    conn,
		connPkt: connPkt,
		ctx:     ctx,
		cancel:  cancel,
	}
	connections.Store("", c)
}

type connImpl struct {
	conn    net.Conn
	connPkt *libmqtt.ConnPacket
	ctx     context.Context
	cancel  context.CancelFunc
}

func (c *connImpl) handleConnRecv() {
	// TODO: complete receive logic
}

func (c *connImpl) handleConnSend() {
	// TODO: complete send logic
}
