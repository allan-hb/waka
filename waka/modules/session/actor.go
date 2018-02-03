package session

import (
	"os"
	"time"

	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/davyxu/cellnet"
	"github.com/sirupsen/logrus"
)

var (
	log = logrus.WithFields(logrus.Fields{
		"pid":    os.Getpid(),
		"module": "waka.session",
	})
)

type actorT struct {
	option Option
	conn   cellnet.Session

	log    *logrus.Entry
	pid    *actor.PID
	target *actor.PID

	heart time.Time
}

func (my *actorT) Receive(context actor.Context) {
	if my.ReceiveActor(context) {
		return
	}
	if my.ReceiveGateway(context) {
		return
	}
	if my.ReceiveTarget(context) {
		return
	}
	if my.ReceiveClock(context) {
		return
	}
}

// 消息转发目标创建器
type TargetCreator func(remote string, pid *actor.PID) *actor.PID

// 会话配置
type Option struct {
	// 消息接收者创建者
	TargetCreator TargetCreator

	// 启用日志
	EnableLog bool
	// 启用心跳
	EnableHeart bool
	// 启用心跳日志
	EnableHeartLog bool

	// 心跳周期
	HeartPeriod time.Duration
	// 死亡时长
	HeartDeadPeriod time.Duration
}

// 创建会话
func Spawn(option Option, conn cellnet.Session) *actor.PID {
	return actor.Spawn(
		actor.FromInstance(
			&actorT{
				option: option,
				conn:   conn,
			},
		),
	)
}
