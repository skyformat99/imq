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

package util

import (
	"time"

	"github.com/urfave/cli/altsrc"
	"gopkg.in/urfave/cli.v1"
)

func StringFlag(name, value, usage string) cli.Flag {
	return altsrc.NewStringFlag(cli.StringFlag{
		Name:  name,
		Value: value,
		Usage: usage,
	})
}

func BoolFlag(name string, usage string) cli.Flag {
	return altsrc.NewBoolFlag(cli.BoolFlag{
		Name:  name,
		Usage: usage,
	})
}

func IntFlag(name string, value int, usage string) cli.Flag {
	return altsrc.NewIntFlag(cli.IntFlag{
		Name:  name,
		Value: value,
		Usage: usage,
	})
}

func DurationFlag(name string, value time.Duration, usage string) cli.Flag {
	return altsrc.NewDurationFlag(cli.DurationFlag{
		Name:  name,
		Value: value,
		Usage: usage,
	})
}
