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

static JavaVM *jvm;

static jclass libmqtt_class;

static jmethodID on_conn_msg_mid;
static jmethodID on_sub_msg_mid;
static jmethodID on_pub_msg_mid;
static jmethodID on_unsub_msg_mid;
static jmethodID on_net_msg_mid;
static jmethodID on_persist_err_mid;
static jmethodID on_topic_msg_mid;

void conn_handler(int client, char *server, libmqtt_connack_t code, char *err) {
  JNIEnv *g_env;
  (*jvm)->AttachCurrentThread(jvm, (void **)&g_env, NULL);

  if (g_env == NULL || libmqtt_class == NULL || on_conn_msg_mid == 0) {
    return;
  } else {
    (*g_env)->CallStaticVoidMethod(g_env, libmqtt_class,
                                   on_conn_msg_mid, client,
                                   0,
                                   (*g_env)->NewStringUTF(g_env, err));
  }
}

void sub_handler(int client, char *topic, int qos, char *err) {
  JNIEnv *g_env;
  (*jvm)->AttachCurrentThread(jvm, (void **)&g_env, NULL);

  if (g_env == NULL || libmqtt_class == NULL || on_sub_msg_mid == 0) {
    return;
  } else {
    (*g_env)->CallStaticVoidMethod(g_env, libmqtt_class,
                                   on_sub_msg_mid, client,
                                   (*g_env)->NewStringUTF(g_env, topic),
                                   qos,
                                   (*g_env)->NewStringUTF(g_env, err));
  }
}

void pub_handler(int client, char *topic, char *err) {
  JNIEnv *g_env;
  (*jvm)->AttachCurrentThread(jvm, (void **)&g_env, NULL);

  if (g_env == NULL || libmqtt_class == NULL || on_pub_msg_mid == 0) {
    return;
  } else {
    (*g_env)->CallStaticVoidMethod(g_env, libmqtt_class,
                                   on_pub_msg_mid, client,
                                   (*g_env)->NewStringUTF(g_env, topic),
                                   (*g_env)->NewStringUTF(g_env, err));
  }
}

void unsub_handler(int client, char *topic, char *err) {
  JNIEnv *g_env;
  (*jvm)->AttachCurrentThread(jvm, (void **)&g_env, NULL);

  if (g_env == NULL || libmqtt_class == NULL || on_unsub_msg_mid == 0) {
    return;
  } else {
    (*g_env)->CallStaticVoidMethod(g_env, libmqtt_class,
                                   on_unsub_msg_mid, client,
                                   (*g_env)->NewStringUTF(g_env, topic),
                                   (*g_env)->NewStringUTF(g_env, err));
  }
}

void net_handler(int client, char *server, char *err) {
  JNIEnv *g_env;
  (*jvm)->AttachCurrentThread(jvm, (void **)&g_env, NULL);

  if (g_env == NULL || libmqtt_class == NULL || on_net_msg_mid == 0) {
    return;
  } else {
    (*g_env)->CallStaticVoidMethod(g_env, libmqtt_class,
                                   on_net_msg_mid, client,
                                   (*g_env)->NewStringUTF(g_env, err));
  }
}

void persist_handler(int client, char *err) {
  JNIEnv *g_env;
  (*jvm)->AttachCurrentThread(jvm, (void **)&g_env, NULL);

  if (g_env == NULL || libmqtt_class == NULL || on_topic_msg_mid == 0) {
    return;
  } else {
    (*g_env)->CallStaticVoidMethod(g_env, libmqtt_class,
                                   on_persist_err_mid, client,
                                   (*g_env)->NewStringUTF(g_env, err));
  }
}

void topic_handler(int client, char *topic, int qos, char *payload, int size) {
  JNIEnv *g_env;
  (*jvm)->AttachCurrentThread(jvm, (void **)&g_env, NULL);

  if (g_env == NULL || libmqtt_class == NULL || on_topic_msg_mid == 0) {
    return;
  } else {
    jbyteArray result = (*g_env)->NewByteArray(g_env, size);
    (*g_env)->SetByteArrayRegion(g_env, result, 0, size, (jbyte *)payload);

    (*g_env)->CallStaticVoidMethod(g_env, libmqtt_class,
                                   on_topic_msg_mid, client,
                                   (*g_env)->NewStringUTF(g_env, topic),
                                   qos, result);
  }
}

/*
 * Method:    _init
 * Signature: ()V
 */
JNIEXPORT void JNICALL
Java_cc_goiiot_libmqtt_LibMQTT__1init
(JNIEnv *env, jclass c) {

  (*env)->GetJavaVM(env, &jvm);

  libmqtt_class = (jclass)(*env)->NewGlobalRef(
                    env, (*env)->FindClass(env,
                        "cc/goiiot/libmqtt/LibMQTT"));;

  on_conn_msg_mid = (*env)->GetStaticMethodID(
                      env, c, "onConnMessage",
                      "(IILjava/lang/String;)V");

  on_net_msg_mid = (*env)->GetStaticMethodID(
                     env, c, "onNetMessage",
                     "(ILjava/lang/String;)V");

  on_pub_msg_mid = (*env)->GetStaticMethodID(
                     env, c, "onPubMessage",
                     "(ILjava/lang/String;Ljava/lang/String;)V");

  on_sub_msg_mid = (*env)->GetStaticMethodID(
                     env, c, "onSubMessage",
                     "(ILjava/lang/String;ILjava/lang/String;)V");

  on_unsub_msg_mid = (*env)->GetStaticMethodID(
                       env, c, "onUnsubMessage",
                       "(ILjava/lang/String;Ljava/lang/String;)V");

  on_persist_err_mid = (*env)->GetStaticMethodID(
                         env, c, "onPersistError",
                         "(ILjava/lang/String;)V");

  on_topic_msg_mid = (*env)->GetStaticMethodID(
                       env, c, "onTopicMessage",
                       "(ILjava/lang/String;I[B)V");
}