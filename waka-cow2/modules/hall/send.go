package hall

import (
	"reflect"

	"github.com/golang/protobuf/proto"
	"github.com/sirupsen/logrus"

	"github.com/liuhan907/waka/waka-cow2/conf"
	"github.com/liuhan907/waka/waka-cow2/database"
	"github.com/liuhan907/waka/waka-cow2/proto"
	"github.com/liuhan907/waka/waka/modules/supervisor/supervisor_message"
)

func (my *actorT) send(player database.Player, m proto.Message) {
	if playerData, being := my.players[player]; !being || playerData.Remote == "" {
		return
	}

	log.WithFields(logrus.Fields{
		"player":  player,
		"type":    reflect.TypeOf(m).Elem().Name(),
		"payload": m.String(),
	}).Debugln("send")

	my.supervisor.Tell(&supervisor_message.SendFromHall{uint64(player), m})
}

// ---------------------------------------------------------------------------------------------------------------------

func (my *actorT) sendPlayer(player database.Player) {
	my.send(player, my.ToPlayer(player))
}

func (my *actorT) sendPlayerSecret(player database.Player) {
	my.send(player, my.ToPlayerSecret(player))
}

func (my *actorT) sendHallEntered(player database.Player) {
	my.send(player, &cow_proto.HallEntered{
		Player: my.ToPlayerSecret(player),
	})
}

func (my *actorT) sendWelcome(player database.Player) {
	my.send(player, &cow_proto.Welcome{
		Customers:   database.GetCustomerServices(),
		Notice:      database.GetNotice(),
		NoticeBig:   database.GetNoticeBig(),
		PayUrl:      database.GetPayURL(),
		RegisterUrl: database.GetRegisterURL(),
		LoginUrl:    database.GetLoginURL(),
	})
}

func (my *actorT) sendPlayerNumber(player database.Player, number int32) {
	my.send(player, &cow_proto.PlayerNumber{
		Number: number + conf.Option.Hall.MinPlayerNumber,
	})
}

func (my *actorT) sendRecover(player database.Player, is bool, name string) {
	my.send(player, &cow_proto.Recover{
		Is:   is,
		Name: name,
	})
}

// ---------------------------------------------------------------------------------------------------------------------

func (my *actorT) sendNiuniuCreateRoomFailed(player database.Player, reason int32) {
	my.send(player, &cow_proto.NiuniuCreateRoomFailed{reason})
}

func (my *actorT) sendNiuniuJoinRoomFailed(player database.Player, reason int32) {
	my.send(player, &cow_proto.NiuniuJoinRoomFailed{reason})
}

func (my *actorT) sendNiuniuRoomJoined(player database.Player, room cowRoomT) {
	my.send(player, &cow_proto.NiuniuRoomJoined{room.NiuniuRoomData1()})
}

func (my *actorT) sendNiuniuRoomCreated(player database.Player, id int32) {
	my.send(player, &cow_proto.NiuniuRoomCreated{id})
}

func (my *actorT) sendNiuniuRoomLeft(player database.Player) {
	my.send(player, &cow_proto.NiuniuRoomLeft{})
}

func (my *actorT) sendNiuniuRoomLeftByDismiss(player database.Player) {
	my.send(player, &cow_proto.NiuniuRoomLeftByDismiss{})
}

func (my *actorT) sendNiuniuUpdateRoom(player database.Player, room cowRoomT) {
	my.send(player, &cow_proto.NiuniuUpdateRoom{room.NiuniuRoomData1()})
}

func (my *actorT) sendNiuniuUpdateRound(player database.Player, room cowRoomT) {
	my.send(player, &cow_proto.NiuniuUpdateRound{room.NiuniuRoundStatus(player)})
}

func (my *actorT) sendNiuniuCountdown(player database.Player, number int32) {
	my.send(player, &cow_proto.NiuniuCountdown{number})
}

func (my *actorT) sendNiuniuStarted(player database.Player, number int32) {
	my.send(player, &cow_proto.NiuniuStarted{number})
}

func (my *actorT) sendNiuniuDeal4(player database.Player, pokers []string) {
	my.send(player, &cow_proto.NiuniuDeal4{pokers})
}

func (my *actorT) sendNiuniuRequireGrab(player database.Player) {
	my.send(player, &cow_proto.NiuniuRequireGrab{})
}

func (my *actorT) sendNiuniuRequireSpecifyBanker(player database.Player, is bool) {
	my.send(player, &cow_proto.NiuniuRequireSpecifyBanker{is})
}

func (my *actorT) sendNiuniuGrabAnimation(player database.Player, room cowRoomT) {
	my.send(player, room.NiuniuGrabAnimation())
}

func (my *actorT) sendNiuniuRequireSpecifyRate(player database.Player, is bool) {
	my.send(player, &cow_proto.NiuniuRequireSpecifyRate{is})
}

func (my *actorT) sendNiuniuRoundClear(player database.Player, room cowRoomT) {
	my.send(player, room.NiuniuRoundClear())
}

func (my *actorT) sendNiuniuRoundFinally(player database.Player, room cowRoomT) {
	my.send(player, room.NiuniuRoundFinally())
}

func (my *actorT) sendNiuniuRequireCommitConfirm(player database.Player) {
	my.send(player, &cow_proto.NiuniuRequireCommitConfirm{})
}

func (my *actorT) sendNiuniuRoomMessage(player database.Player, sender database.Player, content string) {
	my.send(player, &cow_proto.NiuniuRoomMessage{
		Sender:  int32(sender),
		Content: content,
	})
}

// ----------------------------------------------------

func (my *actorT) sendNiuniuUpdateRoomForAll(room cowRoomT) {
	for _, player := range room.GetPlayers() {
		my.sendNiuniuUpdateRoom(player, room)
	}
}

func (my *actorT) sendNiuniuUpdateRoundForAll(room cowRoomT) {
	for _, player := range room.GetPlayers() {
		my.sendNiuniuUpdateRound(player, room)
	}
}

func (my *actorT) sendNiuniuCountdownForAll(room cowRoomT, number int32) {
	for _, player := range room.GetPlayers() {
		my.sendNiuniuCountdown(player, number)
	}
}

func (my *actorT) sendNiuniuStartedForAll(room cowRoomT, number int32) {
	for _, player := range room.GetPlayers() {
		my.sendNiuniuStarted(player, number)
	}
}

// ---------------------------------------------------------------------------------------------------------------------
