package libmqtt

import (
	"bufio"
	"net"
	"time"
)

// clientConn is the wrapper of connection to server
// tend to actual packet send and receive
type clientConn struct {
	parent     *client       // client which created this connection
	name       string        // server addr info
	conn       net.Conn      // connection to server
	connW      *bufio.Writer // make buffered connection
	logicSendC chan Packet   // logic send channel
	netRecvC   chan Packet   // received packet from server
	keepaliveC chan int      // keepalive packet
}

// start mqtt logic
func (c *clientConn) logic() {
	defer func() {
		c.conn.Close()
		c.parent.workers.Done()
		c.parent.log.e("CONN exit logic for server =", c.name)
	}()

	// start keepalive if required
	if c.parent.options.keepalive > 0 {
		c.parent.workers.Add(1)
		go c.keepalive()
	}

	for {
		select {
		case <-c.parent.ctx.Done():
			return
		case pkt, more := <-c.netRecvC:
			if !more {
				return
			}

			switch pkt.(type) {
			case *SubAckPacket:
				p := pkt.(*SubAckPacket)
				c.parent.log.d("NET received SubAck, id =", p.PacketID)

				if originPkt, ok := c.parent.idGen.getExtra(p.PacketID); ok {
					switch originPkt.(type) {
					case *SubscribePacket:
						originSub := originPkt.(*SubscribePacket)
						N := len(p.Codes)
						for i, v := range originSub.Topics {
							if i < N {
								v.Qos = p.Codes[i]
							}
						}
						c.parent.log.d("CLIENT subscribed topics =", originSub.Topics)
						notifyMsg(c.parent.msgC, newSubMsg(originSub.Topics, nil))
						c.parent.idGen.free(p.PacketID)

						notifyMsg(c.parent.msgC, newPersistMsg(
							c.parent.persist.Delete(sendKey(p.PacketID))))
					}
				}
			case *UnSubAckPacket:
				p := pkt.(*UnSubAckPacket)
				c.parent.log.d("NET received UnSubAck, id =", p.PacketID)

				if originPkt, ok := c.parent.idGen.getExtra(p.PacketID); ok {
					switch originPkt.(type) {
					case *UnSubPacket:
						originUnSub := originPkt.(*UnSubPacket)
						c.parent.log.d("CLIENT unSubscribed topics", originUnSub.TopicNames)
						notifyMsg(c.parent.msgC, newUnSubMsg(originUnSub.TopicNames, nil))
						c.parent.idGen.free(p.PacketID)

						notifyMsg(c.parent.msgC, newPersistMsg(
							c.parent.persist.Delete(sendKey(p.PacketID))))
					}
				}
			case *PublishPacket:
				p := pkt.(*PublishPacket)
				c.parent.log.d("NET received publish, topic =", p.TopicName, "id =", p.PacketID, "QoS =", p.Qos)
				// received server publish, send to client
				c.parent.recvC <- p

				// tend to QoS
				switch p.Qos {
				case Qos1:
					c.parent.log.d("NET send PubAck for Publish, id =", p.PacketID)
					c.send(&PubAckPacket{PacketID: p.PacketID})

					notifyMsg(c.parent.msgC, newPersistMsg(
						c.parent.persist.Store(recvKey(p.PacketID), pkt)))
				case Qos2:
					c.parent.log.d("NET send PubRecv for Publish, id =", p.PacketID)
					c.send(&PubRecvPacket{PacketID: p.PacketID})

					notifyMsg(c.parent.msgC, newPersistMsg(
						c.parent.persist.Store(recvKey(p.PacketID), pkt)))
				}
			case *PubAckPacket:
				p := pkt.(*PubAckPacket)
				c.parent.log.d("NET received PubAck, id =", p.PacketID)

				if originPkt, ok := c.parent.idGen.getExtra(p.PacketID); ok {
					switch originPkt.(type) {
					case *PublishPacket:
						originPub := originPkt.(*PublishPacket)
						if originPub.Qos == Qos1 {
							c.parent.log.d("CLIENT published qos1 packet, topic =", originPub.TopicName)
							notifyMsg(c.parent.msgC, newPubMsg(originPub.TopicName, nil))
							c.parent.idGen.free(p.PacketID)

							notifyMsg(c.parent.msgC, newPersistMsg(
								c.parent.persist.Delete(sendKey(p.PacketID))))
						}
					}
				}
			case *PubRecvPacket:
				p := pkt.(*PubRecvPacket)
				c.parent.log.d("NET received PubRec, id =", p.PacketID)

				if originPkt, ok := c.parent.idGen.getExtra(p.PacketID); ok {
					switch originPkt.(type) {
					case *PublishPacket:
						originPub := originPkt.(*PublishPacket)
						if originPub.Qos == Qos2 {
							c.send(&PubRelPacket{PacketID: p.PacketID})
							c.parent.log.d("NET send PubRel, id =", p.PacketID)
						}
					}
				}
			case *PubRelPacket:
				p := pkt.(*PubRelPacket)
				c.parent.log.d("NET send PubRel, id =", p.PacketID)

				if originPkt, ok := c.parent.idGen.getExtra(p.PacketID); ok {
					switch originPkt.(type) {
					case *PublishPacket:
						originPub := originPkt.(*PublishPacket)
						if originPub.Qos == Qos2 {
							c.send(&PubCompPacket{PacketID: p.PacketID})
							c.parent.log.d("NET send PubComp, id =", p.PacketID)

							notifyMsg(c.parent.msgC, newPersistMsg(
								c.parent.persist.Store(recvKey(p.PacketID), pkt)))
						}
					}
				}
			case *PubCompPacket:
				p := pkt.(*PubCompPacket)
				c.parent.log.d("NET received PubComp, id =", p.PacketID)

				if originPkt, ok := c.parent.idGen.getExtra(p.PacketID); ok {
					switch originPkt.(type) {
					case *PublishPacket:
						originPub := originPkt.(*PublishPacket)
						if originPub.Qos == Qos2 {
							c.send(&PubRelPacket{PacketID: p.PacketID})
							c.parent.log.d("NET send PubRel, id =", p.PacketID)
							c.parent.log.d("CLIENT published qos2 packet, topic =", originPub.TopicName)
							notifyMsg(c.parent.msgC, newPubMsg(originPub.TopicName, nil))
							c.parent.idGen.free(p.PacketID)

							notifyMsg(c.parent.msgC, newPersistMsg(
								c.parent.persist.Delete(sendKey(p.PacketID))))
						}
					}
				}
			default:
				c.parent.log.d("NET received packet, type =", pkt.Type())
			}
		}
	}
}

