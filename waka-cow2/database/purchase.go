package database

import (
	"time"

	"github.com/jinzhu/gorm"
)

// 牛牛约战房间消费记录
type CowOrderRoomPurchaseHistory struct {
	// 主键
	Id int32 `gorm:"index;primary_key;AUTO_INCREMENT"`
	// 谁
	Player Player `gorm:"index"`
	// 房间 ID
	RoomId int32
	// 金币数
	Number int32
	// 时间
	CreatedAt time.Time
}

// ---------------------------------------------------------------------------------------------------------------------

// 牛牛代开房间消费记录
type CowPayForAnotherRoomPurchaseHistory struct {
	// 主键
	Id int32 `gorm:"index;primary_key;AUTO_INCREMENT"`
	// 谁
	Player Player `gorm:"index"`
	// 房间 ID
	RoomId int32
	// 金币数
	Number int32
	// 时间
	CreatedAt time.Time
}

// ---------------------------------------------------------------------------------------------------------------------

// 场费结算数据
type CowPlayerRoomCost struct {
	Player Player
	Number int32
}

// 牛牛约战房间场费结算
func CowOrderSettle(room int32, players []*CowPlayerRoomCost) error {
	var changed []Player
	var modifies []*modifyDiamondsAction

	for i := range players {
		player := players[i]

		modifies = append(modifies, &modifyDiamondsAction{
			Player: player.Player,
			Number: player.Number * (-1),
			After: func(ts *gorm.DB, self *modifyDiamondsAction) error {
				consume := &CowOrderRoomPurchaseHistory{
					Player:    player.Player,
					RoomId:    room,
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
			delete(playersByUnionId, playerData.UnionId)
			delete(playersByToken, playerData.Token)
		}
	}
	playersLock.Unlock()

	return nil
}

// 牛牛代开房间场费结算
func CowPayForAnotherSettle(room int32, player *CowPlayerRoomCost) error {
	var changed []Player
	var modifies []*modifyDiamondsAction

	modifies = append(modifies, &modifyDiamondsAction{
		Player: player.Player,
		Number: player.Number * (-1),
		After: func(ts *gorm.DB, self *modifyDiamondsAction) error {
			consume := &CowPayForAnotherRoomPurchaseHistory{
				Player:    player.Player,
				RoomId:    room,
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
			delete(playersByUnionId, playerData.UnionId)
			delete(playersByToken, playerData.Token)
		}
	}
	playersLock.Unlock()

	return nil
}
