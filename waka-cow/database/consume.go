package database

import (
	"time"

	"github.com/jinzhu/gorm"

	"github.com/liuhan907/waka/waka-cow/conf"
)

// 场费结算数据
type CowPlayerRoomCost struct {
	Player Player
	Number int32
}

// 牛牛约战房间场费结算
func CowOrderRoomSettle(room int32, players []*CowPlayerRoomCost) error {
	var changed []Player
	var modifies []*modifyMoneyAction

	for i := range players {
		player := players[i]

		modifies = append(modifies, &modifyMoneyAction{
			Player: player.Player,
			Number: player.Number * (-1),
		})
		changed = append(changed, player.Player)

		modifies = append(modifies, &modifyMoneyAction{
			Player: player.Player.PlayerData().Supervisor,
			Number: int32(float64(player.Number)*float64(float64(player.Player.SupervisorData().BonusRate)/100) + 0.5),
			After: func(ts *gorm.DB, self *modifyMoneyAction) error {
				consume := &CowPlayerRoomPurchaseHistory{
					Player:    player.Player,
					Room:      room,
					Number:    player.Number,
					CreatedAt: time.Now(),
				}
				if err := ts.Create(consume).Error; err != nil {
					return err
				}
				if player.Player.PlayerData().Supervisor > 0 {
					bonus := newBonusByPlayerRoomCost(
						player.Player.PlayerData().Supervisor,
						player.Player,
						int32(float64(player.Number)*float64(float64(player.Player.PlayerData().Supervisor.SupervisorData().BonusRate)/100)+0.5),
						consume.Ref,
					)
					if err := ts.Create(bonus).Error; err != nil {
						return err
					}
				}
				return nil
			},
		})
		changed = append(changed, player.Player.PlayerData().Supervisor)
	}

	ts := mysql.Begin()

	err := applyModifyMoneyAction(ts, modifies)
	if err != nil {
		ts.Rollback()
		return err
	}

	ts.Commit()

	playersLock.Lock()
	for _, player := range changed {
		playerData, being := playersByPlayer[player]
		if being {
			delete(playersByPlayer, player)
			delete(playersByUnionID, playerData.UnionID)
			delete(playersByToken, playerData.Token)
		}
	}
	playersLock.Unlock()

	return nil
}

// 牛牛代开房间场费结算
func CowPayForAnotherRoomSettle(room int32, player *CowPlayerRoomCost) error {
	var changed []Player
	var modifies []*modifyMoneyAction

	modifies = append(modifies, &modifyMoneyAction{
		Player: player.Player,
		Number: player.Number * (-1),
	})
	changed = append(changed, player.Player)

	modifies = append(modifies, &modifyMoneyAction{
		Player: player.Player.PlayerData().Supervisor,
		Number: int32(float64(player.Number)*float64(float64(player.Player.PlayerData().Supervisor.SupervisorData().BonusRate)/100) + 0.5),
		After: func(ts *gorm.DB, self *modifyMoneyAction) error {
			consume := &CowPlayerRoomPurchaseHistory{
				Player:    player.Player,
				Room:      room,
				Number:    player.Number,
				CreatedAt: time.Now(),
			}
			if err := ts.Create(consume).Error; err != nil {
				return err
			}
			if player.Player.PlayerData().Supervisor > 0 {
				bonus := newBonusByPayForAnotherRoomCost(
					player.Player.PlayerData().Supervisor,
					player.Player,
					int32(float64(player.Number)*float64(float64(player.Player.PlayerData().Supervisor.SupervisorData().BonusRate)/100)+0.5),
					consume.Ref,
				)
				if err := ts.Create(bonus).Error; err != nil {
					return err
				}
			}
			return nil
		},
	})
	changed = append(changed, player.Player.PlayerData().Supervisor)

	ts := mysql.Begin()

	err := applyModifyMoneyAction(ts, modifies)
	if err != nil {
		ts.Rollback()
		return err
	}

	ts.Commit()

	playersLock.Lock()
	for _, player := range changed {
		playerData, being := playersByPlayer[player]
		if being {
			delete(playersByPlayer, player)
			delete(playersByUnionID, playerData.UnionID)
			delete(playersByToken, playerData.Token)
		}
	}
	playersLock.Unlock()

	return nil
}

