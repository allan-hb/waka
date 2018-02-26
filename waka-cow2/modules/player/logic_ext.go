package player

import (
	"github.com/sirupsen/logrus"

	"github.com/liuhan907/waka/waka-cow2/database"
	"github.com/liuhan907/waka/waka-cow2/modules/hall/hall_message"
	"github.com/liuhan907/waka/waka-cow2/proto"
)

func (my *actorT) NiuniuShareContinue(ev *cow_proto.NiuniuShareContinue) {
	number, err := database.PlayerShared(my.player)
	if err != nil {
		log.WithFields(logrus.Fields{
			"player": my.player,
			"err":    err,
		}).Warnln("share continue failed")
	}

	log.WithFields(logrus.Fields{
		"player": my.player,
		"number": number,
	}).Debugln("share continue number")

	if number > 0 {
		my.hall.Tell(&hall_message.UpdatePlayerSecret{my.player})

		log.WithFields(logrus.Fields{
			"player":   my.player,
			"diamonds": number,
		}).Debugln("share continue and get diamonds")
	}
}
