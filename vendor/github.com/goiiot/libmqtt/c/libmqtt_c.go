// +build cgo lib

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

package main

/*
#include <stdio.h>
#include <stdlib.h>
#include <stdbool.h>

#ifndef _LIBMQTT_CONN_H_
#define _LIBMQTT_CONN_H_

typedef enum {
  libmqtt_connack_accepted = 0,
  libmqtt_connack_bad_proto = 1,
  libmqtt_connack_id_rejected = 2,
  libmqtt_connack_srv_unavail = 3,
  libmqtt_connack_bad_identity = 4,
  libmqtt_connack_auth_fail = 5,
} libmqtt_connack_t;

#endif

#ifndef _LIBMQTT_SUB_H_
#define _LIBMQTT_SUB_H_

typedef enum {
  libmqtt_suback_max_qos0 = 0,
  libmqtt_suback_max_qos1 = 1,
  libmqtt_suback_max_qos2 = 2,
  libmqtt_suback_fail = 0x80,
} libmqtt_suback_t;

#endif

#ifndef _LIBMQTT_BRIDGE_H_
#define _LIBMQTT_BRIDGE_H_

typedef void (*libmqtt_conn_handler)
  (int client, char *server, libmqtt_connack_t code, char *err);

typedef void (*libmqtt_pub_handler)
  (int client, char *topic, char *err);

typedef void (*libmqtt_sub_handler)
  (int client, char *topic, int qos, char *err);

typedef void (*libmqtt_unsub_handler)
  (int client, char *topic, char *err);

typedef void (*libmqtt_net_handler)
  (int client, char *server, char *err);

typedef void (*libmqtt_persist_handler)
  (int client, char *err);

typedef void (*libmqtt_topic_handler)
  (int client, char *topic, int qos, char *msg, int size);

static inline void call_conn_handler
  (libmqtt_conn_handler h, int client, char * server, libmqtt_connack_t code, char * err) {
  if (h != NULL) {
    h(client, server, code, err);
    free(server);
    free(err);
  }
}

static inline void call_pub_handler
  (libmqtt_pub_handler h, int client, char * topic, char * err) {
  if (h != NULL) {
    h(client, topic, err);
    free(topic);
    free(err);
  }
}

static inline void call_sub_handler
  (libmqtt_sub_handler h, int client, char * topic, int qos, char * err) {
  if (h != NULL) {
    h(client, topic, qos, err);
    free(topic);
    free(err);
  }
}

static inline void call_unsub_handler
  (libmqtt_unsub_handler h, int client, char * topic, char * err) {
  if (h != NULL) {
    h(client, topic, err);
    free(topic);
    free(err);
  }
}

static inline void call_net_handler
  (libmqtt_net_handler h, int client, char * server, char * err) {
  if (h != NULL) {
    h(client, server, err);
    free(server);
    free(err);
  }
}

static inline void call_persist_handler
  (libmqtt_persist_handler h, int client, char *err) {
  if (h != NULL) {
    h(client, err);
    free(err);
  }
}

static inline void call_topic_handler
  (libmqtt_topic_handler h, int client, char * topic, int qos , char * msg, int size) {
  if (h != NULL) {
    h(client, topic, qos, msg, size);
    free(topic);
    free(msg);
  }
}
#endif
*/
import "C"
import (
	"unsafe"

	lib "github.com/goiiot/libmqtt"
)

var (
	clients = make(map[int]lib.Client)
)

// Libmqtt_handle (int client, char *topic, libmqtt_topic_handler h)
//export Libmqtt_handle
func Libmqtt_handle(client C.int, topic *C.char, h C.libmqtt_topic_handler) {
	if c, ok := clients[int(client)]; ok {
		c.Handle(C.GoString(topic), wrapTopicHandler(client, h))
	}
}

// Libmqtt_connect (int client)
//export Libmqtt_connect
func Libmqtt_connect(client C.int, h C.libmqtt_conn_handler) {
	if c, ok := clients[int(client)]; ok {
		c.Connect(func(server string, code lib.ConnAckCode, err error) {
			var c C.libmqtt_connack_t
			switch code {
			case lib.ConnAccepted:
				c = C.libmqtt_connack_accepted
			case lib.ConnBadProtocol:
				c = C.libmqtt_connack_bad_proto
			case lib.ConnIDRejected:
				c = C.libmqtt_connack_id_rejected
			case lib.ConnServerUnavailable:
				c = C.libmqtt_connack_srv_unavail
			case lib.ConnBadIdentity:
				c = C.libmqtt_connack_bad_identity
			case lib.ConnAuthFail:
				c = C.libmqtt_connack_auth_fail
			}

			var er *C.char
			if err != nil {
				er = C.CString(err.Error())
			}
			C.call_conn_handler(h, client, C.CString(server), c, er)
		})
	}
}

