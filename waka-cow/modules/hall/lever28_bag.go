package hall

import (
	"time"

	"github.com/liuhan907/waka/waka-cow/conf"
	"github.com/sirupsen/logrus"

	"github.com/liuhan907/waka/waka-cow/database"
	"github.com/liuhan907/waka/waka-cow/modules/hall/tools/lever28"
	"github.com/liuhan907/waka/waka-cow/proto"
)

type lever28BagPlayerT struct {
	Bag        *lever28BagT
	Player     database.Player
	Freeze     database.Freeze
	Pay        int32
	Grab       int32
	GrabCharge int32
	Get        int32
	GetCharge  int32
	GrabAt     time.Time
	Mahjong    []int32
	Lookup     bool
}

func (player *lever28BagPlayerT) PbPlayer() *waka.Player {
	return player.Bag.Hall.ToPlayer(player.Player)
}

func (player *lever28BagPlayerT) Lever28RedPaperBag3RedPlayer() *waka.Lever28RedPaperBag3_RedPlayer {
	r := &waka.Lever28RedPaperBag3_RedPlayer{
		Player:     player.Bag.Hall.ToPlayer(player.Player),
		Pay:        player.Pay,
		Grab:       player.Grab,
		GrabCharge: player.GrabCharge,
		Get:        player.Get,
		GetCharge:  player.GetCharge,
		Mahjong:    player.Mahjong,
		GrabAt:     player.GrabAt.Format("2006-01-02 15:04:05"),
		Creator:    player.Bag.Creator.Player == player.Player,
	}
	return r
}

// ---------------------------------------------------------------------------------------------------------------------

type lever28BagPlayerMapT map[database.Player]*lever28BagPlayerT

func (players lever28BagPlayerMapT) PbPlayer() []*waka.Player {
	var d []*waka.Player
	for _, player := range players {
		d = append(d, player.PbPlayer())
	}
	return d
}

func (players lever28BagPlayerMapT) Lever28RedPaperBag3RedPlayer() []*waka.Lever28RedPaperBag3_RedPlayer {
	var d []*waka.Lever28RedPaperBag3_RedPlayer
	for _, player := range players {
		d = append(d, player.Lever28RedPaperBag3RedPlayer())
	}
	return d
}

// ---------------------------------------------------------------------------------------------------------------------

type lever28CreatorT struct {
	Player database.Player
	Freeze database.Freeze
}

type lever28BagT struct {
	Hall *actorT

	Id        int32
	Option    *waka.Lever28RedPaperBagOption
	Creator   *lever28CreatorT
	Players   lever28BagPlayerMapT
	CreateAt  time.Time
	DeadAt    time.Time
	FinallyAt time.Time

	Settled bool

	RemainMoney []int32
}

func (bag *lever28BagT) Lever28RedPaperBag1(player database.Player) *waka.Lever28RedPaperBag1 {
	playerData := bag.Creator.Player.PlayerData()
	_, myGrabbed := bag.Players[player]
	return &waka.Lever28RedPaperBag1{
		Id:           bag.Id,
		Option:       bag.Option,
		PlayerNumber: int32(len(bag.Players)),
		Creator: &waka.Lever28RedPaperBag1_RedPlayer{
			Nickname: playerData.Nickname,
			Head:     playerData.Head,
		},
		MyGrabbed: myGrabbed,
	}
}

func (bag *lever28BagT) Lever28RedPaperBag2() *waka.Lever28RedPaperBag2 {
	return &waka.Lever28RedPaperBag2{
		Id:      bag.Id,
		Option:  bag.Option,
		Players: bag.Players.PbPlayer(),
	}
}

func (bag *lever28BagT) Lever28RedPaperBag3() *waka.Lever28RedPaperBag3 {
	return &waka.Lever28RedPaperBag3{
		Id:       bag.Id,
		Option:   bag.Option,
		Players:  bag.Players.Lever28RedPaperBag3RedPlayer(),
		UsedTime: int32(bag.FinallyAt.Sub(bag.CreateAt).Seconds()),
	}
}

