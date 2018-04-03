package hall

import (
	"math"
	"time"

	"github.com/sirupsen/logrus"
	"gopkg.in/ahmetb/go-linq.v3"

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
	GrabAt time.Time

	Pay    int32
	Charge int32

	Lookup bool
}

func (player *redBagPlayerT) RedBagClearPlayerData() *cow_proto.RedBagClear_PlayerData {
	return &cow_proto.RedBagClear_PlayerData{
		Player:  int32(player.Player),
		Grab:    player.Grab,
		GrabAt:  player.GrabAt.Format("2006-01-02 15:04:05"),
		Creator: player.Bag.Creator.Player == player.Player,
		Pay:     player.Pay,
		Charge:  player.Charge,
	}
}

type redBagCreatorT struct {
	Bag    *redBagT
	Player database.Player
	Freeze database.Freeze
	Cost   int32
	Get    int32
	Charge int32
}

func (player *redBagCreatorT) RedBagClearCreatorData() *cow_proto.RedBagClear_CreatorData {
	return &cow_proto.RedBagClear_CreatorData{
		Player: int32(player.Player),
		Cost:   player.Cost,
		Get:    player.Get,
		Charge: player.Charge,
	}
}

// ---------------------------------------------------------------------------------------------------------------------

type redBagPlayerMapT map[database.Player]*redBagPlayerT

func (players redBagPlayerMapT) RedBagClearPlayerData() (r []*cow_proto.RedBagClear_PlayerData) {
	for _, player := range players {
		r = append(r, player.RedBagClearPlayerData())
	}
	return r
}

// ---------------------------------------------------------------------------------------------------------------------

type redBagT struct {
	Hall *actorT

	Id        int32
	Option    *cow_proto.RedBagOption
	Creator   *redBagCreatorT
	Players   redBagPlayerMapT
	CreateAt  time.Time
	DeadAt    time.Time
	FinallyAt time.Time

	Settled bool

	RemainMoney []int32
}

func (bag *redBagT) RedBag() *cow_proto.RedBag {
	r := &cow_proto.RedBag{
		Id:      bag.Id,
		Option:  bag.Option,
		Creator: int32(bag.Creator.Player),
	}
	linq.From(bag.Players).SelectT(func(in linq.KeyValue) int32 {
		return int32(in.Value.(*redBagPlayerT).Player)
	}).ToSlice(&r.Players)
	return r
}

func (bag *redBagT) RedBagClear() *cow_proto.RedBagClear {
	return &cow_proto.RedBagClear{
		Id:       bag.Id,
		Option:   bag.Option,
		Creator:  bag.Creator.RedBagClearCreatorData(),
		Players:  bag.Players.RedBagClearPlayerData(),
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

func (bags redBagMapT) RedBag(player database.Player) (r []*cow_proto.RedBag) {
	for _, bag := range bags {
		if int32(len(bag.Players)) < bag.Option.Number ||
			(bag.Players[player] != nil && !bag.Players[player].Lookup) {
			r = append(r, bag.RedBag())
		}
	}
	return r
}

// ---------------------------------------------------------------------------------------------------------------------

func (bag *redBagT) Left(player database.Player) {}

func (bag *redBagT) Recover(player database.Player) {
	bag.Hall.sendRedGrabSuccess(player)
	bag.Hall.sendRedDeadline(player, bag.DeadAt.Unix())
	bag.Hall.sendRedUpdateBag(player, bag)
}

func (bag *redBagT) Create(hall *actorT, id int32, option *cow_proto.RedBagOption, creator database.Player) {
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
		bag.Hall.sendRedCreateBagFailed(creator, 1)
		return
	}

	remainMoney, err := red.SplitMoney(option.GetMoney(), option.GetNumber())
	if err != nil {
		log.WithFields(logrus.Fields{
			"creator": creator,
			"option":  option.String(),
			"err":     err,
		}).Warnln("create red but split money failed")
		bag.Hall.sendRedCreateBagFailed(creator, 0)
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
		bag.Hall.sendRedCreateBagFailed(creator, 0)
		return
	}

	bag.Creator.Freeze = freeze

	bag.Hall.redBags[bag.Id] = bag

	bag.Hall.sendRedCreateBagSuccess(creator, bag.Id)
	for _, player := range bag.Hall.players.SelectOnline() {
		bag.Hall.sendRedUpdateBagList(player.Player, bag.Hall.redBags)
	}

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
		bag.Hall.sendRedUpdateBag(player.Player, bag)
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
		bag.Hall.sendRedDeadline(player.Player, bag.DeadAt.Unix())
		bag.Hall.sendRedUpdateBagForAll(bag)
		for _, player := range bag.Hall.players.SelectOnline() {
			bag.Hall.sendRedUpdateBagList(player.Player, bag.Hall.redBags)
		}

		if int32(len(bag.Players)) == bag.Option.Number {
			bag.settle()
		}
	}
}

