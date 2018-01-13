# Copyright GoIIoT (https://github.com/goiiot)
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

ANSWER_CA_SSL = -subj "/C=ZH/ST=JiangSu/L=Nanjing/O=GoIIoT/CN=imq.goiiot.cc"
ANSWER_SERVER_SSL = -subj "/C=ZH/ST=JiangSu/L=Nanjing/O=GoIIoT/CN=imq.goiiot.cc"
ANSWER_CLIENT_SSL = -subj "/C=ZH/ST=JiangSu/L=Nanjing/O=GoIIoT/CN=client.goiiot.cc"

CERT_DIR = cred

build:
	go build

clean-certs:
	rm -rf $(CERT_DIR)

gen-certs: gen-ca gen-server-cert gen-client-cert

gen-ca:
	mkdir -p $(CERT_DIR)
	# gen key for ca
	openssl genrsa -out $(CERT_DIR)/ca-key.pem 2048
	# gen ca cert
	openssl req \
		-new \
		-x509 \
		-key $(CERT_DIR)/ca-key.pem \
		-out $(CERT_DIR)/ca-cert.pem \
		$(ANSWER_CA_SSL)

gen-server-cert:
	mkdir -p $(CERT_DIR)
	# gen key for server csr
	openssl genrsa -out $(CERT_DIR)/server-key.pem 2048
	# gen server csr
	openssl req \
		-new \
		-key $(CERT_DIR)/server-key.pem \
		-out $(CERT_DIR)/server.csr \
		$(ANSWER_SERVER_SSL)
	# sign server certificate with csr
	openssl x509 \
		-req \
		-in $(CERT_DIR)/server.csr \
		-CA $(CERT_DIR)/ca-cert.pem \
		-CAkey $(CERT_DIR)/ca-key.pem \
		-CAcreateserial \
		-out $(CERT_DIR)/server-cert.pem

gen-client-cert:
	mkdir -p $(CERT_DIR)
	# gen key for client csr
	openssl genrsa -out $(CERT_DIR)/client-key.pem 2048
	# gen client csr
	openssl req \
		-new \
		-key $(CERT_DIR)/client-key.pem \
		-out $(CERT_DIR)/client.csr \
		$(ANSWER_CLIENT_SSL)
	# sign client certificate with csr
	openssl x509 \
		-req \
		-in $(CERT_DIR)/client.csr \
		-CA $(CERT_DIR)/ca-cert.pem \
		-CAkey $(CERT_DIR)/ca-key.pem \
		-CAcreateserial \
		-out $(CERT_DIR)/client-cert.pem