// 胜负结算数据
type CowGoldRoomCost struct {
	Victory Player
	Loser   Player
	Number  int32
}

// 牛牛代理房间胜负结算
func CowGoldRoomSettle(room int32, players []*CowGoldRoomCost) error {
	var changed []Player
	var modifies []*modifyMoneyAction

	for i := range players {
		player := players[i]

		modifies = append(modifies, &modifyMoneyAction{
			Player: player.Loser,
			Number: player.Number * (-1),
		})
		changed = append(changed, player.Loser)

		modifies = append(modifies, &modifyMoneyAction{
			Player: player.Victory,
			Number: int32(float64(player.Number) * (1.0 - float64(conf.Option.Hall.WaterRate)/100)),
		})
		changed = append(changed, player.Victory)

		modifies = append(modifies, &modifyMoneyAction{
			Player: player.Victory.PlayerData().Supervisor,
			Number: int32(float64(player.Number) * (float64(conf.Option.Hall.WaterRate) / 100) *
				(float64(player.Victory.PlayerData().Supervisor.SupervisorData().BonusRate) / 100)),
			After: func(ts *gorm.DB, self *modifyMoneyAction) error {
				consume := &CowGoldRoomPurchaseHistory{
					Victory:   player.Victory,
					Loser:     player.Loser,
					Room:      room,
					Number:    player.Number,
					CreatedAt: time.Now(),
				}
				if err := ts.Create(consume).Error; err != nil {
					return err
				}
				if player.Victory.PlayerData().Supervisor > 0 {
					bonus := newBonusByGoldRoomCost(
						player.Victory.PlayerData().Supervisor,
						player.Victory,
						int32(float64(player.Number)*float64(float64(player.Victory.PlayerData().Supervisor.SupervisorData().BonusRate)/100)+0.5),
						consume.Ref,
					)
					if err := ts.Create(bonus).Error; err != nil {
						return err
					}
				}
				return nil
			},
		})
		changed = append(changed, player.Victory.PlayerData().Supervisor)
	}

	ts := mysql.Begin()

	err := applyModifyMoneyAction(ts, modifies)
	if err != nil {
		ts.Rollback()
		return err
	}

	ts.Commit()

	playersLock.Lock()
	for _, player := range changed {
		playerData, being := playersByPlayer[player]
		if being {
			delete(playersByPlayer, player)
			delete(playersByUnionID, playerData.UnionID)
			delete(playersByToken, playerData.Token)
		}
	}
	playersLock.Unlock()

	return nil
}

// 冻结玩家金币
func FreezeMoney(player Player, number int32) (Freeze, error) {

	ts := mysql.Begin()

	freeze, err := freezeMoney(ts, player, number)
	if err != nil {
		ts.Rollback()
		return 0, err
	}

	ts.Commit()

	playersLock.Lock()
	playerData, being := playersByPlayer[player]
	if being {
		delete(playersByPlayer, player)
		delete(playersByUnionID, playerData.UnionID)
		delete(playersByToken, playerData.Token)
	}
	playersLock.Unlock()

	return freeze, nil
}

// 解冻玩家金币
func RecoverFreezeMoney(freeze Freeze) error {
	ts := mysql.Begin()

	player, err := recoverFreezeMoney(ts, freeze)
	if err != nil {
		ts.Rollback()
		return err
	}

	ts.Commit()

	playersLock.Lock()
	playerData, being := playersByPlayer[player]
	if being {
		delete(playersByPlayer, player)
		delete(playersByUnionID, playerData.UnionID)
		delete(playersByToken, playerData.Token)
	}
	playersLock.Unlock()

	return nil
}

