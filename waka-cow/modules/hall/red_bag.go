package hall

import (
	"math"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/liuhan907/waka/waka-cow/conf"
	"github.com/liuhan907/waka/waka-cow/database"
	"github.com/liuhan907/waka/waka-cow/modules/hall/tools/red"
	"github.com/liuhan907/waka/waka-cow/proto"
)

const (
	// 红包存在时间
	kBagAliveTime = time.Second * 120
)

type redBagPlayerT struct {
	Bag    *redBagT
	Player database.Player
	Freeze database.Freeze
	Grab   int32
	Pay    int32
	Charge int32
	GrabAt time.Time
	Lookup bool
}

func (player *redBagPlayerT) PbPlayer() *waka.Player {
	return player.Bag.Hall.ToPlayer(player.Player)
}

func (player *redBagPlayerT) RedRedPaperBag3RedPlayer() *waka.RedRedPaperBag3_RedPlayer {
	r := &waka.RedRedPaperBag3_RedPlayer{
		Player:  player.Bag.Hall.ToPlayer(player.Player),
		Grab:    player.Grab,
		Pay:     player.Pay,
		Charge:  player.Charge,
		GrabAt:  player.GrabAt.Format("2006-01-02 15:04:05"),
		Creator: player.Bag.Creator.Player == player.Player,
	}
	return r
}

// ---------------------------------------------------------------------------------------------------------------------

type redBagPlayerMapT map[database.Player]*redBagPlayerT

func (players redBagPlayerMapT) PbPlayer() []*waka.Player {
	var d []*waka.Player
	for _, player := range players {
		d = append(d, player.PbPlayer())
	}
	return d
}

func (players redBagPlayerMapT) RedRedPaperBag3RedPlayer() []*waka.RedRedPaperBag3_RedPlayer {
	var d []*waka.RedRedPaperBag3_RedPlayer
	for _, player := range players {
		d = append(d, player.RedRedPaperBag3RedPlayer())
	}
	return d
}

// ---------------------------------------------------------------------------------------------------------------------

type redBagCreatorT struct {
	Player database.Player
	Freeze database.Freeze
	Get    int32
	Cost   int32
	Charge int32
}

type redBagT struct {
	Hall *actorT

	Id        int32
	Option    *waka.RedRedPaperBagOption
	Creator   *redBagCreatorT
	Players   redBagPlayerMapT
	CreateAt  time.Time
	DeadAt    time.Time
	FinallyAt time.Time

	Settled bool

	RemainMoney []int32
}

func (bag *redBagT) RedRedPaperBag1(player database.Player) *waka.RedRedPaperBag1 {
	playerData := bag.Creator.Player.PlayerData()
	_, myGrabbed := bag.Players[player]
	return &waka.RedRedPaperBag1{
		Id:           bag.Id,
		Option:       bag.Option,
		PlayerNumber: int32(len(bag.Players)),
		Creator: &waka.RedRedPaperBag1_RedPlayer{
			Nickname: playerData.Nickname,
			Head:     playerData.Head,
		},
		MyGrabbed: myGrabbed,
	}
}

func (bag *redBagT) RedRedPaperBag2() *waka.RedRedPaperBag2 {
	return &waka.RedRedPaperBag2{
		Id:      bag.Id,
		Option:  bag.Option,
		Players: bag.Players.PbPlayer(),
	}
}

func (bag *redBagT) RedRedPaperBag3() *waka.RedRedPaperBag3 {
	return &waka.RedRedPaperBag3{
		Id:     bag.Id,
		Option: bag.Option,
		Creator: &waka.RedRedPaperBag3_RedCreator{
			Player: bag.Hall.ToPlayer(bag.Creator.Player),
			Get:    bag.Creator.Get,
			Cost:   bag.Creator.Cost,
			Charge: bag.Creator.Charge,
		},
		Players:  bag.Players.RedRedPaperBag3RedPlayer(),
		UsedTime: int32(bag.FinallyAt.Sub(bag.CreateAt).Seconds()),
	}
}

func (bag *redBagT) CreateMoney() int32 {
	return bag.Option.Money
}

func (bag *redBagT) EnterMoney() int32 {
	if bag.Option.Number == 7 {
		if len(bag.Option.Mantissa) == 1 {
			return int32(float64(bag.Option.Money)*1.5 + 0.5)
		} else if len(bag.Option.Mantissa) == 2 {
			return int32(float64(bag.Option.Money)*1.6 + 0.5)
		} else if len(bag.Option.Mantissa) == 3 {
			return int32(float64(bag.Option.Money)*2.0 + 0.5)
		}
	} else if bag.Option.Number == 10 {
		return bag.Option.Money
	}
	return math.MaxInt32
}

