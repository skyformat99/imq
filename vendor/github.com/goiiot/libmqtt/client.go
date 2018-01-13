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
	"bufio"
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"io/ioutil"
	"math"
	"net"
	"sync"
	"time"
)

var (
	// ErrTimeOut connection timeout error
	ErrTimeOut = errors.New("connection timeout ")
)

// Option is client option for connection options
type Option func(*client) error

// WithPersist defines the persist method to be used
func WithPersist(method PersistMethod) Option {
	return func(c *client) error {
		if method != nil {
			c.persist = method
		}
		return nil
	}
}

// WithCleanSession will set clean flag in connect packet
func WithCleanSession(f bool) Option {
	return func(c *client) error {
		c.options.cleanSession = f
		return nil
	}
}

// WithIdentity for username and password
func WithIdentity(username, password string) Option {
	return func(c *client) error {
		c.options.username = username
		c.options.password = password
		return nil
	}
}

// WithKeepalive set the keepalive interval (time in second)
func WithKeepalive(keepalive uint16, factor float64) Option {
	return func(c *client) error {
		c.options.keepalive = time.Duration(keepalive) * time.Second
		if factor > 1 {
			c.options.keepaliveFactor = factor
		} else {
			factor = 1.2
		}
		return nil
	}
}

// WithBackoffStrategy will set reconnect backoff strategy
// firstDelay is the time to wait before retrying after the first failure
// maxDelay defines the upper bound of backoff delay
// factor is applied to the backoff after each retry.
// e.g. FirstDelay = 1s and Factor = 2, then the SecondDelay is 2s, the ThirdDelay is 4s
func WithBackoffStrategy(firstDelay, maxDelay time.Duration, factor float64) Option {
	return func(c *client) error {
		if firstDelay < time.Millisecond {
			firstDelay = time.Millisecond
		}

		if maxDelay < firstDelay {
			maxDelay = firstDelay
		}

		if factor < 1 {
			factor = 1
		}

		c.options.firstDelay = firstDelay
		c.options.maxDelay = maxDelay
		c.options.backoffFactor = factor
		return nil
	}
}

// WithClientID set the client id for connection
func WithClientID(clientID string) Option {
	return func(c *client) error {
		c.options.clientID = clientID
		return nil
	}
}

// WithWill mark this connection as a will teller
func WithWill(topic string, qos QosLevel, retain bool, payload []byte) Option {
	return func(c *client) error {
		c.options.isWill = true
		c.options.willTopic = topic
		c.options.willQos = qos
		c.options.willRetain = retain
		c.options.willPayload = payload
		return nil
	}
}

// WithServer adds servers as client server
// Just use "ip:port" or "domain.name:port"
// However, only TCP connection supported for now
func WithServer(servers ...string) Option {
	return func(c *client) error {
		c.options.servers = servers
		return nil
	}
}

// WithTLS for client tls certification
func WithTLS(certFile, keyFile string, caCert string, serverNameOverride string, skipVerify bool) Option {
	return func(c *client) error {
		b, err := ioutil.ReadFile(caCert)
		if err != nil {
			return err
		}
		cp := x509.NewCertPool()
		if !cp.AppendCertsFromPEM(b) {
			return err
		}
		cert, err := tls.LoadX509KeyPair(certFile, keyFile)
		if err != nil {
			return err
		}

		c.options.tlsConfig = &tls.Config{
			Certificates:       []tls.Certificate{cert},
			InsecureSkipVerify: skipVerify,
			ClientCAs:          cp,
			ServerName:         serverNameOverride,
		}
		return nil
	}
}

// WithDialTimeout for connection time out (time in second)
func WithDialTimeout(timeout uint16) Option {
	return func(c *client) error {
		c.options.dialTimeout = time.Duration(timeout) * time.Second
		return nil
	}
}

// WithSendBuf designate the channel size of send
func WithSendBuf(size int) Option {
	return func(c *client) error {
		if size < 1 {
			size = 1
		} else if size > 1024 {
			size = 1024
		}
		c.options.sendChanSize = size
		return nil
	}
}

