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

#include "cc_goiiot_libmqtt_LibMQTT.h"
#include "libmqtt.h"
#include "handlers_jni.h"

/*
 * Method:    _newClient
 * Signature: ()I
 */
JNIEXPORT jint JNICALL
Java_cc_goiiot_libmqtt_LibMQTT__1newClient
(JNIEnv *env, jclass c) {

  return Libmqtt_new_client();
}

/*
 * Method:    _setServer
 * Signature: (ILjava/lang/String;)V
 */
JNIEXPORT void JNICALL
Java_cc_goiiot_libmqtt_LibMQTT__1setServer
(JNIEnv *env, jclass c, jint id, jstring server) {

  const char *srv = (*env)->GetStringUTFChars(env, server, 0);

  Libmqtt_client_with_server(id, srv);

  (*env)->ReleaseStringUTFChars(env, server, srv);
}

/*
 * Method:    _setCleanSession
 * Signature: (IZ)V
 */
JNIEXPORT void JNICALL
Java_cc_goiiot_libmqtt_LibMQTT__1setCleanSession
(JNIEnv *env, jclass c, jint id, jboolean flag) {

  Libmqtt_client_with_clean_session(id, flag);
}

/*
 * Method:    _setKeepalive
 * Signature: (IID)V
 */
JNIEXPORT void JNICALL
Java_cc_goiiot_libmqtt_LibMQTT__1setKeepalive
(JNIEnv *env, jclass c, jint id, jint keepalive, jdouble factor) {

  Libmqtt_client_with_keepalive(id, keepalive, factor);
}

/*
 * Method:    _setClientID
 * Signature: (ILjava/lang/String;)V
 */
JNIEXPORT void JNICALL
Java_cc_goiiot_libmqtt_LibMQTT__1setClientID
(JNIEnv *env, jclass c, jint id, jstring client_id) {

  const char *cid = (*env)->GetStringUTFChars(env, client_id, 0);

  Libmqtt_client_with_client_id(id, cid);

  (*env)->ReleaseStringUTFChars(env, client_id, cid);
}

/*
 * Method:    _setDialTimeout
 * Signature: (II)V
 */
JNIEXPORT void JNICALL
Java_cc_goiiot_libmqtt_LibMQTT__1setDialTimeout
(JNIEnv *env, jclass c, jint id, jint timeout) {

  Libmqtt_client_with_dial_timeout(id, timeout);
}

/*
 * Method:    _setIdentity
 * Signature: (ILjava/lang/String;Ljava/lang/String;)V
 */
JNIEXPORT void JNICALL
Java_cc_goiiot_libmqtt_LibMQTT__1setIdentity
(JNIEnv *env, jclass c, jint id, jstring username, jstring password) {

  const char *user = (*env)->GetStringUTFChars(env, username, 0);
  const char *pass = (*env)->GetStringUTFChars(env, password, 0);

  Libmqtt_client_with_identity(id, user, pass);

  (*env)->ReleaseStringUTFChars(env, username, user);
  (*env)->ReleaseStringUTFChars(env, password, pass);
}

/*
 * Method:    _setLog
 * Signature: (II)V
 */
JNIEXPORT void JNICALL
Java_cc_goiiot_libmqtt_LibMQTT__1setLog
(JNIEnv *env, jclass c, jint id, jint log_level) {

  Libmqtt_client_with_log(id, log_level);
}

/*
 * Method:    _setSendBuf
 * Signature: (II)V
 */
JNIEXPORT void JNICALL
Java_cc_goiiot_libmqtt_LibMQTT__1setSendBuf
(JNIEnv *env, jclass c, jint id, jint size) {

  Libmqtt_client_with_send_buf(id, size);
}

/*
 * Method:    _setRecvBuf
 * Signature: (II)V
 */
JNIEXPORT void JNICALL
Java_cc_goiiot_libmqtt_LibMQTT__1setRecvBuf
(JNIEnv *env, jclass c, jint id, jint size) {

  Libmqtt_client_with_send_buf(id, size);
}

