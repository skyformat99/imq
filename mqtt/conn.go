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
	"bufio"
	"context"
	"net"
	"net/http"
	"time"

	mqtt "github.com/goiiot/libmqtt"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

var upGrader = &websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	Subprotocols:    []string{"mqtt", "mqtts"},
}

func handleWS(w http.ResponseWriter, r *http.Request) {
	conn, err := upGrader.Upgrade(w, r, make(http.Header))
	if err != nil {
		log.Error("establish ws connection fail", zap.Error(err))
		return
	}

	defer conn.Close()

	handleConn(conn.UnderlyingConn())
}

func handleConn(conn net.Conn) {
	var (
		pkt mqtt.Packet
		err error
	)
	connRW := bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn))

	// TODO: determine version via initial connect packet
	switch conf.version {
	case mqtt.V311:
		pkt, err = mqtt.Decode(mqtt.V311, connRW)
		if err != nil {
			log.Error("connection error", zap.Error(err))
			return
		}
	case mqtt.V5:
		pkt, err = mqtt.Decode(mqtt.V5, connRW)
		if err != nil {
			log.Error("connection error", zap.Error(err))
			return
		}
	}

	connPkt, ok := pkt.(*mqtt.ConnPacket)
	if !ok {
		log.Error("Expect connect packet")
		conn.Close()
		return
	}

	c := newConn(mqtt.V311, conn, connRW, connPkt)
	// TODO store client connection

	go c.handleConnRecv()
	go c.handleConnSend()
}

func newConn(version mqtt.ProtoVersion, conn net.Conn, connRW *bufio.ReadWriter, connPkt *mqtt.ConnPacket) *connImpl {
	ctx, cancel := context.WithCancel(context.TODO())
	c := &connImpl{
		conn:    conn,
		connRW:  connRW,
		connPkt: connPkt,
		version: version,
		recvC:   make(chan mqtt.Packet),
		sendC:   make(chan mqtt.Packet),
		pubC:    make(chan *mqtt.PublishPacket),
		ctx:     ctx,
		exit:    cancel,
	}
	return c
}

type connImpl struct {
	conn    net.Conn          // actual connection with client
	connRW  *bufio.ReadWriter // buffered connection
	version mqtt.ProtoVersion // mqtt version in use
	connPkt *mqtt.ConnPacket  // initial connect packet

	// channels for client server communication
	recvC chan mqtt.Packet         // server recv channel
	sendC chan mqtt.Packet         // server send channel
	keepC chan int                 // keepalive channel
	pubC  chan *mqtt.PublishPacket // server publish channel

	// context for exit client
	ctx  context.Context
	exit context.CancelFunc
}

func (c *connImpl) handleConnRecv() {
	// TODO: complete receive logic
	for {
		select {
		case <-c.ctx.Done():
		default:
			pkt, err := mqtt.Decode(c.version, c.connRW)
			if err != nil {
				return
			}

			// TODO: react to packet received
			switch pkt.(type) {
			}
		}
	}
}

func (c *connImpl) handleConnSend() {
	// TODO: complete send logic
	for {
		select {
		case <-c.ctx.Done():
		case <-c.sendC:
		case <-c.pubC:
		}
	}
}

func (c *connImpl) handleKeepalive() {
	interval := time.Duration(c.connPkt.Keepalive) * time.Second
	timer := time.NewTimer(interval)
	for {
		select {
		case <-c.ctx.Done():
		case <-timer.C:

		}
	}
}

func (c *connImpl) disconnect(reason byte) {
	// TODO: disconnect with reason code
}