// WithRecvBuf designate the channel size of receive
func WithRecvBuf(size int) Option {
	return func(c *client) error {
		if size < 1 {
			size = 1
		} else if size > 1024 {
			size = 1024
		}
		c.options.recvChanSize = size
		return nil
	}
}

// WithRouter set the router for topic dispatch
func WithRouter(r TopicRouter) Option {
	return func(c *client) error {
		if r != nil {
			c.router = r
		}
		return nil
	}
}

// WithLog will create basic logger for log
func WithLog(l LogLevel) Option {
	return func(c *client) error {
		c.log = newLogger(l)
		return nil
	}
}

// WithVersion defines the mqtt protocol version to
func withVersion(version ProtocolVersion, compromise bool) Option {
	return func(c *client) error {
		c.options.protoVersion = version
		c.options.protoCompromise = compromise
		return nil
	}
}

// NewClient create a new mqtt client
func NewClient(options ...Option) (Client, error) {
	c := defaultClient()

	for _, o := range options {
		err := o(c)
		if err != nil {
			return nil, err
		}
	}

	if len(c.options.servers) < 1 {
		return nil, errors.New("no server provided, won't work ")
	}

	c.sendC = make(chan Packet, c.options.sendChanSize)
	c.recvC = make(chan *PublishPacket, c.options.recvChanSize)

	return c, nil
}

// Client act as a mqtt client
type Client interface {
	// Handle register topic handlers, mostly used for RegexHandler, RestHandler
	// the default handler inside the client is TextHandler, which match the exactly same topic
	Handle(topic string, h TopicHandler)

	// Connect to all specified server with client options
	Connect(ConnHandler)

	// Publish a message for the topic
	Publish(packets ...*PublishPacket)

	// Subscribe topic(s)
	Subscribe(topics ...*Topic)

	// UnSubscribe topic(s)
	UnSubscribe(topics ...string)

	// Wait will wait until all connection finished
	Wait()

	// Destroy all client connection
	Destroy(force bool)

	// handlers
	HandlePub(PubHandler)
	HandleSub(SubHandler)
	HandleUnSub(UnSubHandler)
	HandleNet(NetHandler)
	HandlePersist(PersistHandler)
}

// mqtt client implementation
type client struct {
	options *clientOptions      // client connection options
	msgC    chan *message       // error channel
	sendC   chan Packet         // Pub channel for sending publish packet to server
	recvC   chan *PublishPacket // recv channel for server pub receiving
	idGen   *idGenerator        // Packet id generator
	router  TopicRouter         // Topic router
	persist PersistMethod       // Persist method
	workers *sync.WaitGroup     // Workers (goroutines)
	log     *logger             // client logger

	// success/error handlers
	pubHandler     PubHandler
	subHandler     SubHandler
	unSubHandler   UnSubHandler
	netHandler     NetHandler
	persistHandler PersistHandler

	ctx  context.Context    // closure of this channel will signal all client worker to stop
	exit context.CancelFunc // called when client exit
}

// clientOptions is the options for client to connect, reconnect, disconnect
type clientOptions struct {
	protoVersion    ProtocolVersion // mqtt protocol version
	protoCompromise bool            // compromise to server protocol version
	sendChanSize    int             // send channel size
	recvChanSize    int             // recv channel size
	servers         []string        // server address strings
	dialTimeout     time.Duration   // dial timeout in second
	clientID        string          // used by ConnPacket
	username        string          // used by ConnPacket
	password        string          // used by ConnPacket
	keepalive       time.Duration   // used by ConnPacket (time in second)
	keepaliveFactor float64         // used for reasonable amount time to close conn if no ping resp
	cleanSession    bool            // used by ConnPacket
	isWill          bool            // used by ConnPacket
	willTopic       string          // used by ConnPacket
	willPayload     []byte          // used by ConnPacket
	willQos         byte            // used by ConnPacket
	willRetain      bool            // used by ConnPacket
	tlsConfig       *tls.Config     // tls config with client side cert
	maxDelay        time.Duration
	firstDelay      time.Duration
	backoffFactor   float64
}

