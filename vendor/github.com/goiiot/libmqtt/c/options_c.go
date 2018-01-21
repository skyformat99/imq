// +build cgo lib

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

package main

/*
#ifndef _LIBMQTT_LOG_H_
#define _LIBMQTT_LOG_H_

typedef enum {
  libmqtt_log_silent = 0,
  libmqtt_log_verbose = 1,
  libmqtt_log_debug = 2,
  libmqtt_log_info = 3,
  libmqtt_log_warning = 4,
  libmqtt_log_error = 5,
} libmqtt_log_level;

#endif //_LIBMQTT_LOG_H_
*/
import "C"
import (
	"math"
	"time"
	"unsafe"

	mqtt "github.com/goiiot/libmqtt"
)

var (
	clientOptions = make(map[int][]mqtt.Option)
)

// Libmqtt_new_client ()
// create a new client, return the id of client
// if no client id available return -1
//export Libmqtt_new_client
func Libmqtt_new_client() C.int {
	for id := 1; id < int(math.MaxInt64); id++ {
		if _, ok := clientOptions[id]; !ok {
			clientOptions[id] = make([]mqtt.Option, 0)
			return C.int(id)
		}
	}
	return -1
}

// Libmqtt_setup_client (int client)
// setup the client with previously defined options
//export Libmqtt_setup_client
func Libmqtt_setup_client(client C.int) *C.char {
	cid := int(client)
	if v, ok := clientOptions[cid]; ok {
		c, err := mqtt.NewClient(v...)
		if err != nil {
			return C.CString(err.Error())
		}

		clients[cid] = c
	} else {
		return C.CString("no such client")
	}

	return nil
}

// Libmqtt_client_with_backoff_strategy (int client, int first_delay, int maxDelay, double factor)
//Libmqtt_client_with_backoff_strategy
func Libmqtt_client_with_backoff_strategy(client C.int, firstDelay C.int, maxDelay C.int, factor C.double) {
	addOption(client, mqtt.WithBackoffStrategy(
		time.Duration(firstDelay)*time.Millisecond,
		time.Duration(maxDelay)*time.Millisecond,
		float64(factor),
	))
}

// Libmqtt_client_with_clean_session (bool flag)
//export Libmqtt_client_with_clean_session
func Libmqtt_client_with_clean_session(client C.int, flag bool) {
	addOption(client, mqtt.WithCleanSession(flag))
}

// Libmqtt_client_with_client_id (int client, char * client_id)
//export Libmqtt_client_with_client_id
func Libmqtt_client_with_client_id(client C.int, clientID *C.char) {
	addOption(client, mqtt.WithClientID(C.GoString(clientID)))
}

// Libmqtt_client_with_dial_timeout (int client, int timeout)
//export Libmqtt_client_with_dial_timeout
func Libmqtt_client_with_dial_timeout(client C.int, timeout C.int) {
	addOption(client, mqtt.WithDialTimeout(uint16(timeout)))
}

// Libmqtt_client_with_identity (int client, char * username, char * password)
//export Libmqtt_client_with_identity
func Libmqtt_client_with_identity(client C.int, username, password *C.char) {
	addOption(client, mqtt.WithIdentity(C.GoString(username), C.GoString(password)))
}

// Libmqtt_client_with_keepalive (int client, int keepalive, float factor)
//export Libmqtt_client_with_keepalive
func Libmqtt_client_with_keepalive(client C.int, keepalive C.int, factor C.float) {
	addOption(client, mqtt.WithKeepalive(uint16(keepalive), float64(factor)))
}

// Libmqtt_client_with_log (int client, libmqtt_log_level l)
//export Libmqtt_client_with_log
func Libmqtt_client_with_log(client C.int, l C.libmqtt_log_level) {
	addOption(client, mqtt.WithLog(mqtt.LogLevel(l)))
}

// Libmqtt_client_with_none_persist (int client)
//export Libmqtt_client_with_none_persist
func Libmqtt_client_with_none_persist(client C.int) {
	addOption(client, mqtt.WithPersist(mqtt.NonePersist))
}

// Libmqtt_client_with_mem_persist (int client, int maxCount, bool dropOnExceed, bool duplicateReplace)
//export Libmqtt_client_with_mem_persist
func Libmqtt_client_with_mem_persist(client C.int, maxCount C.int, dropOnExceed bool, duplicateReplace bool) {
	addOption(client, mqtt.WithPersist(mqtt.NewMemPersist(&mqtt.PersistStrategy{
		MaxCount:         uint32(maxCount),
		DropOnExceed:     dropOnExceed,
		DuplicateReplace: duplicateReplace,
	})))
}

// Libmqtt_client_with_redis_persist use redis as persist method
func Libmqtt_client_with_redis_persist(client C.int) {

}

// Libmqtt_client_with_file_persist (int client, char *dirPath, int maxCount, bool dropOnExceed, bool duplicateReplace)
//export Libmqtt_client_with_file_persist
func Libmqtt_client_with_file_persist(client C.int, dirPath *C.char, maxCount C.int, dropOnExceed bool, duplicateReplace bool) {
	addOption(client, mqtt.WithPersist(mqtt.NewFilePersist(C.GoString(dirPath), &mqtt.PersistStrategy{
		MaxCount:         uint32(maxCount),
		DropOnExceed:     dropOnExceed,
		DuplicateReplace: duplicateReplace,
	})))
}

// Libmqtt_client_with_buf (int client, int size)
//export Libmqtt_client_with_buf
func Libmqtt_client_with_buf(client C.int, sendBuf C.int, recvBuf C.int) {
	addOption(client, mqtt.WithBuf(int(sendBuf), int(recvBuf)))
}

// Libmqtt_client_with_server (int client, char *server)
//export Libmqtt_client_with_server
func Libmqtt_client_with_server(client C.int, server *C.char) {
	addOption(client, mqtt.WithServer(C.GoString(server)))
}

// Libmqtt_client_with_std_router use standard router
func Libmqtt_client_with_std_router(client C.int) {

}

// Libmqtt_client_with_text_router use text router
func Libmqtt_client_with_text_router(client C.int) {

}

// Libmqtt_client_with_regex_router use regex matching for router
func Libmqtt_client_with_regex_router(client C.int) {

}

// Libmqtt_client_with_http_router use http path router
func Libmqtt_client_with_http_router(client C.int) {

}

// Libmqtt_client_with_tls (int client, char * certFile, char * keyFile, char * caCert, char * serverNameOverride, bool skipVerify)
// use ssl to connect
//export Libmqtt_client_with_tls
func Libmqtt_client_with_tls(client C.int, certFile, keyFile, caCert, serverNameOverride *C.char, skipVerify bool) {
	addOption(client, mqtt.WithTLS(
		C.GoString(certFile),
		C.GoString(keyFile),
		C.GoString(caCert),
		C.GoString(serverNameOverride),
		skipVerify))
}

// Libmqtt_client_with_will (int client, char *topic, int qos, bool retain, char *payload, int payloadSize)
// mark this connection with will message
//export Libmqtt_client_with_will
func Libmqtt_client_with_will(client C.int, topic *C.char, qos C.int, retain bool, payload *C.char, payloadSize C.int) {
	addOption(client, mqtt.WithWill(
		C.GoString(topic),
		mqtt.QosLevel(qos),
		retain,
		C.GoBytes(unsafe.Pointer(payload), payloadSize)))
}

func addOption(client C.int, option mqtt.Option) {
	cid := int(client)
	if v, ok := clientOptions[cid]; ok {
		v = append(v, option)
		clientOptions[cid] = v
	}
}
