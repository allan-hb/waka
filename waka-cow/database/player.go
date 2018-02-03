package database

import (
	"sync"
	"time"

	"github.com/jinzhu/gorm"

	"github.com/liuhan907/waka/waka-cow/conf"
)

// 玩家
type Player int32

func (player Player) PlayerData() *PlayerData {
	if player == 0 {
		return &PlayerData{
			Nickname: "unknown",
		}
	}

	playerData, being, err := QueryPlayerByRef(player)
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

func (player Player) SupervisorData() *SupervisorData {
	if player == 0 {
		return &SupervisorData{
			Ref:       100000,
			Player:    100000,
			BonusRate: 30,
		}
	}

	supervisorData, being, err := QuerySupervisorByPlayer(player)
	if err != nil {
		return &SupervisorData{
			Ref:       100000,
			Player:    100000,
			BonusRate: 30,
		}
	}
	if !being {
		return &SupervisorData{
			Ref:       100000,
			Player:    100000,
			BonusRate: 30,
		}
	}
	return supervisorData
}

// 玩家数据
type PlayerData struct {
	// 主键
	Ref Player `gorm:"index;primary_key;AUTO_INCREMENT"`
	// 微信 UnionID
	UnionID string `gorm:"index;unique;column:union_id"`
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
	// 代理 ID
	Supervisor Player

	// 钱
	Money int32
	// VIP 时间
	Vip time.Time

	// 封禁
	Ban int32

	// 创建时间
	CreatedAt time.Time
}

func (PlayerData) TableName() string {
	return "players"
}

// ---------------------------------------------------------------------------------------------------------------------

var (
	playersLock      sync.RWMutex
	playersByPlayer  = make(map[Player]*PlayerData)
	playersByUnionID = make(map[string]*PlayerData)
	playersByToken   = make(map[string]*PlayerData)
)

// ---------------------------------------------------------------------------------------------------------------------

// 注册玩家
func RegisterPlayer(unionID, nickname string, head, token string) (*PlayerData, error) {
	player := &PlayerData{
		UnionID:   unionID,
		Token:     token,
		Nickname:  nickname,
		Head:      head,
		Money:     conf.Option.Hall.RegisterMoney,
		Vip:       time.Now(),
		Ban:       0,
		CreatedAt: time.Now(),
	}
	if err := mysql.Create(player).Error; err != nil {
		return nil, err
	}

	playersLock.Lock()

	playersByPlayer[player.Ref] = player
	playersByUnionID[player.UnionID] = player
	playersByToken[player.Token] = player

	playersLock.Unlock()

	return player, nil
}

// 更新玩家登录信息
func UpdatePlayerLogin(ref Player, nickname string, head, token string) error {
	if err := mysql.Model(&PlayerData{}).Where("ref = ?", ref).Update(&PlayerData{
		Nickname: nickname,
		Head:     head,
		Token:    token,
	}).Error; err != nil {
		return err
	}

	playersLock.Lock()

	player, being := playersByPlayer[ref]
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

// 更新玩家代理信息
func UpdatePlayerSupervisor(ref, supervisor Player) error {
	if err := mysql.Model(&PlayerData{}).Where("ref = ?", ref).Updates(&PlayerData{
		Supervisor: supervisor,
	}).Error; err != nil {
		return err
	}

	playersLock.Lock()

	player, being := playersByPlayer[ref]
	if being {
		player.Supervisor = supervisor
	}

	playersLock.Unlock()

	return nil
}

// 更新玩家附加信息
func UpdatePlayerExt(ref Player, wechat, name, idcard string) error {
	if err := mysql.Model(&PlayerData{}).Where("ref = ?", ref).Updates(&PlayerData{
		Wechat: wechat,
		Name:   name,
		Idcard: idcard,
	}).Error; err != nil {
		return err
	}

	playersLock.Lock()

	player, being := playersByPlayer[ref]
	if being {
		player.Wechat = wechat
		player.Name = name
		player.Idcard = idcard
	}

	playersLock.Unlock()

	return nil
}

// 根据 Ref 查询玩家
func QueryPlayerByRef(ref Player) (*PlayerData, bool, error) {
	playersLock.RLock()

	player, being := playersByPlayer[ref]
	if being {
		playersLock.RUnlock()
		return player, true, nil
	}

	playersLock.RUnlock()

	player = &PlayerData{}
	if err := mysql.Where("ref = ?", ref).First(player).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, false, nil
		} else {
			return nil, false, err
		}
	}

	playersLock.Lock()

	playersByPlayer[ref] = player
	playersByUnionID[player.UnionID] = player
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

	playersByPlayer[player.Ref] = player
	playersByUnionID[player.UnionID] = player
	playersByToken[player.Token] = player

	playersLock.Unlock()

	return player, true, nil
}

// 根据微信UID 查询玩家
func QueryPlayerByWechatUID(uid string) (*PlayerData, bool, error) {
	playersLock.RLock()

	player, being := playersByUnionID[uid]
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

	playersByPlayer[player.Ref] = player
	playersByUnionID[player.UnionID] = player
	playersByToken[player.Token] = player

	playersLock.Unlock()

	return player, true, nil
}

//刷新缓存
func RefreshPlayer(player Player) {
	playersLock.Lock()

	playerData, being := playersByPlayer[player]
	if being {
		delete(playersByPlayer, player)
		delete(playersByUnionID, playerData.UnionID)
		delete(playersByToken, playerData.Token)
	}

	playersLock.Unlock()
}
