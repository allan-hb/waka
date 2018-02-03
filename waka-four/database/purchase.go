package database

import (
	"time"

	"github.com/jinzhu/gorm"
)

// 场费结算数据
type FourPlayerRoomCost struct {
	Player Player
	Number int32
}

// 四张约战房间场费结算
func FourOrderRoomSettle(room int32, players []*FourPlayerRoomCost) error {
	var changed []Player
	var modifies []*modifyDiamondsAction

	for i := range players {
		player := players[i]

		modifies = append(modifies, &modifyDiamondsAction{
			Player: player.Player,
			Number: player.Number * (-1),
			After: func(ts *gorm.DB, self *modifyDiamondsAction) error {
				consume := &FourOrderRoomPurchaseHistory{
					Player:    player.Player,
					Room:      room,
					Number:    player.Number,
					CreatedAt: time.Now(),
				}
				if err := ts.Create(consume).Error; err != nil {
					return err
				}
				return nil
			},
		})
		changed = append(changed, player.Player)
	}

	ts := mysql.Begin()

	err := applyModifyDiamondsAction(ts, modifies)
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
			delete(playersByUnionID, playerData.UnionId)
			delete(playersByToken, playerData.Token)
		}
	}
	playersLock.Unlock()

	return nil
}

// 四张代开房间场费结算
func FourPayForAnotherRoomSettle(room int32, player *FourPlayerRoomCost) error {
	var changed []Player
	var modifies []*modifyDiamondsAction

	modifies = append(modifies, &modifyDiamondsAction{
		Player: player.Player,
		Number: player.Number * (-1),
		After: func(ts *gorm.DB, self *modifyDiamondsAction) error {
			consume := &FourPayForAnotherRoomPurchaseHistory{
				Player:    player.Player,
				Room:      room,
				Number:    player.Number,
				CreatedAt: time.Now(),
			}
			if err := ts.Create(consume).Error; err != nil {
				return err
			}
			return nil
		},
	})
	changed = append(changed, player.Player)

	ts := mysql.Begin()

	err := applyModifyDiamondsAction(ts, modifies)
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
			delete(playersByUnionID, playerData.UnionId)
			delete(playersByToken, playerData.Token)
		}
	}
	playersLock.Unlock()

	return nil
}
