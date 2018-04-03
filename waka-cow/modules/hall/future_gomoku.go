package hall

import (
	"github.com/golang/protobuf/proto"
	"github.com/sirupsen/logrus"
	. "gopkg.in/ahmetb/go-linq.v3"

	"github.com/liuhan907/waka/waka-cow/database"
	"github.com/liuhan907/waka/waka-cow/proto"
	"github.com/liuhan907/waka/waka/modules/supervisor/supervisor_message"
)

func (my *actorT) playerFutureRequestedGomoku(player *playerT, ev *supervisor_message.PlayerFutureRequested) bool {
	switch evd := ev.Payload.(type) {
	case *cow_proto.GomokuGetHistoryRequest:
		my.GomokuGetHistoryRequest(player, evd, ev.Respond)
	default:
		return false
	}
	return true
}

func (my *actorT) GomokuGetHistoryRequest(player *playerT,
	ev *cow_proto.GomokuGetHistoryRequest,
	respond func(proto.Message, error)) {

	histories, err := database.GomokuQueryHistory(player.Player, 20)
	if err != nil {
		log.WithFields(logrus.Fields{
			"err": err,
		}).Warnln("query gomoku war history failed")
		respond(nil, err)
		return
	}

	var d []*cow_proto.GomokuHistory
	From(histories).SelectT(func(x *database.GomokuHistory) *cow_proto.GomokuHistory {
		return &cow_proto.GomokuHistory{
			PlayerId:  int32(x.Player),
			Cost:      x.Cost / 100,
			CreatedAt: x.CreatedAt.Format("2006-01-02 15:04:05"),
		}
	}).ToSlice(&d)

	respond(&cow_proto.GomokuGetHistoryResponse{d}, nil)
}