func (bag *lever28BagT) CreateMoney() int32 {
	return bag.Option.Money*3 + 10*100
}

func (bag *lever28BagT) EnterMoney() int32 {
	return bag.Option.Money + 10*100
}

func (bag *lever28BagT) LostMoney() int32 {
	return bag.Option.Money
}

func (bag *lever28BagT) RemainTime() int32 {
	return int32(bag.DeadAt.Sub(time.Now()).Seconds())
}

// ---------------------------------------------------------------------------------------------------------------------

type lever28BagMapT map[int32]*lever28BagT

func (bags lever28BagMapT) Lever28RedPaperBag1(player database.Player) []*waka.Lever28RedPaperBag1 {
	var d []*waka.Lever28RedPaperBag1
	for _, bag := range bags {
		if int32(len(bag.Players)) < 4 ||
			(bag.Players[player] != nil && !bag.Players[player].Lookup) {
			d = append(d, bag.Lever28RedPaperBag1(player))
		}
	}
	return d
}

// ---------------------------------------------------------------------------------------------------------------------

func (bag *lever28BagT) Left(player database.Player) {}

func (bag *lever28BagT) Recover(player database.Player) {
	bag.Hall.sendLever28GrabSuccess(player)
	bag.Hall.sendLever28UpdateRedPaperBag(player, bag)
}

func (bag *lever28BagT) Create(hall *actorT, id int32, option *waka.Lever28RedPaperBagOption, creator database.Player) {
	*bag = lever28BagT{
		Hall:   hall,
		Id:     id,
		Option: option,
		Creator: &lever28CreatorT{
			Player: creator,
		},
		CreateAt: time.Now(),
		DeadAt:   time.Now().Add(kBagAliveTime),
		Players:  make(lever28BagPlayerMapT, 4),
	}

	hallPlayer, being := bag.Hall.players[creator]
	if !being {
		log.WithFields(logrus.Fields{
			"creator": creator,
		}).Debugln("create lever28 but player not found")
		bag.Hall.sendLever28CreateRedPaperBagFailed(creator, 0)
		return
	}

	if creator.PlayerData().Money < bag.CreateMoney() {
		log.WithFields(logrus.Fields{
			"creator":      creator,
			"option":       option.String(),
			"create_money": bag.CreateMoney(),
		}).Debugln("create lever28 but money not enough")
		bag.Hall.sendLever28CreateRedPaperBagFailed(creator, 1)
		return
	}

	remainMoney, err := lever28.SplitMoney(40*100, 4)
	if err != nil {
		log.WithFields(logrus.Fields{
			"creator": creator,
			"option":  option.String(),
			"err":     err,
		}).Warnln("create lever28 but split money failed")
		bag.Hall.sendLever28CreateRedPaperBagFailed(creator, 0)
		return
	}

	bag.RemainMoney = remainMoney

	freeze, err := database.FreezeMoney(creator, bag.CreateMoney())
	if err != nil {
		log.WithFields(logrus.Fields{
			"creator": creator,
			"option":  option.String(),
			"err":     err,
		}).Warnln("create lever28 but freeze money failed")
		bag.Hall.sendLever28CreateRedPaperBagFailed(creator, 0)
		return
	}

	bag.Creator.Freeze = freeze

	bag.Hall.lever28Bags[bag.Id] = bag

	bag.Hall.sendLever28CreateRedPaperBagSuccess(creator, bag.Id)

	log.WithFields(logrus.Fields{
		"creator": creator,
		"option":  option.String(),
		"id":      bag.Id,
	}).Debugln("created lever28")

	bag.Grab(hallPlayer)
}

