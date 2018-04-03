package hall

import (
	"reflect"

	"github.com/golang/protobuf/proto"
	"github.com/sirupsen/logrus"

	"github.com/liuhan907/waka/waka-cow/database"
	"github.com/liuhan907/waka/waka-cow/proto"
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

func (my *actorT) sendHallEntered(player database.Player) {
	my.send(player, &cow_proto.HallEntered{})
}

func (my *actorT) sendPlayerNumber(player database.Player, number int32) {
	my.send(player, &cow_proto.PlayerNumber{
		Number: number,
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

func (my *actorT) sendNiuniuCreateRoomSuccess(player database.Player, id int32) {
	my.send(player, &cow_proto.NiuniuCreateRoomSuccess{id})
}

func (my *actorT) sendNiuniuJoinRoomFailed(player database.Player, reason int32) {
	my.send(player, &cow_proto.NiuniuJoinRoomFailed{reason})
}

func (my *actorT) sendNiuniuJoinRoomSuccess(player database.Player) {
	my.send(player, &cow_proto.NiuniuJoinRoomSuccess{})
}

func (my *actorT) sendNiuniuLeftRoom(player database.Player, reason int32) {
	my.send(player, &cow_proto.NiuniuLeftRoom{reason})
}

func (my *actorT) sendNiuniuUpdateRoom(player database.Player, room cowRoom) {
	my.send(player, &cow_proto.NiuniuUpdateRoom{room.NiuniuRoomData()})
}

func (my *actorT) sendNiuniuGameStarted(player database.Player) {
	my.send(player, &cow_proto.NiuniuGameStarted{})
}

func (my *actorT) sendNiuniuRoundStarted(player database.Player, number int32) {
	my.send(player, &cow_proto.NiuniuRoundStarted{number})
}

func (my *actorT) sendNiuniuDeadline(player database.Player, deadline int64) {
	my.send(player, &cow_proto.NiuniuDeadline{deadline})
}

func (my *actorT) sendNiuniuUpdateRound(player database.Player, room cowRoom) {
	my.send(player, &cow_proto.NiuniuUpdateRound{room.NiuniuRoundStatus(player)})
}

func (my *actorT) sendNiuniuRequireSpecifyBanker(player database.Player, is bool) {
	my.send(player, &cow_proto.NiuniuRequireSpecifyBanker{is})
}

func (my *actorT) sendNiuniuRequireGrab(player database.Player) {
	my.send(player, &cow_proto.NiuniuRequireGrab{})
}

func (my *actorT) sendNiuniuRequireGrabShow(player database.Player, room cowRoom) {
	my.send(player, room.NiuniuRequireGrabShow())
}

func (my *actorT) sendNiuniuDeal4(player database.Player, pokers []string) {
	my.send(player, &cow_proto.NiuniuDeal4{pokers})
}

func (my *actorT) sendNiuniuRequireSpecifyRate(player database.Player, is bool) {
	my.send(player, &cow_proto.NiuniuRequireSpecifyRate{is})
}

func (my *actorT) sendNiuniuDeal1(player database.Player, poker, pokersType string, pokers []string) {
	my.send(player, &cow_proto.NiuniuDeal1{
		poker,
		pokersType,
		pokers,
	})
}

func (my *actorT) sendNiuniuRoundClear(player database.Player, room cowRoom) {
	my.send(player, room.NiuniuRoundClear())
}

func (my *actorT) sendNiuniuGameFinally(player database.Player, room cowRoom) {
	my.send(player, room.NiuniuGameFinally())
}

// ----------------------------------------------------

func (my *actorT) sendNiuniuUpdateRoomForAll(room cowRoom) {
	for _, player := range room.GetPlayers() {
		my.sendNiuniuUpdateRoom(player, room)
	}
}

func (my *actorT) sendNiuniuGameStartedForAll(room cowRoom) {
	for _, player := range room.GetPlayers() {
		my.sendNiuniuGameStarted(player)
	}
}

func (my *actorT) sendNiuniuRoundStartedForAll(room cowRoom, number int32) {
	for _, player := range room.GetPlayers() {
		my.sendNiuniuRoundStarted(player, number)
	}
}

func (my *actorT) sendNiuniuUpdateRoundForAll(room cowRoom) {
	for _, player := range room.GetPlayers() {
		my.sendNiuniuUpdateRound(player, room)
	}
}

func (my *actorT) sendNiuniuDeadlineForAll(room cowRoom, deadline int64) {
	for _, player := range room.GetPlayers() {
		my.sendNiuniuDeadline(player, deadline)
	}
}

// ---------------------------------------------------------------------------------------------------------------------

func (my *actorT) sendRedCreateBagFailed(player database.Player, reason int32) {
	my.send(player, &cow_proto.RedCreateBagFailed{reason})
}

func (my *actorT) sendRedCreateBagSuccess(player database.Player, id int32) {
	my.send(player, &cow_proto.RedCreateBagSuccess{id})
}

func (my *actorT) sendRedUpdateBagList(player database.Player, bags redBagMapT) {
	my.send(player, &cow_proto.RedUpdateBagList{bags.RedBag(player)})
}

func (my *actorT) sendRedGrabFailed(player database.Player, reason int32) {
	my.send(player, &cow_proto.RedGrabFailed{reason})
}

func (my *actorT) sendRedGrabSuccess(player database.Player) {
	my.send(player, &cow_proto.RedGrabSuccess{})
}

func (my *actorT) sendRedDeadline(player database.Player, deadline int64) {
	my.send(player, &cow_proto.RedDeadline{deadline})
}

func (my *actorT) sendRedUpdateBag(player database.Player, bag *redBagT) {
	my.send(player, &cow_proto.RedUpdateBag{bag.RedBag()})
}

func (my *actorT) sendRedHandsBagSettled(player database.Player, bag *redBagT) {
	my.send(player, &cow_proto.RedHandsBagSettled{bag.RedBagClear()})
}

func (my *actorT) sendRedBagDestoried(player database.Player, id int32) {
	my.send(player, &cow_proto.RedBagDestroyed{id})
}

// ----------------------------------------------------

func (my *actorT) sendRedUpdateBagForAll(bag *redBagT) {
	for _, player := range bag.Players {
		my.sendRedUpdateBag(player.Player, bag)
	}
}

// ---------------------------------------------------------------------------------------------------------------------

func (my *actorT) sendLever28CreateBagFailed(player database.Player, reason int32) {
	my.send(player, &cow_proto.Lever28CreateBagFailed{reason})
}

func (my *actorT) sendLever28CreateBagSuccess(player database.Player, id int32) {
	my.send(player, &cow_proto.Lever28CreateBagSuccess{id})
}

func (my *actorT) sendLever28UpdateBagList(player database.Player, bags lever28BagMapT) {
	my.send(player, &cow_proto.Lever28UpdateBagList{bags.Lever28Bag(player)})
}

func (my *actorT) sendLever28GrabFailed(player database.Player, reason int32) {
	my.send(player, &cow_proto.Lever28GrabFailed{reason})
}

func (my *actorT) sendLever28GrabSuccess(player database.Player) {
	my.send(player, &cow_proto.Lever28GrabSuccess{})
}

func (my *actorT) sendLever28Deadline(player database.Player, deadline int64) {
	my.send(player, &cow_proto.Lever28Deadline{deadline})
}

func (my *actorT) sendLever28UpdateBag(player database.Player, bag *lever28BagT) {
	my.send(player, &cow_proto.Lever28UpdateBag{bag.Lever28Bag()})
}

func (my *actorT) sendLever28BagDestroyed(player database.Player, id int32) {
	my.send(player, &cow_proto.Lever28BagDestroyed{id})
}

// ----------------------------------------------------

func (my *actorT) Lever28UpdateBagForAll(bag *lever28BagT) {
	for _, player := range bag.Players {
		my.sendLever28UpdateBag(player.Player, bag)
	}
}

// ---------------------------------------------------------------------------------------------------------------------

func (my *actorT) sendGomokuCreateRoomFailed(player database.Player, reason int32) {
	my.send(player, &cow_proto.GomokuCreateRoomFailed{reason})
}

func (my *actorT) sendGomokuCreateRoomSuccess(player database.Player, id int32) {
	my.send(player, &cow_proto.GomokuCreateRoomSuccess{id})
}

func (my *actorT) sendGomokuJoinRoomFailed(player database.Player, reason int32) {
	my.send(player, &cow_proto.GomokuJoinRoomFailed{reason})
}

func (my *actorT) sendGomokuJoinRoomSuccess(player database.Player) {
	my.send(player, &cow_proto.GomokuJoinRoomSuccess{})
}

func (my *actorT) sendGomokuSetRoomCostFailed(player database.Player, reason int32) {
	my.send(player, &cow_proto.GomokuSetRoomCostFailed{reason})
}

func (my *actorT) sendGomokuUpdateRoom(player database.Player, room *gomokuRoomT) {
	my.send(player, &cow_proto.GomokuUpdateRoom{room.GomokuRoom()})
}

func (my *actorT) sendGomokuLeft(player database.Player, reason int32) {
	my.send(player, &cow_proto.GomokuLeft{reason})
}

func (my *actorT) sendGomokuStarted(player database.Player) {
	my.send(player, &cow_proto.GomokuStarted{})
}

func (my *actorT) sendGomokuUpdateRound(player database.Player, room *gomokuRoomT) {
	my.send(player, &cow_proto.GomokuUpdateRound{room.RoundNumber, room.Board.ToSlice()})
}

func (my *actorT) sendGomokuRequirePlay(player database.Player, is bool) {
	my.send(player, &cow_proto.GomokuRequirePlay{is})
}

func (my *actorT) sendGomokuUpdatePlayDeadline(player database.Player, deadline int64, is bool) {
	my.send(player, &cow_proto.GomokuUpdatePlayDeadline{deadline, is})
}

func (my *actorT) sendGomokuVictory(player, victory, loser database.Player) {
	my.send(player, &cow_proto.GomokuVictory{})
}

func (my *actorT) sendGomokuLost(player, victory, loser database.Player) {
	my.send(player, &cow_proto.GomokuLost{})
}

// ----------------------------------------------------

func (my *actorT) sendGomokuUpdateRoomForAll(room *gomokuRoomT) {
	if room.Creator != nil {
		my.sendGomokuUpdateRoom(room.Creator.Player, room)
	}
	if room.Student != nil {
		my.sendGomokuUpdateRoom(room.Student.Player, room)
	}
}

func (my *actorT) sendGomokuLeftForAll(room *gomokuRoomT, reason int32) {
	if room.Creator != nil {
		my.sendGomokuLeft(room.Creator.Player, reason)
	}
	if room.Student != nil {
		my.sendGomokuLeft(room.Student.Player, reason)
	}
}

func (my *actorT) sendGomokuStartedForAll(room *gomokuRoomT) {
	if room.Creator != nil {
		my.sendGomokuStarted(room.Creator.Player)
	}
	if room.Student != nil {
		my.sendGomokuStarted(room.Student.Player)
	}
}

func (my *actorT) sendGomokuUpdateRoundForAll(room *gomokuRoomT) {
	if room.Creator != nil {
		my.sendGomokuUpdateRound(room.Creator.Player, room)
	}
	if room.Student != nil {
		my.sendGomokuUpdateRound(room.Student.Player, room)
	}
}

func (my *actorT) sendGomokuUpdatePlayDeadlineForAll(room *gomokuRoomT, deadline int64) {
	if room.ThisPlayer != nil {
		my.sendGomokuUpdatePlayDeadline(room.ThisPlayer.Player, deadline, true)
	}
	if room.AnotherPlayer != nil {
		my.sendGomokuUpdatePlayDeadline(room.AnotherPlayer.Player, deadline, false)
	}
}

// ---------------------------------------------------------------------------------------------------------------------
