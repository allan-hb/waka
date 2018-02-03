package supervisor

import (
	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/sirupsen/logrus"
)

type actorT struct {
	name   string
	option Option

	log    *logrus.Entry
	pid    *actor.PID
	target *actor.PID

	players map[uint64]*actor.PID
}

func (my *actorT) Receive(context actor.Context) {
	if my.ReceiveActor(context) {
		return
	}
	if my.ReceivePlayer(context) {
		return
	}
	if my.ReceiveHall(context) {
		return
	}
}

// 消息转发目标创建器
type TargetCreator func(pid *actor.PID) *actor.PID

// 会话配置
type Option struct {
	// 消息接收者创建者
	TargetCreator TargetCreator

	// 启用日志
	EnableLog bool
}

func Spawn(name string, option Option) *actor.PID {
	return actor.Spawn(
		actor.FromInstance(
			&actorT{
				name:    name,
				option:  option,
				players: make(map[uint64]*actor.PID, 12800),
			},
		),
	)
}