// create a client with default options
func defaultClient() *client {
	ctx, cancel := context.WithCancel(context.TODO())
	return &client{
		options: &clientOptions{
			sendChanSize:    128,
			recvChanSize:    128,
			maxDelay:        2 * time.Minute, // default max retry delay is 2min
			firstDelay:      5 * time.Second, // first retry delay is 5s
			backoffFactor:   1.5,
			dialTimeout:     20 * time.Second, // default timeout when dial to server
			keepalive:       2 * time.Minute,  // default keepalive interval is 2min
			keepaliveFactor: 1.5,              // default reasonable amount of time 3min
			protoVersion:    V311,             // default protocol version is MQTT 3.1.1
		},
		msgC:    make(chan *message),
		ctx:     ctx,
		exit:    cancel,
		router:  NewTextRouter(),
		idGen:   newIDGenerator(),
		workers: &sync.WaitGroup{},
		persist: NonePersist,
	}
}

// Handle register subscription message route
func (c *client) Handle(topic string, h TopicHandler) {
	if h != nil {
		c.log.d("HANDLE registered topic handler, topic =", topic)
		c.router.Handle(topic, h)
	}
}

// Connect to all designated server
func (c *client) Connect(h ConnHandler) {
	c.log.d("CLIENT connect to server, handler =", h)

	for _, s := range c.options.servers {
		c.workers.Add(1)
		go c.connect(s, h, c.options.firstDelay)
	}

	c.workers.Add(2)
	go c.handleTopicMsg()
	go c.handleMsg()
}

// Publish message(s) to topic(s), one to one
func (c *client) Publish(msg ...*PublishPacket) {
	if c.isClosing() {
		return
	}

	for _, m := range msg {
		if m == nil {
			continue
		}

		p := m
		if p.Qos > Qos2 {
			p.Qos = Qos2
		}

		if p.Qos != Qos0 {
			if p.PacketID == 0 {
				p.PacketID = c.idGen.next(p)
				if err := c.persist.Store(sendKey(p.PacketID), p); err != nil {
					c.msgC <- newPersistMsg(err)
				}
			}
		}
		c.sendC <- p
	}
}

// Subscribe topic(s)
func (c *client) Subscribe(topics ...*Topic) {
	if c.isClosing() {
		return
	}

	c.log.d("CLIENT subscribe, topic(s) =", topics)

	s := &SubscribePacket{Topics: topics}
	s.PacketID = c.idGen.next(s)

	c.sendC <- s
}

// UnSubscribe topic(s)
func (c *client) UnSubscribe(topics ...string) {
	if c.isClosing() {
		return
	}

	c.log.d("CLIENT unsubscribe topic(s) =", topics)

	u := &UnSubPacket{TopicNames: topics}
	u.PacketID = c.idGen.next(u)

	c.sendC <- u
}

// Wait will wait for all connection to exit
func (c *client) Wait() {
	if c.isClosing() {
		return
	}

	c.log.i("CLIENT wait for all workers")
	c.workers.Wait()
}

// Destroy will disconnect form all server
// If force is true, then close connection without sending a DisConnPacket
func (c *client) Destroy(force bool) {
	c.log.d("CLIENT destroying client with force =", force)
	if force {
		c.exit()
	} else {
		c.sendC <- DisConnPacket
	}
}

// HandlePubMsg register handler for pub error
func (c *client) HandlePub(h PubHandler) {
	c.log.d("CLIENT registered pub handler")
	c.pubHandler = h
}

// HandleSubMsg register handler for extra sub info
func (c *client) HandleSub(h SubHandler) {
	c.log.d("CLIENT registered sub handler")
	c.subHandler = h
}

// HandleUnSubMsg register handler for unsubscription error
func (c *client) HandleUnSub(h UnSubHandler) {
	c.log.d("CLIENT registered unsub handler")
	c.unSubHandler = h
}

// HandleNet register handler for net error
func (c *client) HandleNet(h NetHandler) {
	c.log.d("CLIENT registered net handler")
	c.netHandler = h
}

// HandleNet register handler for net error
func (c *client) HandlePersist(h PersistHandler) {
	c.log.d("CLIENT registered persist handler")
	c.persistHandler = h
}

