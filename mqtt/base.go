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
	"crypto/rand"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	"go.uber.org/zap"
	"gopkg.in/urfave/cli.v1"
)

var (
	log  *zap.Logger
	conf *config
	wg   = &sync.WaitGroup{}
)

var (
	tcpService  net.Listener
	tcpsService net.Listener
	wsService   *http.Server
	wssService  *http.Server
)

// Init mqtt service
func Init(exit context.Context, context *cli.Context) {
	var err error
	log, err = zap.NewDevelopment()
	if err != nil {
		panic(fmt.Sprintf("create mqtt logger failed, error = %s", err.Error()))
	}

	conf = getConfig(context)

	if conf.tcpPort > 0 {
		wg.Add(1)
		go initTCPListen()
	}

	if conf.tcpsPort > 0 {
		wg.Add(1)
		go initTCPSListen()
	}

	if conf.wsPort > 0 {
		wg.Add(1)
		go initWSListen()
	}

	if conf.wssPort > 0 {
		wg.Add(1)
		go initWSSListen()
	}

	wg.Add(1)
	go func() {
		<-exit.Done()
		log.Info("exiting mqtt services")
		destroy(10 * time.Second)
		wg.Done()
	}()

	wg.Wait()
}

func destroy(timeout time.Duration) {
	ctx, _ := context.WithTimeout(context.Background(), timeout)

	if tcpService != nil {
		tcpService.Close()
	}

	if tcpsService != nil {
		tcpsService.Close()
	}

	if wsService != nil {
		wg.Add(1)
		go wsService.Shutdown(ctx)
	}

	if wssService != nil {
		wg.Add(1)
		go wssService.Shutdown(ctx)
	}
}

func initTCPListen() {
	defer wg.Done()

	addr, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:%d", conf.listen, conf.tcpPort))
	if err != nil {
		panic(err)
	}

	tcpService, err := net.ListenTCP("tcp", addr)
	if err != nil {
		panic(err)
	}

	log.Debug("tcp service listening")
	for {
		conn, err := tcpService.Accept()
		if err != nil {
			log.Error("accept tcp connection failed", zap.Error(err))
			continue
		}
		log.Debug("accepted tcp connection")
		go handleConn(conn)
	}
}

func initTCPSListen() {
	defer wg.Done()

	cert, err := tls.LoadX509KeyPair(conf.tlsCertFile, conf.tlsKeyFile)
	if err != nil {
		log.Fatal("load x509 key pair for tcps failed", zap.Error(err))
	}

	config := &tls.Config{
		Certificates: []tls.Certificate{cert},
		Rand:         rand.Reader,
	}

	tcpsService, err := tls.Listen("tcp", fmt.Sprintf("%s:%d", conf.listen, conf.tcpsPort), config)
	if err != nil {
		log.Fatal("tcps listen failed", zap.Error(err))
	}

	log.Debug("tcps service listening")
	for {
		conn, err := tcpsService.Accept()
		if err != nil {
			log.Error("accept tcps connection failed", zap.Error(err))
			continue
		}
		log.Debug("accepted tcps connection", zap.String("addr", conn.RemoteAddr().String()))
		go handleConn(conn)
	}

}

func initWSListen() {
	defer wg.Done()

	mux := http.NewServeMux()
	mux.HandleFunc("/mqtt", handleWS)

	wsService := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", conf.listen, conf.wsPort),
		Handler: mux,
	}

	log.Debug("ws service listening")
	err := wsService.ListenAndServe()
	if err != http.ErrServerClosed {
		log.Error("wss service unexpectedly exited", zap.Error(err))
	}
}

func initWSSListen() {
	defer wg.Done()

	cert, err := tls.LoadX509KeyPair(conf.tlsCertFile, conf.tlsKeyFile)
	if err != nil {
		log.Fatal("load x509 key pair for tcps failed", zap.Error(err))
	}

	config := &tls.Config{
		Certificates: []tls.Certificate{cert},
		Rand:         rand.Reader,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/mqtt", handleWS)
	wssService := &http.Server{
		Addr:      fmt.Sprintf("%s:%d", conf.listen, conf.wssPort),
		TLSConfig: config,
		Handler:   mux,
	}

	log.Debug("wss service listening")
	err = wssService.ListenAndServe()
	if err != http.ErrServerClosed {
		log.Error("wss service unexpectedly exited", zap.Error(err))
	}
}
