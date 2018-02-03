package database

import "sync"

// 代理房间黑名单
// 存在名单中的玩家无法进入那个代理的房间
type SupervisorRoomBlacklist struct {
	// 主键
	Ref int32 `gorm:"index;primary_key;AUTO_INCREMENT"`
	// 代理
	Supervisor Supervisor
	// 玩家
	Player Player
}

// ---------------------------------------------------------------------------------------------------------------------

var (
	blacklistsByPlayerSupervisorLock sync.RWMutex
	blacklistsByPlayerSupervisor     map[uint64]bool
)

func querySupervisorRoomBlacklistList() ([]*SupervisorRoomBlacklist, error) {
	var blacklists []*SupervisorRoomBlacklist
	if err := mysql.Model(&SupervisorRoomBlacklist{}).Find(&blacklists).Error; err != nil {
		return nil, err
	}

	return blacklists, nil
}

// ---------------------------------------------------------------------------------------------------------------------

// 刷新代理房间黑名单缓存
func RefreshSupervisorRoomBlacklist() error {
	blacklists, err := querySupervisorRoomBlacklistList()
	if err != nil {
		return err
	}

	blacklistsByPlayerSupervisorLock.Lock()

	blacklistsByPlayerSupervisor = make(map[uint64]bool)
	for _, blacklist := range blacklists {
		key := uint64(blacklist.Player) | (uint64(blacklist.Supervisor) << 32)
		blacklistsByPlayerSupervisor[key] = true
	}

	blacklistsByPlayerSupervisorLock.Unlock()

	return nil
}

// 查询玩家能否进入代理房间
func QueryPlayerCanEnterSupervisorRoom(player Player, supervisor Supervisor) bool {
	key := uint64(player) | (uint64(supervisor) << 32)

	blacklistsByPlayerSupervisorLock.RLock()

	can := blacklistsByPlayerSupervisor[key]

	blacklistsByPlayerSupervisorLock.RUnlock()

	return can
}
