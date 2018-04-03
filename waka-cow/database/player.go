package database

import (
	"sync"
	"time"

	"github.com/jinzhu/gorm"

	"github.com/liuhan907/waka/waka-cow/conf"
)

const (
	DefaultSupervisor    = 100000
	DefaultVictoryWeight = 50
)

// 玩家
type Player int32

func (player Player) PlayerData() *PlayerData {
	playerData, being, err := QueryPlayerByPlayer(player)
	if err != nil || !being {
		return &PlayerData{
			Nickname:      "unknown",
			Supervisor:    DefaultSupervisor,
			VictoryWeight: DefaultVictoryWeight,
		}
	}
	return playerData
}

// 玩家数据
type PlayerData struct {
	// 主键
	Id Player `gorm:"index;unique;primary_key;AUTO_INCREMENT"`
	// 创建时间
	CreatedAt time.Time

	// 微信登陆返回的 unionid
	WechatUnionid string `gorm:"index;unique;column:wechat_unionid"`
	// 微信昵称
	Nickname string
	// 微信头像 URL
	Head string

	// 登陆令牌
	Token string `gorm:"index;unique"`

	// 钱
	Money int32
	// VIP 时间
	Vip time.Time

	// 微信号
	Wechat string
	// 姓名
	Name string
	// 身份证
	Idcard string

	// 上级代理
	Supervisor Player

	// 封禁
	Ban int32

	// 权重
	VictoryWeight int32
}

func (PlayerData) TableName() string {
	return "players"
}

// ---------------------------------------------------------------------------------------------------------------------

var (
	playersLock     sync.RWMutex
	playersById     = make(map[Player]*PlayerData)
	playersByWechat = make(map[string]*PlayerData)
	playersByToken  = make(map[string]*PlayerData)
)

// ---------------------------------------------------------------------------------------------------------------------

// 注册玩家
func RegisterPlayer(uid, nickname string, head, token string) (*PlayerData, error) {
	player := &PlayerData{
		CreatedAt:     time.Now(),
		WechatUnionid: uid,
		Nickname:      nickname,
		Head:          head,
		Token:         token,
		Money:         conf.Option.Hall.RegisterMoney,
		Vip:           time.Now(),
		Supervisor:    DefaultSupervisor,
		VictoryWeight: DefaultVictoryWeight,
	}
	if err := mysql.Create(player).Error; err != nil {
		return nil, err
	}

	playersLock.Lock()

	playersById[player.Id] = player
	playersByWechat[player.WechatUnionid] = player
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

	player, being := playersById[id]
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
func UpdatePlayerSupervisor(id, supervisor Player) error {
	if err := mysql.Model(&PlayerData{}).Where("id = ?", id).Updates(&PlayerData{
		Supervisor: supervisor,
	}).Error; err != nil {
		return err
	}

	playersLock.Lock()

	player, being := playersById[id]
	if being {
		player.Supervisor = supervisor
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

	player, being := playersById[id]
	if being {
		player.Wechat = wechat
		player.Name = name
		player.Idcard = idcard
	}

	playersLock.Unlock()

	return nil
}

// 根据 Player 查询玩家
func QueryPlayerByPlayer(id Player) (*PlayerData, bool, error) {
	playersLock.RLock()

	player, being := playersById[id]
	if being {
		playersLock.RUnlock()
		return player, true, nil
	}

	playersLock.RUnlock()

	player = &PlayerData{
		Id: id,
	}
	if err := mysql.First(player).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, false, nil
		} else {
			return nil, false, err
		}
	}

	playersLock.Lock()

	playersById[id] = player
	playersByWechat[player.WechatUnionid] = player
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

	playersById[player.Id] = player
	playersByWechat[player.WechatUnionid] = player
	playersByToken[player.Token] = player

	playersLock.Unlock()

	return player, true, nil
}

// 根据 WechatUnionid 查询玩家
func QueryPlayerByWechatUnionid(uid string) (*PlayerData, bool, error) {
	playersLock.RLock()

	player, being := playersByWechat[uid]
	if being {
		playersLock.RUnlock()
		return player, true, nil
	}

	playersLock.RUnlock()

	player = &PlayerData{}
	if err := mysql.Where("wechat_uid = ?", uid).First(player).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, false, nil
		} else {
			return nil, false, err
		}
	}

	playersLock.Lock()

	playersById[player.Id] = player
	playersByWechat[player.WechatUnionid] = player
	playersByToken[player.Token] = player

	playersLock.Unlock()

	return player, true, nil
}

// 删除缓存
func RefreshCache(player Player) {
	playersLock.Lock()

	playerData, being := playersById[player]
	if being {
		delete(playersById, player)
		delete(playersByWechat, playerData.WechatUnionid)
		delete(playersByToken, playerData.Token)
	}

	playersLock.Unlock()
}
