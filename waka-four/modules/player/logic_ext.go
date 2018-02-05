package player

import (
	"github.com/liuhan907/waka/waka-four/database"
	"github.com/liuhan907/waka/waka-four/proto"
	"github.com/sirupsen/logrus"
)

func (my *actorT) FourShareContinue(ev *four_proto.FourShareContinue) {
	number, err := database.PlayerShared(my.player)
	if err != nil {
		log.WithFields(logrus.Fields{
			"player": my.player,
			"err":    err,
		}).Warnln("share continue failed")
	}

	if number > 0 {
		log.WithFields(logrus.Fields{
			"player":   my.player,
			"diamonds": number,
		}).Warnln("share continue and get diamonds")
	}
}
