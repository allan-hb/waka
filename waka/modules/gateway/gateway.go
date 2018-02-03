package gateway

import (
	"os"

	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/davyxu/cellnet"
	"github.com/davyxu/cellnet/socket"
	"github.com/liuhan907/waka/waka/proto"
	"github.com/sirupsen/logrus"

	"github.com/liuhan907/waka/waka/modules/gateway/gateway_message"
	"github.com/liuhan907/waka/waka/modules/gateway/tlv"
)

var (
	log = logrus.WithFields(logrus.Fields{
		"pid":    os.Getpid(),
		"module": "waka.gateway",
	})
)

// 消息转发目标创建器
type TargetCreator func(conn cellnet.Session) *actor.PID

// 配置
type Option struct {
	// 创建器
	TargetCreator TargetCreator

	// 监听地址
	Address string
}

// 启动
func Start(option Option) {
	peer := socket.NewAcceptor(nil)

	peer.SetReadWriteChain(func() *cellnet.HandlerChain {
		return cellnet.NewHandlerChain(
			cellnet.NewFixedLengthFrameReader(8),
			tlv.NewReader(),
		)
	}, func() *cellnet.HandlerChain {
		return cellnet.NewHandlerChain(
			tlv.NewWriter(),
			cellnet.NewFixedLengthFrameWriter(),
		)
	})

	peer.Start(option.Address)

	cellnet.RegisterMessage(peer, "coredef.SessionAccepted",
		func(ev *cellnet.Event) {
			ev.Ses.SetTag(option.TargetCreator(ev.Ses))
		})
	cellnet.RegisterMessage(peer, "coredef.SessionClosed",
		func(ev *cellnet.Event) {
			pid := ev.Ses.Tag().(*actor.PID)
			pid.Tell(&gateway_message.Closed{})
		})
	cellnet.RegisterMessage(peer, "waka_proto.Heart",
		func(ev *cellnet.Event) {
			pid := ev.Ses.Tag().(*actor.PID)
			pid.Tell(&gateway_message.Heart{})
		})
	cellnet.RegisterMessage(peer, "waka_proto.Transport",
		func(ev *cellnet.Event) {
			evd := ev.Msg.(*waka_proto.Transport)
			pid := ev.Ses.Tag().(*actor.PID)
			pid.Tell(&gateway_message.Transport{evd.GetId(), evd.GetPayload()})
		})
	cellnet.RegisterMessage(peer, "waka_proto.FutureRequest",
		func(ev *cellnet.Event) {
			evd := ev.Msg.(*waka_proto.FutureRequest)
			pid := ev.Ses.Tag().(*actor.PID)
			pid.Tell(&gateway_message.FutureRequest{evd.GetId(), evd.GetPayload(), evd.GetNumber()})
		})

	log.WithFields(logrus.Fields{
		"address": option.Address,
	}).Infoln("listen started")
}
