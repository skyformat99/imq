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

import (
	"os"
	"os/signal"

	"github.com/goiiot/imq/mqtt"
	"github.com/urfave/cli/altsrc"
	"golang.org/x/net/context"
	"gopkg.in/urfave/cli.v1"
)

const (
	name    = "iMQ"
	desc    = "Message Queue for IoT"
	version = "0.0.1 alpha"
	author  = "GoIIoT Development Group"
)

func main() {
	app := cli.NewApp()
	app.Name = name
	app.Description = desc
	app.Version = version
	app.Author = author
	app.Action = start
	app.Flags = flags()
	// parse toml config file
	app.Before = altsrc.InitInputSourceWithContext(app.Flags,
		altsrc.NewTomlSourceFromFlagFunc("config"))
	app.Run(os.Args)
}

func flags() []cli.Flag {
	flags := []cli.Flag{
		&cli.StringFlag{
			Name:  "config, c",
			Value: "config.ini",
		},
	}
	flags = append(flags, mqtt.Flags()...)
	return flags
}

func start(c *cli.Context) error {
	exitCtx, exit := context.WithCancel(context.Background())

	sigCh := make(chan os.Signal)
	signal.Notify(sigCh, os.Interrupt, os.Kill)
	go func() {
		<-sigCh
		exit()
	}()

	mqtt.Init(exitCtx, c)
	return nil
}
