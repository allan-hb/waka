package logger

import (
	"bytes"

	"github.com/AsynkronIT/protoactor-go/actor"
)

type actorT struct {
	option Option

	name string
	pid  *actor.PID

	w *bytes.Buffer
}

func (my *actorT) Receive(context actor.Context) {
	if my.ReceiveActor(context) {
		return
	}
	if my.ReceiveClock(context) {
		return
	}
	if my.ReceiveLog(context) {
		return
	}
}

// 会话配置
type Option struct {
	// 日志文件名前缀
	Prefix string
}

// 创建会话
func Spawn(option Option) *actor.PID {
	return actor.Spawn(
		actor.FromInstance(
			&actorT{
				option: option,
				w:      bytes.NewBuffer(make([]byte, 0, 1024*4*64)),
			},
		),
	)
}