// connect to one server and start mqtt logic
func (c *client) connect(server string, h ConnHandler, reconnectDelay time.Duration) {
	var conn net.Conn
	var err error

	if c.options.tlsConfig != nil {
		// with tls
		conn, err = tls.DialWithDialer(&net.Dialer{Timeout: c.options.dialTimeout}, "tcp", server, c.options.tlsConfig)
		if err != nil {
			c.log.e("CLIENT connect with tls failed, err =", err, "server =", server)
			if h != nil {
				h(server, math.MaxUint8, err)
			}
			return
		}
	} else {
		// without tls
		conn, err = net.DialTimeout("tcp", server, c.options.dialTimeout)
		if err != nil {
			c.log.e("CLIENT connect failed, err =", err, "server =", server)
			if h != nil {
				h(server, math.MaxUint8, err)
			}
			return
		}
	}

	if c.isClosing() {
		conn.Close()
		return
	}

	connImpl := &clientConn{
		parent:     c,
		name:       server,
		conn:       conn,
		connW:      bufio.NewWriter(conn),
		keepaliveC: make(chan int),
		logicSendC: make(chan Packet),
		netRecvC:   make(chan Packet),
	}

	c.workers.Add(2)
	go connImpl.handleSend()
	go connImpl.handleRecv()

	connImpl.send(&ConnPacket{
		Username:     c.options.username,
		Password:     c.options.password,
		ClientID:     c.options.clientID,
		CleanSession: c.options.cleanSession,
		IsWill:       c.options.isWill,
		WillQos:      c.options.willQos,
		WillTopic:    c.options.willTopic,
		WillMessage:  c.options.willPayload,
		WillRetain:   c.options.willRetain,
		Keepalive:    uint16(c.options.keepalive / time.Second),
	})

	select {
	case <-c.ctx.Done():
		return
	case pkt, more := <-connImpl.netRecvC:
		if more {
			if pkt.Type() == CtrlConnAck {
				p := pkt.(*ConnAckPacket)
				if p.Code != ConnAccepted {
					if h != nil {
						h(server, p.Code, nil)
					}
					return
				}
			} else {
				if h != nil {
					h(server, math.MaxUint8, ErrDecodeBadPacket)
				}
				return
			}
		} else {
			if h != nil {
				h(server, math.MaxUint8, ErrDecodeBadPacket)
			}
			return
		}
	case <-time.After(c.options.dialTimeout):
		if h != nil {
			h(server, math.MaxUint8, ErrTimeOut)
		}
		return
	}

	c.log.i("CLIENT connected to server =", server)
	if h != nil {
		go h(server, ConnAccepted, nil)
	}

	// login success, start mqtt logic
	connImpl.logic()

	if c.isClosing() {
		return
	}

	// reconnect
	c.log.e("CLIENT reconnecting to server =", server, "delay =", reconnectDelay)
	time.Sleep(reconnectDelay)

	if c.isClosing() {
		return
	}

	reconnectDelay = time.Duration(float64(reconnectDelay) * c.options.backoffFactor)
	if reconnectDelay > c.options.maxDelay {
		reconnectDelay = c.options.maxDelay
	}
	c.connect(server, h, reconnectDelay)
}

func (c *client) isClosing() bool {
	select {
	case <-c.ctx.Done():
		return true
	default:
		return false
	}
}

func (c *client) handleTopicMsg() {
	defer c.workers.Done()

	for {
		select {
		case <-c.ctx.Done():
			return
		case pkt, more := <-c.recvC:
			if !more {
				return
			}

			c.router.Dispatch(pkt)
		}
	}
}

func (c *client) handleMsg() {
	defer c.workers.Done()

	for {
		select {
		case <-c.ctx.Done():
			return
		case m, more := <-c.msgC:
			if !more {
				return
			}

			switch m.what {
			case pubMsg:
				if c.pubHandler != nil {
					go c.pubHandler(m.msg, m.err)
				}
			case subMsg:
				if c.subHandler != nil {
					go c.subHandler(m.obj.([]*Topic), m.err)
				}
			case unSubMsg:
				if c.unSubHandler != nil {
					go c.unSubHandler(m.obj.([]string), m.err)
				}
			case netMsg:
				if c.netHandler != nil {
					go c.netHandler(m.msg, m.err)
				}
			case persistMsg:
				if c.persistHandler != nil {
					go c.persistHandler(m.err)
				}
			}
		}
	}
}
