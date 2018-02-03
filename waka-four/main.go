package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/AsynkronIT/protoactor-go/actor"
	protolog "github.com/AsynkronIT/protoactor-go/log"
	"github.com/davyxu/cellnet"
	"github.com/davyxu/golog"
	"github.com/sirupsen/logrus"

	_ "github.com/liuhan907/waka/waka-four/log"
	_ "github.com/liuhan907/waka/waka/vt100"

	"github.com/liuhan907/waka/waka-four/backend"
	"github.com/liuhan907/waka/waka-four/conf"
	"github.com/liuhan907/waka/waka-four/modules/hall"
	"github.com/liuhan907/waka/waka-four/modules/player"
	"github.com/liuhan907/waka/waka/modules/gateway"
	"github.com/liuhan907/waka/waka/modules/session"
	"github.com/liuhan907/waka/waka/modules/supervisor"
)

var (
	log = logrus.WithFields(logrus.Fields{
		"pid":    os.Getpid(),
		"module": "main",
	})
)

func init() {
	logrus.SetLevel(logrus.Level(conf.Option.Log.LogLevel))
	golog.SetLevelByString("*", "fatal")
	actor.SetLogLevel(protolog.OffLevel)
}

func main() {
	startGateway()
	wait()
}

func startGateway() {
	supervisorTargetCreator := func(pid *actor.PID) *actor.PID {
		target := hall.Spawn(pid)
		go func() {
			backendOption := backend.Option{
				TargetCreator: func() *actor.PID {
					return target
				},
				Address: conf.Option.Backend.Listen4,
			}
			backend.Start(backendOption)
		}()
		return target
	}
	supervisorOption := supervisor.Option{
		TargetCreator: supervisorTargetCreator,
		EnableLog:     false,
	}
	supervisorHall := supervisor.Spawn("four", supervisorOption)

	sessionTargetCreator := func(remote string, pid *actor.PID) *actor.PID {
		return player.Spawn(supervisorHall, remote, pid)
	}
	sessionOption := session.Option{
		TargetCreator:   sessionTargetCreator,
		EnableHeart:     true,
		EnableLog:       false,
		EnableHeartLog:  false,
		HeartPeriod:     time.Second * 3,
		HeartDeadPeriod: time.Second * 3 * 10,
	}

	gatewayTargetCreator := func(conn cellnet.Session) *actor.PID {
		return session.Spawn(sessionOption, conn)
	}
	gatewayOption := gateway.Option{
		TargetCreator: gatewayTargetCreator,
		Address:       conf.Option.Gateway.Listen4,
	}
	gateway.Start(gatewayOption)
}

func wait() {
	if runtime.GOOS == "linux" {
		if pid := syscall.Getpid(); pid != 0 {
			name := "kill.sh"
			script := fmt.Sprintf("kill %v", pid)

			if err := ioutil.WriteFile(name, []byte(script), 0777); err != nil {
				log.WithFields(logrus.Fields{
					"err": err,
				}).Errorln("write kill script failed")
			}
			if err := os.Chmod(name, 777); err != nil {
				log.WithFields(logrus.Fields{
					"err": err,
				}).Errorln("chmod kill script failed")
			}

			defer os.Remove(name)
		}
	}

	c := make(chan os.Signal, 0)
	signal.Notify(c, syscall.SIGINT, syscall.SIGKILL, syscall.SIGTERM)
	<-c
	log.Infoln("exit signal received")
}