func (bag *lever28BagT) Grab(player *playerT) {
	_, being := bag.Players[player.Player]
	if being {
		player.InsideLever28 = bag.Id
		bag.Hall.sendLever28GrabSuccess(player.Player)
		bag.Hall.sendLever28UpdateRedPaperBag(player.Player, bag)
	} else {
		var freeze database.Freeze
		var err error
		if player.Player == bag.Creator.Player {
			freeze = bag.Creator.Freeze
		} else {
			if len(bag.Players) >= 4 {
				log.WithFields(logrus.Fields{
					"player": player,
					"id":     bag.Id,
				}).Warnln("grab lever28 but out of max player number")
				bag.Hall.sendLever28GrabFailed(player.Player, 3)
				return
			}

			if player.Player.PlayerData().Money < bag.EnterMoney() {
				log.WithFields(logrus.Fields{
					"player":      player,
					"enter_money": bag.EnterMoney(),
				}).Debugln("grab lever28 but money not enough")
				bag.Hall.sendLever28GrabFailed(player.Player, 2)
				return
			}

			freeze, err = database.FreezeMoney(player.Player, bag.EnterMoney())
			if err != nil {
				log.WithFields(logrus.Fields{
					"player":      player,
					"enter_money": bag.EnterMoney(),
					"err":         err,
				}).Warnln("grab lever28 but freeze money failed")
				bag.Hall.sendLever28GrabFailed(player.Player, 0)
				return
			}
		}

		grab := bag.RemainMoney[len(bag.RemainMoney)-1]
		lever28BagPlayer := &lever28BagPlayerT{
			Bag:     bag,
			Player:  player.Player,
			Freeze:  freeze,
			Grab:    grab,
			GrabAt:  time.Now(),
			Mahjong: []int32{(grab / 10) % 10, grab % 10},
		}
		bag.RemainMoney = bag.RemainMoney[:len(bag.RemainMoney)-1]

		bag.Players[player.Player] = lever28BagPlayer

		player.InsideLever28 = bag.Id

		bag.Hall.sendLever28GrabSuccess(player.Player)
		bag.Hall.sendLever28UpdateRedPaperBagForAll(bag)
		for _, player := range bag.Hall.players.SelectOnline() {
			bag.Hall.sendLever28UpdateRedPaperBagList(player.Player, bag.Hall.lever28Bags)
		}

		log.WithFields(logrus.Fields{
			"player": player.Player,
			"id":     bag.Id,
		}).Debugln("grab lever28")

		bag.Hall.sendPlayerSecret(player.Player)

		if int32(len(bag.Players)) == 4 {
			bag.settle()
		}
	}
}

