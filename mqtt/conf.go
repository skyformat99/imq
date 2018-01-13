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
	"strings"
	"time"

	"github.com/goiiot/imq/util"
	"github.com/goiiot/libmqtt"
	"go.uber.org/zap/zapcore"
	"gopkg.in/urfave/cli.v1"
)

// service config
const (
	cfgVersion    = "mqtt-service.mqtt_version"
	cfgCompatible = "mqtt-service.compatible"
	cfgListen     = "mqtt-service.listen"
	cfgTcpPort    = "mqtt-service.tcp"
	cfgTcpsPort   = "mqtt-service.tcps"
	cfgWsPort     = "mqtt-service.ws"
	cfgWssPort    = "mqtt-service.wss"
	cfgTcpMax     = "mqtt-service.max_tcp"
	cfgTcpsMax    = "mqtt-service.max_tcps"
	cfgWsMax      = "mqtt-service.max_ws"
	cfgWssMax     = "mqtt-service.max_wss"
	cfgTlsCert    = "mqtt-service.tls_cert"
	cfgTlsKey     = "mqtt-service.tls_key"
	cfgGraceTime  = "mqtt-service.grace_shutdown_time"
)

// log config
const (
	cfgLogLevel = "mqtt-log.level"
	cfgLogDir   = "mqtt-log.dir"
)

// persist config
const (
	// common persist config
	cfgPersistMethod           = "mqtt-persist.method"
	cfgPersistMaxCount         = "mqtt-persist.max_count"
	cfgPersistDropOnExceed     = "mqtt-persist.drop_on_exceed"
	cfgPersistDuplicateReplace = "mqtt-persist.duplicate_replace"

	// file persist config
	cfgFilePersistInterval = "mqtt-persist.interval"
	cfgFilePersistDir      = "mqtt-persist.file_path"

	// redis persist config
	cfgRedisAddr = "mqtt-persist.redis_addr"
	cfgRedisAuth = "mqtt-persist.redis_auth"
	cfgRedisDB   = "mqtt-persist.redis_db"

	// etcd persist config
	cfgEtcdAddr = "mqtt-persist.etcd_addr"
)

type config struct {
	// service config
	version                            libmqtt.ProtocolVersion
	compatible                         bool
	listen, tlsCertFile, tlsKeyFile    string
	tcpPort, tcpsPort, wsPort, wssPort int
	maxTcp, maxTcps, maxWs, maxWss     int
	graceShutdownTime                  time.Duration

	// log config
	logLevel zapcore.Level
	logDir   string

	// persist common config
	persistMethod           string
	persistMaxCount         int
	persistDropOnExceed     bool
	persistDuplicateReplace bool

	// file persist config
	filePersistInterval time.Duration
	filePersistDir      string

	// redis persist config
	redisAddr string
	redisAuth string
	redisDB   int

	// etcd persist config
	etcdAddr string
}

func Flags() []cli.Flag {
	return []cli.Flag{
		// service config
		util.StringFlag(cfgVersion, "5", ""),
		util.BoolFlag(cfgCompatible, ""),
		util.StringFlag(cfgListen, "0.0.0.0", ""),
		util.IntFlag(cfgTcpPort, 1883, ""),
		util.IntFlag(cfgTcpsPort, 8883, ""),
		util.IntFlag(cfgWsPort, 8083, ""),
		util.IntFlag(cfgWssPort, 18083, ""),
		util.IntFlag(cfgTcpMax, 0, ""),
		util.IntFlag(cfgTcpsMax, 0, ""),
		util.IntFlag(cfgWsMax, 0, ""),
		util.IntFlag(cfgWssMax, 0, ""),
		util.StringFlag(cfgTlsCert, "cred/cert", ""),
		util.StringFlag(cfgTlsKey, "cred/key", ""),
		util.DurationFlag(cfgGraceTime, 10*time.Second, ""),
		// log config
		util.StringFlag(cfgLogLevel, "info", ""),
		util.StringFlag(cfgLogDir, "/var/log/imq/mqtt", ""),
		// persist config
		util.StringFlag(cfgPersistMethod, "none", ""),
		util.IntFlag(cfgPersistMaxCount, 1000, ""),
		util.BoolFlag(cfgPersistDropOnExceed, ""),
		util.BoolFlag(cfgPersistDuplicateReplace, ""),
		// file persist config
		util.DurationFlag(cfgFilePersistInterval, time.Minute, ""),
		util.StringFlag(cfgFilePersistDir, "", ""),
		// redis persist config
		util.StringFlag(cfgRedisAddr, "", ""),
		util.StringFlag(cfgRedisAuth, "", ""),
		util.IntFlag(cfgRedisDB, 0, ""),
		// etcd persist config
		util.StringFlag(cfgEtcdAddr, "", ""),
	}
}

func getConfig(ctx *cli.Context) *config {
	return &config{
		// service config
		version: func() libmqtt.ProtocolVersion {
			switch ctx.String(cfgVersion) {
			case "3.1.1":
				return libmqtt.V311
			case "5":
				return libmqtt.V5
			default:
				panic("not supported mqtt version: " + ctx.String(cfgVersion))
			}
		}(),
		compatible:        ctx.Bool(cfgCompatible),
		listen:            ctx.String(cfgListen),
		tcpPort:           ctx.Int(cfgTcpPort),
		tcpsPort:          ctx.Int(cfgTcpsPort),
		wsPort:            ctx.Int(cfgWsPort),
		wssPort:           ctx.Int(cfgWssPort),
		maxTcp:            ctx.Int(cfgTcpMax),
		maxTcps:           ctx.Int(cfgTcpsMax),
		maxWs:             ctx.Int(cfgWsMax),
		maxWss:            ctx.Int(cfgWssMax),
		tlsCertFile:       ctx.String(cfgTlsCert),
		tlsKeyFile:        ctx.String(cfgTlsKey),
		graceShutdownTime: ctx.Duration(cfgGraceTime),
		// log config
		logDir: ctx.String(cfgLogDir),
		logLevel: func() zapcore.Level {
			switch strings.ToLower(ctx.String(cfgLogLevel)) {
			case "debug":
				return zapcore.DebugLevel
			case "info":
				return zapcore.InfoLevel
			case "warn":
				return zapcore.WarnLevel
			case "error":
				return zapcore.ErrorLevel
			case "panic":
				return zapcore.PanicLevel
			case "fatal":
				return zapcore.FatalLevel
			default:
				panic("not supported mqtt log level: " + ctx.String(cfgLogLevel))
			}
		}(),
		// persist common config
		persistMethod:           ctx.String(cfgPersistMethod),
		persistMaxCount:         ctx.Int(cfgPersistMaxCount),
		persistDropOnExceed:     ctx.Bool(cfgPersistDropOnExceed),
		persistDuplicateReplace: ctx.Bool(cfgPersistDuplicateReplace),
		// file persist config
		filePersistInterval: ctx.Duration(cfgFilePersistInterval),
		filePersistDir:      ctx.String(cfgFilePersistDir),
		// redis persist config
		redisAddr: ctx.String(cfgRedisAddr),
		redisAuth: ctx.String(cfgRedisAuth),
		redisDB:   ctx.Int(cfgRedisDB),
		// etcd persist config
		etcdAddr: ctx.String(cfgEtcdAddr),
	}
}
