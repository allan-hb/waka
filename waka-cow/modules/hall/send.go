package hall

import (
	"reflect"

	"github.com/golang/protobuf/proto"
	"github.com/liuhan907/waka/waka-cow/conf"
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

func (my *actorT) sendPlayer(player database.Player) {
	my.send(player, my.ToPlayer(player))
}

func (my *actorT) sendPlayerSecret(player database.Player) {
	my.send(player, my.ToPlayerSecret(player))
}

func (my *actorT) sendHallEntered(player database.Player) {
	my.send(player, &waka.HallEntered{
		Player: my.ToPlayerSecret(player),
	})
}

func (my *actorT) sendWelcome(player database.Player) {
	my.send(player, &waka.Welcome{
		Customers:   database.GetCustomerServices(),
		Notice:      database.GetNotice(),
		NoticeBig:   database.GetNoticeBig(),
		PayUrl:      database.GetPayURL(),
		RegisterUrl: database.GetRegisterURL(),
		LoginUrl:    database.GetLoginURL(),
	})
}

func (my *actorT) sendPlayerNumber(player database.Player, number int32) {
	my.send(player, &waka.PlayerNumber{
		Number: number + conf.Option.Hall.MinPlayerNumber,
	})
}

func (my *actorT) sendRecover(player database.Player, is bool, name string) {
	my.send(player, &waka.Recover{
		Is:   is,
		Name: name,
	})
}

// ---------------------------------------------------------------------------------------------------------------------

func (my *actorT) sendNiuniuCreateRoomFailed(player database.Player, reason int32) {
	my.send(player, &waka.NiuniuCreateRoomFailed{reason})
}

func (my *actorT) sendNiuniuJoinRoomFailed(player database.Player, reason int32) {
	my.send(player, &waka.NiuniuJoinRoomFailed{reason})
}

func (my *actorT) sendNiuniuRoomJoined(player database.Player, room cowRoom) {
	my.send(player, &waka.NiuniuRoomJoined{room.NiuniuRoomData2()})
}

func (my *actorT) sendNiuniuRoomCreated(player database.Player) {
	my.send(player, &waka.NiuniuRoomCreated{})
}

func (my *actorT) sendNiuniuRoomLeft(player database.Player) {
	my.send(player, &waka.NiuniuRoomLeft{})
}

func (my *actorT) sendNiuniuRoomLeftByDismiss(player database.Player) {
	my.send(player, &waka.NiuniuRoomLeftByDismiss{})
}

func (my *actorT) sendNiuniuRoomLeftByMoneyNotEnough(player database.Player) {
	my.send(player, &waka.NiuniuRoomLeftByMoneyNotEnough{})
}

func (my *actorT) sendNiuniuUpdateRoom(player database.Player, room cowRoom) {
	my.send(player, &waka.NiuniuUpdateRoom{room.NiuniuRoomData2()})
}

func (my *actorT) sendNiuniuUpdateRound(player database.Player, room cowRoom) {
	my.send(player, &waka.NiuniuUpdateRound{room.NiuniuRoundStatus(player)})
}

func (my *actorT) sendNiuniuCountdown(player database.Player, number int32) {
	my.send(player, &waka.NiuniuCountdown{number})
}

func (my *actorT) sendNiuniuStarted(player database.Player, number int32) {
	my.send(player, &waka.NiuniuStarted{number})
}

func (my *actorT) sendNiuniuDeal4(player database.Player, pokers []string) {
	my.send(player, &waka.NiuniuDeal4{pokers})
}

func (my *actorT) sendNiuniuRequireGrab(player database.Player) {
	my.send(player, &waka.NiuniuRequireGrab{})
}

func (my *actorT) sendNiuniuRequireSpecifyBanker(player database.Player, is bool) {
	my.send(player, &waka.NiuniuRequireSpecifyBanker{is})
}

func (my *actorT) sendNiuniuGrabAnimation(player database.Player, room cowRoom) {
	my.send(player, room.NiuniuGrabAnimation())
}

func (my *actorT) sendNiuniuRequireSpecifyRate(player database.Player, is bool) {
	my.send(player, &waka.NiuniuRequireSpecifyRate{is})
}

func (my *actorT) sendNiuniuDeal1(player database.Player, poker, pokersType string, pokers []string) {
	my.send(player, &waka.NiuniuDeal1{
		poker,
		pokersType,
		pokers,
	})
}

func (my *actorT) sendNiuniuSettleSuccess(player database.Player) {
	my.send(player, &waka.NiuniuSettleSuccess{})
}

func (my *actorT) sendNiuniuRoundClear(player database.Player, room cowRoom) {
	my.send(player, room.NiuniuRoundClear())
}

func (my *actorT) sendNiuniuRoundFinally(player database.Player, room cowRoom) {
	my.send(player, room.NiuniuRoundFinally())
}

// ----------------------------------------------------

func (my *actorT) sendNiuniuUpdateRoomForAll(room cowRoom) {
	for _, player := range room.GetPlayers() {
		my.sendNiuniuUpdateRoom(player, room)
	}
	for _, player := range room.GetObservers() {
		my.sendNiuniuUpdateRoom(player, room)
	}
}

func (my *actorT) sendNiuniuUpdateRoundForAll(room cowRoom) {
	for _, player := range room.GetPlayers() {
		my.sendNiuniuUpdateRound(player, room)
	}
	for _, player := range room.GetObservers() {
		my.sendNiuniuUpdateRound(player, room)
	}
}

func (my *actorT) sendNiuniuCountdownForAll(room cowRoom, number int32) {
	for _, player := range room.GetPlayers() {
		my.sendNiuniuCountdown(player, number)
	}
	for _, player := range room.GetObservers() {
		my.sendNiuniuCountdown(player, number)
	}
}

func (my *actorT) sendNiuniuStartedForAll(room cowRoom, number int32) {
	for _, player := range room.GetPlayers() {
		my.sendNiuniuStarted(player, number)
	}
	for _, player := range room.GetObservers() {
		my.sendNiuniuStarted(player, number)
	}
}

// ---------------------------------------------------------------------------------------------------------------------

func (my *actorT) sendRedCreateRedPaperBagFailed(player database.Player, reason int32) {
	my.send(player, &waka.RedCreateRedPaperBagFailed{reason})
}

func (my *actorT) sendRedGrabFailed(player database.Player, reason int32) {
	my.send(player, &waka.RedGrabFailed{reason})
}

func (my *actorT) sendRedCreateRedPaperBagSuccess(player database.Player, id int32) {
	my.send(player, &waka.RedCreateRedPaperBagSuccess{id})
}

func (my *actorT) sendRedUpdateRedPaperBagList(player database.Player, bags redBagMapT) {
	my.send(player, &waka.RedUpdateRedPaperBagList{bags.RedRedPaperBag1(player)})
}

func (my *actorT) sendRedGrabSuccess(player database.Player) {
	my.send(player, &waka.RedGrabSuccess{})
}

func (my *actorT) sendRedUpdateRedPaperBag(player database.Player, bag *redBagT) {
	my.send(player, &waka.RedUpdateRedPaperBag{bag.RedRedPaperBag2()})
}

func (my *actorT) sendRedHandsRedPaperBagSettled(player database.Player, bag *redBagT) {
	my.send(player, &waka.RedHandsRedPaperBagSettled{bag.RedRedPaperBag3()})
}

func (my *actorT) sendRedRedPaperBagCountdown(player database.Player, number int32) {
	my.send(player, &waka.RedRedPaperBagCountdown{number})
}

func (my *actorT) sendRedRedPaperBagDestory(player database.Player, id int32) {
	my.send(player, &waka.RedRedPaperBagDestory{id})
}

// ----------------------------------------------------

func (my *actorT) sendRedUpdateRedPaperBagForAll(bag *redBagT) {
	for _, player := range bag.Players {
		my.sendRedUpdateRedPaperBag(player.Player, bag)
	}
}

// ---------------------------------------------------------------------------------------------------------------------

func (my *actorT) sendLever28CreateRedPaperBagFailed(player database.Player, reason int32) {
	my.send(player, &waka.Lever28CreateRedPaperBagFailed{reason})
}

func (my *actorT) sendLever28GrabFailed(player database.Player, reason int32) {
	my.send(player, &waka.Lever28GrabFailed{reason})
}

func (my *actorT) sendLever28CreateRedPaperBagSuccess(player database.Player, id int32) {
	my.send(player, &waka.Lever28CreateRedPaperBagSuccess{id})
}

func (my *actorT) sendLever28UpdateRedPaperBagList(player database.Player, bags lever28BagMapT) {
	my.send(player, &waka.Lever28UpdateRedPaperBagList{bags.Lever28RedPaperBag1(player)})
}

func (my *actorT) sendLever28GrabSuccess(player database.Player) {
	my.send(player, &waka.Lever28GrabSuccess{})
}

func (my *actorT) sendLever28UpdateRedPaperBag(player database.Player, bag *lever28BagT) {
	my.send(player, &waka.Lever28UpdateRedPaperBag{bag.Lever28RedPaperBag2()})
}

func (my *actorT) sendLever28RedPaperBagCountdown(player database.Player, number int32) {
	my.send(player, &waka.Lever28RedPaperBagCountdown{number})
}

func (my *actorT) sendLever28RedPaperBagDestory(player database.Player, id int32) {
	my.send(player, &waka.Lever28RedPaperBagDestory{id})
}

// ----------------------------------------------------

func (my *actorT) sendLever28UpdateRedPaperBagForAll(bag *lever28BagT) {
	for _, player := range bag.Players {
		my.sendLever28UpdateRedPaperBag(player.Player, bag)
	}
}

// ---------------------------------------------------------------------------------------------------------------------

func (my *actorT) sendGomokuCreateRoomFailed(player database.Player, reason int32) {
	my.send(player, &waka.GomokuCreateRoomFailed{reason})
}

func (my *actorT) sendGomokuJoinRoomFailed(player database.Player, reason int32) {
	my.send(player, &waka.GomokuJoinRoomFailed{reason})
}

func (my *actorT) sendGomokuSetRoomCostFailed(player database.Player, reason int32) {
	my.send(player, &waka.GomokuSetRoomCostFailed{reason})
}

func (my *actorT) sendGomokuRoomCreated(player database.Player, id int32) {
	my.send(player, &waka.GomokuRoomCreated{id})
}

func (my *actorT) sendGomokuRoomEntered(player database.Player) {
	my.send(player, &waka.GomokuRoomEntered{})
}

func (my *actorT) sendGomokuLeft(player database.Player) {
	my.send(player, &waka.GomokuLeft{})
}

func (my *actorT) sendGomokuLeftByDismiss(player database.Player) {
	my.send(player, &waka.GomokuLeftByDismiss{})
}

func (my *actorT) sendGomokuUpdateRoom(player database.Player, room *gomokuRoomT) {
	my.send(player, &waka.GomokuUpdateRoom{room.GomokuRoom()})
}

func (my *actorT) sendGomokuStarted(player database.Player) {
	my.send(player, &waka.GomokuStarted{})
}

func (my *actorT) sendGomokuRequirePlay(player database.Player, is bool) {
	my.send(player, &waka.GomokuRequirePlay{is})
}

func (my *actorT) sendGomokuUpdatePlayCountdown(player database.Player, number int32, is bool) {
	my.send(player, &waka.GomokuUpdatePlayCountdown{number, is})
}

func (my *actorT) sendGomokuUpdateRound(player database.Player, room *gomokuRoomT) {
	my.send(player, &waka.GomokuUpdateRound{room.RoundNumber, room.Board.ToSlice()})
}

func (my *actorT) sendGomokuVictory(player database.Player) {
	my.send(player, &waka.GomokuVictory{})
}

func (my *actorT) sendGomokuLost(player database.Player) {
	my.send(player, &waka.GomokuLost{})
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

func (my *actorT) sendGomokuLeftForAll(room *gomokuRoomT) {
	if room.Creator != nil {
		my.sendGomokuLeft(room.Creator.Player)
	}
	if room.Student != nil {
		my.sendGomokuLeft(room.Student.Player)
	}
}

func (my *actorT) sendGomokuLeftByDismissForAll(room *gomokuRoomT) {
	if room.Creator != nil {
		my.sendGomokuLeftByDismiss(room.Creator.Player)
	}
	if room.Student != nil {
		my.sendGomokuLeftByDismiss(room.Student.Player)
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

func (my *actorT) sendGomokuUpdatePlayCountdownForAll(room *gomokuRoomT, number int32) {
	if room.ThisPlayer != nil {
		my.sendGomokuUpdatePlayCountdown(room.ThisPlayer.Player, number, true)
	}
	if room.AnotherPlayer != nil {
		my.sendGomokuUpdatePlayCountdown(room.AnotherPlayer.Player, number, false)
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

// ---------------------------------------------------------------------------------------------------------------------
