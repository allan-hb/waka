package hall

import (
	"errors"
	"reflect"
	"strings"

	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/sirupsen/logrus"

	"github.com/liuhan907/waka/waka-four/database"
	"github.com/liuhan907/waka/waka/modules/supervisor/supervisor_message"
)

func (my *actorT) ReceiveSupervisor(context actor.Context) bool {
	switch ev := context.Message().(type) {
	case *supervisor_message.PlayerEntered:
		my.playerEntered(ev)
	case *supervisor_message.PlayerExchanged:
		my.playerExchanged(ev)
	case *supervisor_message.PlayerLeft:
		my.playerLeft(ev)
	case *supervisor_message.PlayerTransported:
		my.playerTransported(ev)
	case *supervisor_message.PlayerFutureRequested:
		my.playerFutureRequested(ev)
	default:
		return false
	}
	return true
}

// ---------------------------------------------------------------------------------------------------------------------

func (my *actorT) playerEntered(ev *supervisor_message.PlayerEntered) {
	log.WithFields(logrus.Fields{
		"player": ev.Player,
	}).Debugln("player entered")

	my.playerEnteredExchanged(database.Player(ev.Player), ev.Remote)
}

func (my *actorT) playerExchanged(ev *supervisor_message.PlayerExchanged) {
	log.WithFields(logrus.Fields{
		"player": ev.Player,
	}).Debugln("player exchanged")

	my.playerEnteredExchanged(database.Player(ev.Player), ev.Remote)
}

func (my *actorT) playerEnteredExchanged(player database.Player, remote string) {
	if lines := strings.Split(remote, ":"); len(lines) == 2 {
		remote = lines[0]
	}

	playerData, being := my.players[player]
	if !being {
		playerData = &playerT{
			Player: player,
			Remote: remote,
		}
		my.players[player] = playerData
	} else {
		playerData.Remote = remote
	}

	players := my.players.SelectOnline()
	playerNumber := int32(len(players))
	my.sendHallEntered(player)
	for _, player := range players {
		my.sendPlayerNumber(player.Player, playerNumber)
	}

	if playerData.InsideFour != 0 {
		room, being := my.fourRooms[playerData.InsideFour]
		if being {
			my.sendRecover(player, true, "four")
			room.Recover(playerData)
		} else {
			playerData.InsideFour = 0
			my.sendRecover(player, false, "")
		}
	} else {
		my.sendRecover(player, false, "")
	}
}

func (my *actorT) playerLeft(ev *supervisor_message.PlayerLeft) {
	log.WithFields(logrus.Fields{
		"player": ev.Player,
	}).Debugln("player left")

	player := database.Player(ev.Player)

	playerData, being := my.players[player]
	if !being {
		log.WithFields(logrus.Fields{
			"player": ev.Player,
		}).Warnln("player left but player not found")
		return
	}

	playerData.Remote = ""

	if playerData.InsideFour != 0 {
		room, being := my.fourRooms[playerData.InsideFour]
		if being {
			room.Left(playerData)
		} else {
			playerData.InsideFour = 0
		}
	} else {
		delete(my.players, player)
	}

	players := my.players.SelectOnline()
	playerNumber := int32(len(players))
	for _, player := range players {
		my.sendPlayerNumber(player.Player, playerNumber)
	}
}

// ---------------------------------------------------------------------------------------------------------------------

func (my *actorT) playerTransported(ev *supervisor_message.PlayerTransported) {
	log.WithFields(logrus.Fields{
		"player":  ev.Player,
		"type":    reflect.TypeOf(ev.Payload).Elem().Name(),
		"payload": ev.Payload.String(),
	}).Debugln("player transport")

	player := database.Player(ev.Player)

	playerData, being := my.players[player]
	if !being {
		log.WithFields(logrus.Fields{
			"player":  ev.Player,
			"type":    reflect.TypeOf(ev.Payload).Elem().Name(),
			"payload": ev.Payload.String(),
		}).Warnln("player transport but player not found")
		return
	}

	if my.playerTransportedFour(playerData, ev) {
		return
	}

	log.WithFields(logrus.Fields{
		"player":  ev.Player,
		"type":    reflect.TypeOf(ev.Payload).Elem().Name(),
		"payload": ev.Payload.String(),
	}).Warnln("unknown player transport type")
}

// ---------------------------------------------------------------------------------------------------------------------

func (my *actorT) playerFutureRequested(ev *supervisor_message.PlayerFutureRequested) {
	log.WithFields(logrus.Fields{
		"player":  ev.Player,
		"type":    reflect.TypeOf(ev.Payload).Elem().Name(),
		"payload": ev.Payload.String(),
	}).Debugln("player future")

	player := database.Player(ev.Player)

	playerData, being := my.players[player]
	if !being {
		log.WithFields(logrus.Fields{
			"player":  ev.Player,
			"type":    reflect.TypeOf(ev.Payload).Elem().Name(),
			"payload": ev.Payload.String(),
		}).Warnln("player future but player not found")
		return
	}

	if my.playerFutureRequestedFour(playerData, ev) {
		return
	}

	ev.Respond(nil, errors.New("unsupported request"))

	log.WithFields(logrus.Fields{
		"player":  ev.Player,
		"type":    reflect.TypeOf(ev.Payload).Elem().Name(),
		"payload": ev.Payload.String(),
	}).Warnln("unknown player future type")
}