/*
 * Method:    _setTLS
 * Signature: (ILjava/lang/String;Ljava/lang/String;Ljava/lang/String;Ljava/lang/String;Z)V
 */
JNIEXPORT void JNICALL
Java_cc_goiiot_libmqtt_LibMQTT__1setTLS
(JNIEnv *env, jclass c, jint id, jstring cert,
 jstring key, jstring ca, jstring srv_name, jboolean skip_verify) {

  const char *c_cert = (*env)->GetStringUTFChars(env, cert, 0);
  const char *c_key = (*env)->GetStringUTFChars(env, key, 0);
  const char *c_ca = (*env)->GetStringUTFChars(env, ca, 0);
  const char *c_srv = (*env)->GetStringUTFChars(env, srv_name, 0);

  Libmqtt_client_with_tls(id, c_cert, c_key, c_ca, c_srv, skip_verify);

  (*env)->ReleaseStringUTFChars(env, cert, c_cert);
  (*env)->ReleaseStringUTFChars(env, key, c_key);
  (*env)->ReleaseStringUTFChars(env, ca, c_ca);
  (*env)->ReleaseStringUTFChars(env, srv_name, c_srv);
}

/*
 * Method:    _setWill
 * Signature: (ILjava/lang/String;IZ[B)V
 */
JNIEXPORT void JNICALL
Java_cc_goiiot_libmqtt_LibMQTT__1setWill
(JNIEnv *env, jclass c, jint id, jstring topic,
 jint qos, jboolean retain, jbyteArray payload) {

  const char *c_topic = (*env)->GetStringUTFChars(env, topic, 0);
  jsize len = (*env)->GetArrayLength(env, payload);
  jbyte *body = (*env)->GetByteArrayElements(env, payload, 0);

  Libmqtt_client_with_will(id, c_topic, qos, retain, body, len);

  (*env)->ReleaseStringUTFChars(env, topic, c_topic);
  (*env)->ReleaseByteArrayElements(env, payload, body, 0);
}

/*
 * Method:    _setNonePersist
 * Signature: (I)V
 */
JNIEXPORT void JNICALL
Java_cc_goiiot_libmqtt_LibMQTT__1setNonePersist
(JNIEnv *env, jclass c, jint id) {

  Libmqtt_client_with_none_persist(id);
}

/*
 * Method:    _setMemPersist
 * Signature: (IIZZ)V
 */
JNIEXPORT void JNICALL
Java_cc_goiiot_libmqtt_LibMQTT__1setMemPersist
(JNIEnv *env, jclass c, jint id, jint max_count,
 jboolean ex_drop, jboolean dup_replace) {

  Libmqtt_client_with_mem_persist(id, max_count, ex_drop, dup_replace);
}

/*
 * Method:    _setFilePersist
 * Signature: (ILjava/lang/String;IZZ)V
 */
JNIEXPORT void JNICALL
Java_cc_goiiot_libmqtt_LibMQTT__1setFilePersist
(JNIEnv *env, jclass c, jint id, jstring dir,
 jint max_count, jboolean ex_drop, jboolean dup_replace) {

  const char *c_dir = (*env)->GetStringUTFChars(env, dir, 0);

  Libmqtt_client_with_file_persist(id, c_dir, max_count, ex_drop, dup_replace);

  (*env)->ReleaseStringUTFChars(env, dir, c_dir);
}

/*
 * Method:    _setup
 * Signature: (I)Ljava/lang/String;
 */
JNIEXPORT jstring JNICALL
Java_cc_goiiot_libmqtt_LibMQTT__1setup
(JNIEnv *env, jclass c, jint id) {

  char *err = Libmqtt_setup_client(id);

  if (err != NULL) {
    return (*env)->NewStringUTF(env, err);
  }

  Libmqtt_set_pub_handler(id, &pub_handler);
  Libmqtt_set_sub_handler(id, &sub_handler);
  Libmqtt_set_net_handler(id, &net_handler);
  Libmqtt_set_unsub_handler(id, &unsub_handler);
  Libmqtt_set_persist_handler(id, &persist_handler);

  return NULL;
}
