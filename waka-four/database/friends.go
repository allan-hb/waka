package database

import (
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// 好友
type FriendData struct {
	// 主键
	Id int32 `gorm:"index;primary_key;AUTO_INCREMENT"`

	// 玩家 ID
	Player Player `gorm:"index"`
	// 好友 ID
	Friend Player
	// 屏蔽
	Ban bool

	// 创建时间
	CreatedAt time.Time
}

// 申请
type AskData struct {
	// 主键
	Id int32 `gorm:"index;primary_key;AUTO_INCREMENT"`

	// 玩家 ID
	Player Player `gorm:"index"`
	// 发送者
	Sender Player `gorm:"index"`
	// 状态
	// 0 未处理
	// 1 已拒绝
	// 2 已通过
	Status int32

	// 创建时间
	CreatedAt time.Time
}

// 查询好友
func QueryFriendList(player Player) (d []*FriendData, e error) {
	if err := mysql.Where("player = ? and ban = ?", player, false).Find(&d).Error; err != nil {
		return nil, err
	}
	return d, nil
}

// 查询被屏蔽好友
func QueryBanFriendList(player Player) (d []*FriendData, e error) {
	if err := mysql.Where("player = ? and ban = ?", player, true).Find(&d).Error; err != nil {
		return nil, err
	}
	return d, nil
}

// 查询已发送的添加列表
func QueryWantListSend(player Player) (d []*AskData, e error) {
	if err := mysql.Where("sender = ? and status = ?", player, 0).Find(&d).Error; err != nil {
		return nil, err
	}
	return d, nil
}

// 查询已被处理的添加列表
func QueryWantListDeal(player Player, limit int32) (d []*AskData, e error) {
	if err := mysql.Where("sender = ? and status <> ?", player, 0).Limit(limit).Find(&d).Error; err != nil {
		return nil, err
	}
	return d, nil
}

// 查询未处理的申请列表
func QueryAskListUndeal(player Player) (d []*AskData, e error) {
	if err := mysql.Where("player = ? and status = ?", player, 0).Find(&d).Error; err != nil {
		return nil, err
	}
	return d, nil
}

// 查询已处理的申请列表
func QueryAskListDeal(player Player, limit int32) (d []*AskData, e error) {
	if err := mysql.Where("player = ? and status <> ?", player, 0).Limit(limit).Find(&d).Error; err != nil {
		return nil, err
	}
	return d, nil
}

// 屏蔽好友
func BanFriend(player, friend Player) error {
	if err := mysql.Model(new(FriendData)).Where("player = ? and friend = ?", player, friend).Updates(map[string]interface{}{
		"ban": true,
	}).Error; err != nil {
		return err
	}

	friendsLock.Lock()
	friendsByPlayer[uint64(player)<<32|uint64(friend)] = false
	friendsLock.Unlock()

	return nil
}

// 解除屏蔽好友
func CancelBanFriend(player, friend Player) error {
	if err := mysql.Model(new(FriendData)).Where("player = ? and friend = ?", player, friend).Updates(map[string]interface{}{
		"ban": false,
	}).Error; err != nil {
		return err
	}

	friendsLock.Lock()
	friendsByPlayer[uint64(player)<<32|uint64(friend)] = true
	friendsLock.Unlock()

	return nil
}

// 发送申请
func WantFriend(player, friend Player) error {
	c := 0
	if err := mysql.Model(new(FriendData)).Where("player = ? and friend = ?", player, friend).Count(&c).Error; err != nil {
		return err
	}

	if c > 0 {
		return nil
	}

	if err := mysql.Create(&AskData{
		Player:    friend,
		Sender:    player,
		Status:    0,
		CreatedAt: time.Now(),
	}).Error; err != nil {
		return err
	}

	return nil
}

// 回应申请
func ReplayAskFriend(number, operate int32) error {
	ask := AskData{
		Id: number,
	}
	if err := mysql.First(&ask).Error; err != nil {
		return err
	}

	if ask.Status != 0 {
		return nil
	}

	ts := mysql.Begin()

	c := 0
	if err := mysql.Model(new(FriendData)).Where("player = ? and friend = ?", ask.Player, ask.Sender).Count(&c).Error; err != nil {
		return err
	}
	if c == 0 && operate == 1 {
		if err := ts.Create(&FriendData{
			Player:    ask.Player,
			Friend:    ask.Sender,
			CreatedAt: time.Now(),
		}).Error; err != nil {
			ts.Rollback()
			return err
		}

		friendsLock.Lock()
		friendsByPlayer[uint64(ask.Player)<<32|uint64(ask.Sender)] = true
		friendsLock.Unlock()
	}
	if err := mysql.Model(new(FriendData)).Where("player = ? and friend = ?", ask.Sender, ask.Player).Count(&c).Error; err != nil {
		return err
	}
	if c == 0 && operate == 1 {
		if err := ts.Create(&FriendData{
			Player:    ask.Sender,
			Friend:    ask.Player,
			CreatedAt: time.Now(),
		}).Error; err != nil {
			ts.Rollback()
			return err
		}

		friendsLock.Lock()
		friendsByPlayer[uint64(ask.Sender)<<32|uint64(ask.Player)] = true
		friendsLock.Unlock()
	}

	if operate == 1 {
		ask.Status = 2
	} else {
		ask.Status = 1
	}

	if err := ts.Save(&ask).Error; err != nil {
		ts.Rollback()
		return err
	}

	ts.Commit()

	return nil
}

var (
	friendsLock     sync.RWMutex
	friendsByPlayer = make(map[uint64]bool, 12800)
)

// 查询玩家是否能加入私密房间
func QueryPlayerCanJoin(creator, player Player) bool {
	friendsLock.RLock()
	can, being := friendsByPlayer[uint64(creator)<<32|uint64(player)]
	friendsLock.RUnlock()

	if !being {
		c := 0
		if err := mysql.Model(new(FriendData)).Where("player = ? and friend = ? and ban = ?", creator, player, false).Count(&c).Error; err != nil {
			log.WithFields(logrus.Fields{
				"creator": creator,
				"player":  player,
				"err":     err,
			}).Warnln("query friend failed")
			return false
		}
		if c > 0 {
			can = true
		} else {
			can = false
		}
		friendsByPlayer[uint64(creator)<<32|uint64(player)] = can
	}

	return can
}
