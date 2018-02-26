package database

import (
	"sync"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"

	"github.com/liuhan907/waka/waka-cow2/conf"
)

var (
	ErrPlayerNotFound = errors.New("player not found")
)

// 玩家
type Player int32

func (player Player) PlayerData() *PlayerData {
	if player == 0 {
		return &PlayerData{
			Nickname: "unknown",
		}
	}

	playerData, being, err := QueryPlayerById(player)
	if err != nil {
		return &PlayerData{
			Nickname: "failed",
		}
	}
	if !being {
		return &PlayerData{
			Nickname: "unknown",
		}
	}
	return playerData
}

// 玩家数据
type PlayerData struct {
	// 主键
	Id Player `gorm:"index;primary_key;AUTO_INCREMENT"`
	// 微信 UnionId
	UnionId string `gorm:"index;unique"`
	// 登陆令牌
	Token string `gorm:"index"`

	// 昵称
	Nickname string
	// 头像URL
	Head string
	// 微信号
	Wechat string
	// 姓名
	Name string
	// 身份证
	Idcard string

	// 钱
	Diamonds int32

	// 封禁
	Ban int32

	// 创建时间
	CreatedAt time.Time

	// 胜率
	VictoryRate int32

	// 上次分享时间
	SharedAt time.Time
}

func (PlayerData) TableName() string {
	return "players"
}

// ---------------------------------------------------------------------------------------------------------------------

var (
	playersLock      sync.RWMutex
	playersByPlayer  = make(map[Player]*PlayerData)
	playersByUnionId = make(map[string]*PlayerData)
	playersByToken   = make(map[string]*PlayerData)
)

// ---------------------------------------------------------------------------------------------------------------------

// 注册玩家
func RegisterPlayer(unionId, nickname string, head, token string) (*PlayerData, error) {
	player := &PlayerData{
		UnionId:   unionId,
		Token:     token,
		Nickname:  nickname,
		Head:      head,
		Diamonds:  conf.Option.Hall.RegisterDiamonds,
		Ban:       0,
		CreatedAt: time.Now(),
		SharedAt:  time.Date(2018, 1, 1, 0, 0, 0, 0, time.Now().Location()),
	}
	if err := mysql.Create(player).Error; err != nil {
		return nil, err
	}

	playersLock.Lock()

	playersByPlayer[player.Id] = player
	playersByUnionId[player.UnionId] = player
	playersByToken[player.Token] = player

	playersLock.Unlock()

	return player, nil
}

// 更新玩家登录信息
func UpdatePlayerLogin(id Player, nickname string, head, token string) error {
	if err := mysql.Model(&PlayerData{}).Where("id = ?", id).Update(&PlayerData{
		Nickname: nickname,
		Head:     head,
		Token:    token,
	}).Error; err != nil {
		return err
	}

	playersLock.Lock()

	player, being := playersByPlayer[id]
	if being {
		delete(playersByToken, player.Token)

		player.Nickname = nickname
		player.Head = head
		player.Token = token

		playersByToken[player.Token] = player
	}

	playersLock.Unlock()

	return nil
}

// 更新玩家附加信息
func UpdatePlayerExt(id Player, wechat, name, idcard string) error {
	if err := mysql.Model(&PlayerData{}).Where("id = ?", id).Updates(&PlayerData{
		Wechat: wechat,
		Name:   name,
		Idcard: idcard,
	}).Error; err != nil {
		return err
	}

	playersLock.Lock()

	player, being := playersByPlayer[id]
	if being {
		player.Wechat = wechat
		player.Name = name
		player.Idcard = idcard
	}

	playersLock.Unlock()

	return nil
}

// 根据 ID 查询玩家
func QueryPlayerById(id Player) (*PlayerData, bool, error) {
	playersLock.RLock()

	player, being := playersByPlayer[id]
	if being {
		playersLock.RUnlock()
		return player, true, nil
	}

	playersLock.RUnlock()

	player = &PlayerData{}
	if err := mysql.Where("id = ?", id).First(player).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, false, nil
		} else {
			return nil, false, err
		}
	}

	playersLock.Lock()

	playersByPlayer[id] = player
	playersByUnionId[player.UnionId] = player
	playersByToken[player.Token] = player

	playersLock.Unlock()

	return player, true, nil
}

// 根据 Token 查询玩家
func QueryPlayerByToken(token string) (*PlayerData, bool, error) {
	playersLock.RLock()

	player, being := playersByToken[token]
	if being {
		playersLock.RUnlock()
		return player, true, nil
	}

	playersLock.RUnlock()

	player = &PlayerData{}
	if err := mysql.Where("token = ?", token).First(player).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, false, nil
		} else {
			return nil, false, err
		}
	}

	playersLock.Lock()

	playersByPlayer[player.Id] = player
	playersByUnionId[player.UnionId] = player
	playersByToken[player.Token] = player

	playersLock.Unlock()

	return player, true, nil
}

// 根据微信UID 查询玩家
func QueryPlayerByWechatUid(uid string) (*PlayerData, bool, error) {
	playersLock.RLock()

	player, being := playersByUnionId[uid]
	if being {
		playersLock.RUnlock()
		return player, true, nil
	}

	playersLock.RUnlock()

	player = &PlayerData{}
	if err := mysql.Where("union_id = ?", uid).First(player).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, false, nil
		} else {
			return nil, false, err
		}
	}

	playersLock.Lock()

	playersByPlayer[player.Id] = player
	playersByUnionId[player.UnionId] = player
	playersByToken[player.Token] = player

	playersLock.Unlock()

	return player, true, nil
}

// 分享送钻
func PlayerShared(id Player) (int32, error) {
	playerData, being, err := QueryPlayerById(id)
	if err != nil {
		return 0, err
	}
	if !being {
		return 0, ErrPlayerNotFound
	}

	year, month, day := playerData.SharedAt.Date()
	yearNow, monthNow, dayNow := time.Now().Date()
	if yearNow <= year {
		if monthNow <= month {
			if dayNow <= day {
				return 0, nil
			}
		}
	}

	var changed []Player
	var modifies []*modifyDiamondsAction

	modifies = append(modifies, &modifyDiamondsAction{
		Player: id,
		Number: conf.Option.Hall.ShareDiamonds,
		After: func(ts *gorm.DB, modify *modifyDiamondsAction) error {
			if err := ts.Model(new(PlayerData)).Where("id = ?", id).Updates(&PlayerData{
				SharedAt: time.Now(),
			}).Error; err != nil {
				return err
			}
			return nil
		},
	})
	changed = append(changed, id)

	ts := mysql.Begin()

	err = applyModifyDiamondsAction(ts, modifies)
	if err != nil {
		ts.Rollback()
		return 0, err
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

	return conf.Option.Hall.ShareDiamonds, nil
}

//刷新缓存
func RefreshPlayer(player Player) {
	playersLock.Lock()

	playerData, being := playersByPlayer[player]
	if being {
		delete(playersByPlayer, player)
		delete(playersByUnionId, playerData.UnionId)
		delete(playersByToken, playerData.Token)
	}

	playersLock.Unlock()
}