func (bag *redBagT) LostMoney() int32 {
	return bag.EnterMoney()
}

func (bag *redBagT) RemainTime() int32 {
	return int32(bag.DeadAt.Sub(time.Now()).Seconds())
}

// ---------------------------------------------------------------------------------------------------------------------

type redBagMapT map[int32]*redBagT

func (bags redBagMapT) RedRedPaperBag1(player database.Player) []*waka.RedRedPaperBag1 {
	var d []*waka.RedRedPaperBag1
	for _, bag := range bags {
		if int32(len(bag.Players)) < bag.Option.Number ||
			(bag.Players[player] != nil && !bag.Players[player].Lookup) {
			d = append(d, bag.RedRedPaperBag1(player))
		}
	}
	return d
}

// ---------------------------------------------------------------------------------------------------------------------

func (bag *redBagT) Left(player database.Player) {}

func (bag *redBagT) Recover(player database.Player) {
	bag.Hall.sendRedGrabSuccess(player)
	bag.Hall.sendRedUpdateRedPaperBag(player, bag)
}

func (bag *redBagT) Create(hall *actorT, id int32, option *waka.RedRedPaperBagOption, creator database.Player) {
	*bag = redBagT{
		Hall:   hall,
		Id:     id,
		Option: option,
		Creator: &redBagCreatorT{
			Player: creator,
		},
		CreateAt: time.Now(),
		DeadAt:   time.Now().Add(kBagAliveTime),
		Players:  make(redBagPlayerMapT, 10),
	}

	if creator.PlayerData().Money < bag.CreateMoney() {
		log.WithFields(logrus.Fields{
			"creator":      creator,
			"option":       option.String(),
			"create_money": bag.CreateMoney(),
		}).Debugln("create red but money not enough")
		bag.Hall.sendRedCreateRedPaperBagFailed(creator, 1)
		return
	}

	remainMoney, err := red.SplitMoney(option.GetMoney(), option.GetNumber())
	if err != nil {
		log.WithFields(logrus.Fields{
			"creator": creator,
			"option":  option.String(),
			"err":     err,
		}).Warnln("create red but split money failed")
		bag.Hall.sendRedCreateRedPaperBagFailed(creator, 0)
		return
	}

	bag.RemainMoney = remainMoney

	freeze, err := database.FreezeMoney(creator, bag.CreateMoney())
	if err != nil {
		log.WithFields(logrus.Fields{
			"creator": creator,
			"option":  option.String(),
			"err":     err,
		}).Warnln("create red but freeze money failed")
		bag.Hall.sendRedCreateRedPaperBagFailed(creator, 0)
		return
	}

	bag.Creator.Freeze = freeze

	bag.Hall.redBags[bag.Id] = bag

	bag.Hall.sendRedCreateRedPaperBagSuccess(creator, bag.Id)
	for _, player := range bag.Hall.players.SelectOnline() {
		bag.Hall.sendRedUpdateRedPaperBagList(player.Player, bag.Hall.redBags)
	}

	bag.Hall.sendPlayerSecret(creator)

	log.WithFields(logrus.Fields{
		"creator": creator,
		"option":  option.String(),
		"id":      bag.Id,
	}).Debugln("created red")
}

func (bag *redBagT) Grab(player *playerT) {
	_, being := bag.Players[player.Player]
	if being {
		player.InsideRed = bag.Id
		bag.Hall.sendRedGrabSuccess(player.Player)
		bag.Hall.sendRedUpdateRedPaperBag(player.Player, bag)
	} else {
		if len(bag.Players) >= int(bag.Option.Number) {
			log.WithFields(logrus.Fields{
				"player": player,
				"id":     bag.Id,
			}).Warnln("grab red but out of max player number")
			bag.Hall.sendRedGrabFailed(player.Player, 3)
			return
		}

		if player.Player.PlayerData().Money < bag.EnterMoney() {
			log.WithFields(logrus.Fields{
				"player":      player,
				"enter_money": bag.EnterMoney(),
			}).Debugln("grab red but money not enough")
			bag.Hall.sendRedGrabFailed(player.Player, 2)
			return
		}

		freeze, err := database.FreezeMoney(player.Player, bag.EnterMoney())
		if err != nil {
			log.WithFields(logrus.Fields{
				"player":      player,
				"enter_money": bag.EnterMoney(),
				"err":         err,
			}).Warnln("grab red but freeze money failed")
			bag.Hall.sendRedGrabFailed(player.Player, 0)
			return
		}

		log.WithFields(logrus.Fields{
			"player": player.Player,
			"id":     bag.Id,
		}).Debugln("grab red")

		redBagPlayer := &redBagPlayerT{
			Bag:    bag,
			Player: player.Player,
			Freeze: freeze,
			Grab:   bag.RemainMoney[len(bag.RemainMoney)-1],
			GrabAt: time.Now(),
		}
		bag.RemainMoney = bag.RemainMoney[:len(bag.RemainMoney)-1]

		bag.Players[player.Player] = redBagPlayer

		player.InsideRed = bag.Id

		bag.Hall.sendRedGrabSuccess(player.Player)
		bag.Hall.sendRedUpdateRedPaperBagForAll(bag)
		for _, player := range bag.Hall.players.SelectOnline() {
			bag.Hall.sendRedUpdateRedPaperBagList(player.Player, bag.Hall.redBags)
		}

		bag.Hall.sendPlayerSecret(player.Player)

		if int32(len(bag.Players)) == bag.Option.Number {
			bag.settle()
		}
	}
}

