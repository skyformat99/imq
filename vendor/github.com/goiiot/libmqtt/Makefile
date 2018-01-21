# Copyright Go-IIoT (https://github.com/goiiot)
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

.PHONY: test lib client clean fuzz-test

test:
	go test -v -run=. -count=1 -race -coverprofile=coverage.txt -covermode=atomic -timeout 20m

.PHONY: all-lib c-lib java-lib py-lib \
		clean-c-lib clean-java-lib clean-py-lib

all-lib: c-lib java-lib

clean-all-lib: clean-c-lib clean-java-lib

c-lib:
	$(MAKE) -C c lib

clean-c-lib:
	$(MAKE) -C c clean

java-lib:
	$(MAKE) -C java build

clean-java-lib:
	$(MAKE) -C java clean

client:
	$(MAKE) -C cmd build

clean: clean-all-lib fuzz-clean
	rm -rf coverage.txt

fuzz-test:
	go-fuzz-build github.com/goiiot/libmqtt
	go-fuzz -bin=./libmqtt-fuzz.zip -workdir=fuzz-test

fuzz-clean:
	rm -rf fuzz-test libmqtt-fuzz.zip