// 红包创建者结算数据
type RedCreatorCost struct {
	Player Player
	Grab   int32
	Pay    int32
	Freeze Freeze
}

// 红包玩家结算数据
type RedPlayerCost struct {
	Player Player
	Grab   int32
	Pay    int32
	Freeze Freeze
}

// 红包结算数据
type RedBagCost struct {
	Creator *RedCreatorCost
	Players []*RedPlayerCost
}

// 红包结算
func RedSettle(bag *RedBagCost) error {
	var changed []Player
	var modifies []*modifyMoneyAction

	modifies = append(modifies, &modifyMoneyAction{
		Player: bag.Creator.Player,
		Number: bag.Creator.Grab,
	})
	changed = append(changed, bag.Creator.Player)

	modifies = append(modifies, &modifyMoneyAction{
		Player: bag.Creator.Player,
		Number: bag.Creator.Pay * (-1),
		Before: func(ts *gorm.DB, modify *modifyMoneyAction) error {
			_, err := recoverFreezeMoney(ts, bag.Creator.Freeze)
			return err
		},
	})
	changed = append(changed, bag.Creator.Player)

	for i := range bag.Players {
		player := bag.Players[i]

		modifies = append(modifies, &modifyMoneyAction{
			Player: player.Player,
			Number: int32(float64(player.Grab) * (1.0 - float64(conf.Option.Hall.WaterRate)/100)),
		})

		modifies = append(modifies, &modifyMoneyAction{
			Player: player.Player,
			Number: player.Pay * (-1),
			Before: func(ts *gorm.DB, modify *modifyMoneyAction) error {
				_, err := recoverFreezeMoney(ts, player.Freeze)
				return err
			},
		})
		changed = append(changed, player.Player)

		modifies = append(modifies, &modifyMoneyAction{
			Player: player.Player.PlayerData().Supervisor,
			Number: int32(float64(player.Grab) * (float64(conf.Option.Hall.WaterRate) / 100) *
				(float64(player.Player.PlayerData().Supervisor.SupervisorData().BonusRate) / 100)),
			After: func(ts *gorm.DB, self *modifyMoneyAction) error {
				if player.Player.PlayerData().Supervisor > 0 {
					bonus := newBonusByRedPaperBag(
						player.Player.PlayerData().Supervisor,
						player.Player,
						int32(float64(player.Grab)*float64(float64(player.Player.PlayerData().Supervisor.SupervisorData().BonusRate)/100)+0.5),
						0,
					)
					if err := ts.Create(bonus).Error; err != nil {
						return err
					}
				}
				return nil
			},
		})
		changed = append(changed, player.Player.PlayerData().Supervisor)
	}

	ts := mysql.Begin()

	err := applyModifyMoneyAction(ts, modifies)
	if err != nil {
		ts.Rollback()
		return err
	}

	ts.Commit()

	playersLock.Lock()
	for _, player := range changed {
		playerData, being := playersByPlayer[player]
		if being {
			delete(playersByPlayer, player)
			delete(playersByUnionID, playerData.UnionID)
			delete(playersByToken, playerData.Token)
		}
	}
	playersLock.Unlock()

	return nil
}

// 二八杠玩家结算数据
type Lever28PlayerCost struct {
	Player Player
	Grab   int32
	Pay    int32
	Freeze Freeze
}

// 二八杠结算数据
type Lever28Cost struct {
	Players []*Lever28PlayerCost
}