func (bag *lever28BagT) settle() {
	bag.FinallyAt = time.Now()

	// 选择闲家
	var players []*lever28BagPlayerT
	for _, player := range bag.Players {
		if player.Player != bag.Creator.Player {
			players = append(players, player)
		}
	}

	// 计算庄家权重
	banker := bag.Players[bag.Creator.Player]
	bw, err := lever28.GetMahjongType(banker.Mahjong, true)
	if err != nil {
		log.WithFields(logrus.Fields{
			"id":      bag.Id,
			"option":  bag.Option.String(),
			"creator": bag.Creator,
			"err":     err,
		}).Warnln("get banker mahjong weight failed")
		return
	}

	// 计算各家赔付
	for _, player := range players {
		w, err := lever28.GetMahjongType(player.Mahjong, false)
		if err != nil {
			log.WithFields(logrus.Fields{
				"id":      bag.Id,
				"option":  bag.Option.String(),
				"creator": bag.Creator,
				"err":     err,
			}).Warnln("settle failed")
			return
		}

		if bw > w {
			banker.Get = banker.Get + bag.Option.Money
			player.Pay = player.Pay + bag.Option.Money
		} else if bw < w {
			player.Get = player.Get + bag.Option.Money
			banker.Pay = banker.Pay + bag.Option.Money
		} else if banker.Grab > player.Grab {
			banker.Get = banker.Get + bag.Option.Money
			player.Pay = player.Pay + bag.Option.Money
		} else if banker.Grab < player.Grab {
			player.Get = player.Get + bag.Option.Money
			banker.Pay = banker.Pay + bag.Option.Money
		}
	}

	// 计算手续费
	for _, player := range bag.Players {
		player.GrabCharge = int32(float64(player.Grab)*float64(float64(conf.Option.Hall.WaterRate)/100) + 0.5)
		player.GetCharge = int32(float64(player.Get)*float64(float64(conf.Option.Hall.WaterRate)/100) + 0.5)
	}

	cost := &database.Lever28Cost{}
	for _, player := range bag.Players {
		cost.Players = append(cost.Players, &database.Lever28PlayerCost{
			Player: player.Player,
			Grab:   player.Grab + player.Get,
			Charge: player.GrabCharge + player.GetCharge,
			Pay:    player.Pay + 10*100,
			Freeze: player.Freeze,
		})
	}

	// 结算
	if err := database.Lever28Settle(cost); err != nil {
		log.WithFields(logrus.Fields{
			"id":      bag.Id,
			"option":  bag.Option.String(),
			"creator": bag.Creator,
			"err":     err,
		}).Warnln("lever28 settle failed")
	} else {
		err := database.Lever28AddHandWarHistory(bag.Creator.Player, bag.Lever28RedPaperBag3())
		if err != nil {
			log.WithFields(logrus.Fields{
				"err": err,
			}).Warnln("add lever28 hands war history failed")
		}
		for _, player := range bag.Players {
			if bag.Creator.Player != player.Player {
				err := database.Lever28AddGrabWarHistory(player.Player, bag.Lever28RedPaperBag3())
				if err != nil {
					log.WithFields(logrus.Fields{
						"err": err,
					}).Warnln("add lever28 grab war history failed")
				}
			}
		}

		bag.Settled = true

		for _, player := range bag.Hall.players.SelectOnline() {
			bag.Hall.sendLever28UpdateRedPaperBagList(player.Player, bag.Hall.lever28Bags)
		}

		for _, player := range bag.Players {
			bag.Hall.sendPlayerSecret(player.Player)
		}

		log.WithFields(logrus.Fields{
			"id": bag.Id,
		}).Debugln("settled")
	}
}

func (bag *lever28BagT) Clock() {
	if bag.RemainTime() > 0 && len(bag.Players) < 4 {
		for _, player := range bag.Players {
			if player := bag.Hall.players[player.Player]; player != nil && player.InsideLever28 == bag.Id {
				bag.Hall.sendLever28RedPaperBagCountdown(player.Player, bag.RemainTime())
			}
		}
	} else if bag.RemainTime() <= 0 {
		delete(bag.Hall.lever28Bags, bag.Id)

		if !bag.Settled {
			if err := database.RecoverFreezeMoney(bag.Creator.Freeze); err != nil {
				log.WithFields(logrus.Fields{
					"freeze": bag.Creator.Freeze,
					"player": bag.Creator.Player,
					"err":    err,
				}).Warnln("recover freeze money failed")
			}
			for _, player := range bag.Players {
				if err := database.RecoverFreezeMoney(player.Freeze); err != nil {
					log.WithFields(logrus.Fields{
						"freeze": bag.Creator.Freeze,
						"player": bag.Creator.Player,
						"err":    err,
					}).Warnln("recover freeze money failed")
				}
			}

			for _, player := range bag.Players {
				bag.Hall.sendPlayerSecret(player.Player)
			}
		}

		for _, player := range bag.Players {
			if player := bag.Hall.players[player.Player]; player != nil {
				if player.InsideLever28 == bag.Id {
					bag.Hall.sendLever28RedPaperBagDestory(player.Player, bag.Id)
				}
				player.InsideLever28 = 0
			}
		}
	}
}