func (bag *redBagT) settle() {
	bag.FinallyAt = time.Now()

	costs := &database.RedBagCost{
		Creator: &database.RedCreatorCost{
			Player: bag.Creator.Player,
			Freeze: bag.Creator.Freeze,
		},
	}
	for _, player := range bag.Players {
		costs.Grabs = append(costs.Grabs, &database.RedGrabCost{
			Player: player.Player,
			Freeze: player.Freeze,
			Number: player.Grab,
		})
	}
	if linq.From(bag.Players).
		GroupByT(func(in linq.KeyValue) int32 {
			return in.Value.(*redBagPlayerT).Grab % 10
		}, func(in linq.KeyValue) int32 {
			return in.Value.(*redBagPlayerT).Grab % 10
		}).
		WhereT(func(in linq.Group) bool {
			return linq.From(bag.Option.Mantissa).AnyWithT(func(any int32) bool {
				return in.Key.(int32) == any
			})
		}).
		Count() == len(bag.Option.Mantissa) {
		for _, player := range bag.Players {
			suffix := player.Grab % 10
			if linq.From(bag.Option.Mantissa).AnyWithT(func(in int32) bool {
				return suffix == in
			}) {
				costs.Pays = append(costs.Pays, &database.RedPayCost{
					Player: player.Player,
					Number: bag.LostMoney(),
				})
			}
		}
	}

	for _, player := range bag.Players {
		player.Charge = int32(float64(player.Grab)*0.05 + 0.5)
	}

	for _, pay := range costs.Pays {
		if player := bag.Players[pay.Player]; player != nil {
			player.Pay = pay.Number
		}
	}

	for _, player := range bag.Players {
		if bag.Creator.Player != player.Player {
			bag.Creator.Get += player.Pay
		}
	}

	bag.Creator.Charge = int32(float64(bag.Creator.Get)*0.05 + 0.5)

	bag.Creator.Cost = bag.CreateMoney()

	if err := database.RedBagCostSettle(costs); err != nil {
		log.WithFields(logrus.Fields{
			"id":     bag.Id,
			"option": bag.Option.String(),
			"costs":  costs,
			"err":    err,
		}).Warnln("red settle failed")
	} else {
		err := database.RedAddHandHistory(bag.Creator.Player, bag.RedBagClear())
		if err != nil {
			log.WithFields(logrus.Fields{
				"err": err,
			}).Warnln("add red hands war history failed")
		}
		for _, player := range bag.Players {
			if bag.Creator.Player != player.Player {
				err := database.RedAddGrabHistory(player.Player, bag.RedBagClear())
				if err != nil {
					log.WithFields(logrus.Fields{
						"err": err,
					}).Warnln("add red grab war history failed")
				}
			}
		}

		bag.Settled = true

		if bag.Players[bag.Creator.Player] == nil {
			bag.Hall.sendRedHandsBagSettled(bag.Creator.Player, bag)
		}

		for _, player := range bag.Hall.players.SelectOnline() {
			bag.Hall.sendRedUpdateBagList(player.Player, bag.Hall.redBags)
		}

		log.WithFields(logrus.Fields{
			"id": bag.Id,
		}).Debugln("settled")
	}

}

func (bag *redBagT) Clock() {
	if bag.RemainTime() <= 0 {
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
		}

		for _, player := range bag.Players {
			if player := bag.Hall.players[player.Player]; player != nil {
				if player.InsideRed == bag.Id {
					bag.Hall.sendRedBagDestoried(player.Player, bag.Id)
				}
				player.InsideRed = 0
			}
		}
	}
}