func (bag *redBagT) settle() {
	bag.FinallyAt = time.Now()

	// 计算闲家的赔付
	booms := make([]bool, len(bag.Option.Mantissa))
	for _, player := range bag.Players {
		suffix := player.Grab % 10
		for i, mantissa := range bag.Option.Mantissa {
			if suffix == mantissa {
				booms[i] = true
				player.Pay = bag.LostMoney()
			}
		}
	}
	boom := true
	for _, v := range booms {
		boom = boom && v
	}
	if !boom {
		for _, player := range bag.Players {
			player.Pay = 0
		}
	}

	// 计算闲家的手续费
	for _, player := range bag.Players {
		player.Charge = int32(float64(player.Grab)*float64(float64(conf.Option.Hall.WaterRate)/100) + 0.5)
	}

	// 计算庄家获得的赔付
	for _, player := range bag.Players {
		if bag.Creator.Player != player.Player {
			bag.Creator.Get += player.Pay
		}
	}

	// 计算庄家的手续费
	bag.Creator.Charge = int32(float64(bag.Creator.Get)*float64(float64(conf.Option.Hall.WaterRate)/100) + 0.5)

	// 计算庄家的支出
	bag.Creator.Cost = bag.CreateMoney()

	// 结算
	cost := &database.RedBagCost{
		Creator: &database.RedCreatorCost{
			Player: bag.Creator.Player,
			Get:    bag.Creator.Get,
			Charge: bag.Creator.Charge,
			Cost:   bag.Creator.Cost,
			Freeze: bag.Creator.Freeze,
		},
	}
	for _, player := range bag.Players {
		cost.Players = append(cost.Players, &database.RedPlayerCost{
			Player: player.Player,
			Grab:   player.Grab,
			Charge: player.Charge,
			Pay:    player.Pay,
			Freeze: player.Freeze,
		})
	}
	if err := database.RedSettle(cost); err != nil {
		log.WithFields(logrus.Fields{
			"id":      bag.Id,
			"option":  bag.Option.String(),
			"creator": bag.Creator,
			"err":     err,
		}).Warnln("red settle failed")
	} else {
		err := database.RedAddHandWarHistory(bag.Creator.Player, bag.RedRedPaperBag3())
		if err != nil {
			log.WithFields(logrus.Fields{
				"err": err,
			}).Warnln("add red hands war history failed")
		}
		for _, player := range bag.Players {
			if bag.Creator.Player != player.Player {
				err := database.RedAddGrabWarHistory(player.Player, bag.RedRedPaperBag3())
				if err != nil {
					log.WithFields(logrus.Fields{
						"err": err,
					}).Warnln("add red grab war history failed")
				}
			}
		}

		bag.Settled = true

		// 如果庄家没抢，发送通知
		if bag.Players[bag.Creator.Player] == nil {
			bag.Hall.sendRedHandsRedPaperBagSettled(bag.Creator.Player, bag)
		}

		for _, player := range bag.Hall.players.SelectOnline() {
			bag.Hall.sendRedUpdateRedPaperBagList(player.Player, bag.Hall.redBags)
		}

		for _, player := range bag.Players {
			bag.Hall.sendPlayerSecret(player.Player)
		}

		log.WithFields(logrus.Fields{
			"id": bag.Id,
		}).Debugln("settled")
	}
}

func (bag *redBagT) Clock() {
	if bag.RemainTime() > 0 && len(bag.Players) < int(bag.Option.Number) {
		for _, player := range bag.Players {
			if player := bag.Hall.players[player.Player]; player != nil && player.InsideRed == bag.Id {
				bag.Hall.sendRedRedPaperBagCountdown(player.Player, bag.RemainTime())
			}
		}
	} else if bag.RemainTime() <= 0 {
		delete(bag.Hall.redBags, bag.Id)

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
				if player.InsideRed == bag.Id {
					bag.Hall.sendRedRedPaperBagDestory(player.Player, bag.Id)
				}
				player.InsideRed = 0
			}
		}
	}
}