// keepalive with server
func (c *clientConn) keepalive() {
	c.parent.log.d("NET start keepalive")

	t := time.NewTicker(c.parent.options.keepalive * 3 / 4)
	timeout := time.Duration(float64(c.parent.options.keepalive) * c.parent.options.keepaliveFactor)
	timeoutTimer := time.NewTimer(timeout)

	defer func() {
		t.Stop()
		c.parent.workers.Done()
		c.parent.log.d("NET stop keepalive for server =", c.name)
	}()

	for {
		select {
		case <-c.parent.ctx.Done():
			return
		case <-t.C:
			c.send(PingReqPacket)

			select {
			case <-c.parent.ctx.Done():
				return
			case _, more := <-c.keepaliveC:
				if !more {
					return
				}

				timeoutTimer.Reset(timeout)
			case <-timeoutTimer.C:
				c.parent.log.i("NET keepalive timeout")
				return
			}
		}
	}
}

// handle mqtt logic control packet send
func (c *clientConn) handleSend() {
	defer func() {
		c.parent.workers.Done()
		c.parent.log.e("CONN exit send handler for server =", c.name)
	}()

	for {
		select {
		case <-c.parent.ctx.Done():
			return
		case pkt, more := <-c.parent.sendC:
			if !more {
				return
			}

			if err := EncodeOnePacket(c.parent.options.protoVersion, pkt, c.connW); err != nil {
				c.parent.log.e("NET encode error", err)
				return
			}

			if err := c.connW.Flush(); err != nil {
				c.parent.log.e("NET flush error", err)
				return
			}

			switch pkt.Type() {
			case CtrlPublish:
				p := pkt.(*PublishPacket)
				if p.Qos == 0 {
					c.parent.log.d("CLIENT published qos0 packet, topic =", p.TopicName)
					notifyMsg(c.parent.msgC, newPubMsg(p.TopicName, nil))
				}
			case CtrlDisConn:
				// client exit with disconn
				c.parent.exit()
				return
			}
		case pkt, more := <-c.logicSendC:
			if !more {
				return
			}

			if err := EncodeOnePacket(c.parent.options.protoVersion, pkt, c.connW); err != nil {
				c.parent.log.e("NET encode error", err)
				return
			}

			if err := c.connW.Flush(); err != nil {
				c.parent.log.e("NET flush error", err)
				return
			}

			switch pkt.Type() {
			case CtrlPubRel:
				notifyMsg(c.parent.msgC, newPersistMsg(
					c.parent.persist.Store(sendKey(pkt.(*PubRelPacket).PacketID), pkt)))
			case CtrlPubAck:
				notifyMsg(c.parent.msgC, newPersistMsg(
					c.parent.persist.Delete(sendKey(pkt.(*PubAckPacket).PacketID))))
			case CtrlPubComp:
				notifyMsg(c.parent.msgC, newPersistMsg(
					c.parent.persist.Delete(sendKey(pkt.(*PubCompPacket).PacketID))))
			case CtrlDisConn:
				// disconnect to server
				c.conn.Close()
				return
			}
		}
	}
}

// handle all message receive
func (c *clientConn) handleRecv() {
	defer func() {
		c.parent.workers.Done()
		c.parent.log.e("CONN exit recv handler for server =", c.name)
	}()

	for {
		select {
		case <-c.parent.ctx.Done():
			return
		default:
			pkt, err := DecodeOnePacket(c.parent.options.protoVersion, c.conn)
			if err != nil {
				c.parent.log.e("NET connection broken, server =", c.name, "err =", err)

				// notify net receive no packet will coming
				close(c.netRecvC)

				// notify keepalive worker to exit
				close(c.keepaliveC)

				// TODO send proper net error to net handler
				//if err != ErrDecodeBadPacket {
				//	c.parent.msgC <- newNetMsg(c.name, err)
				//}
				return
			}

			if pkt == PingRespPacket {
				c.parent.log.d("NET received keepalive message")
				c.keepaliveC <- 1
			} else {
				c.netRecvC <- pkt
			}
		}
	}
}

// send mqtt logic packet
func (c *clientConn) send(pkt Packet) {
	if c.parent.isClosing() {
		return
	}

	c.logicSendC <- pkt
}