// Libmqtt_subscribe (int client, char *topic, int qos)
//export Libmqtt_subscribe
func Libmqtt_subscribe(client C.int, topic *C.char, qos C.int) {
	if c, ok := clients[int(client)]; ok {
		c.Subscribe(&lib.Topic{
			Name: C.GoString(topic),
			Qos:  lib.QosLevel(qos),
		})
	}
}

// Libmqtt_publish (int client, char *topic, int qos, char *payload, int payloadSize)
//export Libmqtt_publish
func Libmqtt_publish(client C.int, topic *C.char, qos C.int, payload *C.char, payloadSize C.int) {
	if c, ok := clients[int(client)]; ok {
		c.Publish(&lib.PublishPacket{
			TopicName: C.GoString(topic),
			Qos:       lib.QosLevel(qos),
			Payload:   C.GoBytes(unsafe.Pointer(payload), payloadSize),
		})
	}
}

// Libmqtt_unsubscribe (int client, char *topic)
//export Libmqtt_unsubscribe
func Libmqtt_unsubscribe(client C.int, topic *C.char) {
	cid := int(client)
	if c, ok := clients[cid]; ok {
		c.UnSubscribe(C.GoString(topic))
	}
}

// Libmqtt_wait (int client)
//export Libmqtt_wait
func Libmqtt_wait(client C.int) {
	if c, ok := clients[int(client)]; ok {
		c.Wait()
	}
}

// Libmqtt_destroy (int client, bool force)
//export Libmqtt_destroy
func Libmqtt_destroy(client C.int, force C.bool) {
	if c, ok := clients[int(client)]; ok {
		c.Destroy(bool(force))
	}
}

// Libmqtt_set_pub_handler (int client, libmqtt_pub_handler h)
//export Libmqtt_set_pub_handler
func Libmqtt_set_pub_handler(client C.int, h C.libmqtt_pub_handler) {
	if c, ok := clients[int(client)]; ok {
		c.HandlePub(func(topic string, err error) {
			var er *C.char
			if err != nil {
				er = C.CString(err.Error())
			}
			C.call_pub_handler(h, client, C.CString(topic), er)
		})
	}
}

// Libmqtt_set_sub_handler (int client, libmqtt_sub_handler h)
//export Libmqtt_set_sub_handler
func Libmqtt_set_sub_handler(client C.int, h C.libmqtt_sub_handler) {
	if c, ok := clients[int(client)]; ok {
		c.HandleSub(func(topics []*lib.Topic, err error) {
			for _, t := range topics {
				var er *C.char
				if err != nil {
					er = C.CString(err.Error())
				}
				C.call_sub_handler(h, client, C.CString(t.Name), C.int(t.Qos), er)
			}
		})
	}
}

// Libmqtt_set_unsub_handler (int client, libmqtt_unsub_handler h)
//export Libmqtt_set_unsub_handler
func Libmqtt_set_unsub_handler(client C.int, h C.libmqtt_unsub_handler) {
	if c, ok := clients[int(client)]; ok {
		c.HandleUnSub(func(topics []string, err error) {
			for _, t := range topics {
				var er *C.char
				if err != nil {
					er = C.CString(err.Error())
				}
				C.call_unsub_handler(h, client, C.CString(t), er)
			}
		})
	}
}

// Libmqtt_set_net_handler (int client, libmqtt_net_handler h)
//export Libmqtt_set_net_handler
func Libmqtt_set_net_handler(client C.int, h C.libmqtt_net_handler) {
	if c, ok := clients[int(client)]; ok {
		c.HandleNet(func(server string, err error) {
			if err != nil {
				C.call_net_handler(h, client, C.CString(server), C.CString(err.Error()))
			}
		})
	}
}

// Libmqtt_set_persist_handler (int client, libmqtt_persist_handler h)
//export Libmqtt_set_persist_handler
func Libmqtt_set_persist_handler(client C.int, h C.libmqtt_persist_handler) {
	if c, ok := clients[int(client)]; ok {
		c.HandlePersist(func(err error) {
			if err != nil {
				C.call_persist_handler(h, client, C.CString(err.Error()))
			}
		})
	}
}

func wrapTopicHandler(client C.int, h C.libmqtt_topic_handler) lib.TopicHandler {
	return func(topic string, qos lib.QosLevel, msg []byte) {
		C.call_topic_handler(h, client, C.CString(topic), C.int(qos),
			(*C.char)(C.CBytes(msg)), C.int(len(msg)))
	}
}

func main() {}
