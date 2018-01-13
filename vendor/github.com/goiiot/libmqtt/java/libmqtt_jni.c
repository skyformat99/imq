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
 * Method:    _handle
 * Signature: (ILjava/lang/String;Lcc/goiiot/libmqtt/LibMQTT/TopicMessageCallback;)V
 */
JNIEXPORT void JNICALL
Java_cc_goiiot_libmqtt_LibMQTT__1handle
(JNIEnv *env, jclass c, jint id, jstring topic, jobject callback) {

  if (callback == NULL) {
    return;
  }

  const char *c_topic = (*env)->GetStringUTFChars(env, topic, 0);

  Libmqtt_handle(id, c_topic, &topic_handler);

  (*env)->ReleaseStringUTFChars(env, topic, c_topic);
}

/*
 * Method:    _conn
 * Signature: (I)V
 */
JNIEXPORT void JNICALL
Java_cc_goiiot_libmqtt_LibMQTT__1connect
(JNIEnv *env, jclass c, jint id) {

  Libmqtt_connect(id, &conn_handler);
}

/*
 * Method:    _wait
 * Signature: (I)V
 */
JNIEXPORT void JNICALL
Java_cc_goiiot_libmqtt_LibMQTT__1wait
(JNIEnv *env, jclass c, jint id) {

  (*env)->MonitorEnter(env, c);
  Libmqtt_wait(id);
  (*env)->MonitorExit(env, c);
}

/*
 * Method:    _pub
 * Signature: (ILjava/lang/String;I[B)V
 */
JNIEXPORT void JNICALL
Java_cc_goiiot_libmqtt_LibMQTT__1pub
(JNIEnv *env, jclass c, jint id, jstring topic, jint qos, jbyteArray payload) {

  const char *c_topic = (*env)->GetStringUTFChars(env, topic, 0);
  jsize len = (*env)->GetArrayLength(env, payload);
  jbyte *body = (*env)->GetByteArrayElements(env, payload, 0);

  Libmqtt_publish(id, c_topic, qos, body, len);

  (*env)->ReleaseStringUTFChars(env, topic, c_topic);
  (*env)->ReleaseByteArrayElements(env, payload, body, 0);
}

/*
 * Method:    _sub
 * Signature: (ILjava/lang/String;I)V
 */
JNIEXPORT void JNICALL
Java_cc_goiiot_libmqtt_LibMQTT__1sub
(JNIEnv *env, jclass c, jint id, jstring topic, jint qos) {

  const char *c_topic = (*env)->GetStringUTFChars(env, topic, 0);

  Libmqtt_subscribe(id, c_topic, qos);

  (*env)->ReleaseStringUTFChars(env, topic, c_topic);
}

/*
 * Method:    _unsub
 * Signature: (ILjava/lang/String;)V
 */
JNIEXPORT void JNICALL
Java_cc_goiiot_libmqtt_LibMQTT__1unsub
(JNIEnv *env, jclass c, jint id, jstring topic) {

  const char *c_topic = (*env)->GetStringUTFChars(env, topic, 0);

  Libmqtt_unsubscribe(id, c_topic);

  (*env)->ReleaseStringUTFChars(env, topic, c_topic);
}

/*
 * Method:    _destroy
 * Signature: (IZ)V
 */
JNIEXPORT void JNICALL
Java_cc_goiiot_libmqtt_LibMQTT__1destroy
(JNIEnv *env, jclass c, jint id, jboolean force) {

  Libmqtt_destroy(id, force);
}