// 二八杠结算
func Lever28Settle(bag *Lever28Cost) error {
	var changed []Player
	var modifies []*modifyMoneyAction

	for i := range bag.Players {
		player := bag.Players[i]

		modifies = append(modifies, &modifyMoneyAction{
			Player: player.Player,
			Number: int32(float64(player.Grab) * (1.0 - float64(conf.Option.Hall.WaterRate)/100)),
		})

		modifies = append(modifies, &modifyMoneyAction{
			Player: player.Player,
			Number: player.Pay * (-1),
			Before: func(ts *gorm.DB, modify *modifyMoneyAction) error {
				_, err := recoverFreezeMoney(ts, player.Freeze)
				return err
			},
		})
		changed = append(changed, player.Player)

		modifies = append(modifies, &modifyMoneyAction{
			Player: player.Player.PlayerData().Supervisor,
			Number: int32(float64(player.Grab) * (float64(conf.Option.Hall.WaterRate) / 100) * (float64(player.Player.PlayerData().Supervisor.SupervisorData().BonusRate) / 100)),
			After: func(ts *gorm.DB, self *modifyMoneyAction) error {
				if player.Player.PlayerData().Supervisor > 0 {
					bonus := newBonusByLever28(
						player.Player.PlayerData().Supervisor,
						player.Player,
						int32(float64(player.Grab)*float64(float64(player.Player.PlayerData().Supervisor.SupervisorData().BonusRate)/100)+0.5),
						0,
					)
					if err := ts.Create(bonus).Error; err != nil {
						return err
					}
				}
				return nil
			},
		})
		changed = append(changed, player.Player.PlayerData().Supervisor)
	}

	ts := mysql.Begin()

	err := applyModifyMoneyAction(ts, modifies)
	if err != nil {
		ts.Rollback()
		return err
	}

	ts.Commit()

	playersLock.Lock()
	for _, player := range changed {
		playerData, being := playersByPlayer[player]
		if being {
			delete(playersByPlayer, player)
			delete(playersByUnionID, playerData.UnionID)
			delete(playersByToken, playerData.Token)
		}
	}
	playersLock.Unlock()

	return nil
}

// 五子棋结算
func GomokuSettle(master, student Player, money int32) error {
	var changed []Player
	var modifies []*modifyMoneyAction

	modifies = append(modifies, &modifyMoneyAction{
		Player: student,
		Number: money * (-1),
	})
	changed = append(changed, student)

	modifies = append(modifies, &modifyMoneyAction{
		Player: master,
		Number: int32(float64(money) * (1.0 - float64(conf.Option.Hall.WaterRate)/100)),
	})
	changed = append(changed, master)

	modifies = append(modifies, &modifyMoneyAction{
		Player: master.PlayerData().Supervisor,
		Number: int32(float64(money) * (float64(conf.Option.Hall.WaterRate) / 100) * (float64(master.PlayerData().Supervisor.SupervisorData().BonusRate) / 100)),
		After: func(ts *gorm.DB, self *modifyMoneyAction) error {
			consume := &GomokuPurchaseHistory{
				Master:    master,
				Student:   student,
				Number:    money,
				CreatedAt: time.Now(),
			}
			if err := ts.Create(consume).Error; err != nil {
				return err
			}
			if master.PlayerData().Supervisor > 0 {
				bonus := newBonusByGomokuCost(
					master.PlayerData().Supervisor,
					master,
					int32(float64(money)*float64(float64(master.PlayerData().Supervisor.SupervisorData().BonusRate)/100)+0.5),
					consume.Ref,
				)
				if err := ts.Create(bonus).Error; err != nil {
					return err
				}
			}
			return nil
		},
	})
	changed = append(changed, master.PlayerData().Supervisor)

	ts := mysql.Begin()

	err := applyModifyMoneyAction(ts, modifies)
	if err != nil {
		ts.Rollback()
		return err
	}

	ts.Commit()

	playersLock.Lock()
	for _, player := range changed {
		playerData, being := playersByPlayer[player]
		if being {
			delete(playersByPlayer, player)
			delete(playersByUnionID, playerData.UnionID)
			delete(playersByToken, playerData.Token)
		}
	}
	playersLock.Unlock()

	return nil
}
