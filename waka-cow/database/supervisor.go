package database

import (
	"sync"

	"github.com/jinzhu/gorm"
)

// 代理
type Supervisor int32

func (supervisor Supervisor) SupervisorData() *SupervisorData {
	if supervisor == 0 {
		return &SupervisorData{
			Ref:       100000,
			Player:    100000,
			BonusRate: 30,
		}
	}

	supervisorData, being, err := QuerySupervisorByRef(supervisor)
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

// 代理数据
type SupervisorData struct {
	// 主键
	Ref Supervisor `gorm:"index;primary_key;AUTO_INCREMENT"`
	// 代理的玩家Ref
	Player Player `gorm:"unique;index"`

	// 分成比率
	// 按百分比的100 倍记录
	BonusRate int32
	// 房间基本分列表
	// 内部使用json 记录
	BaseScores Int32SliceSQLField `gorm:"type:text"`
	// 房间最大数量
	MaxRoomNumber int32
}

func (SupervisorData) TableName() string {
	return "supervisors"
}

// ---------------------------------------------------------------------------------------------------------------------

var (
	supervisorsLock         sync.RWMutex
	supervisorsBySupervisor = make(map[Supervisor]*SupervisorData)
	supervisorsByPlayer     = make(map[Player]*SupervisorData)
)

// ---------------------------------------------------------------------------------------------------------------------

// 根据 Ref 查询代理
func QuerySupervisorByRef(ref Supervisor) (*SupervisorData, bool, error) {
	supervisorsLock.RLock()

	supervisor, being := supervisorsBySupervisor[ref]
	if being {
		supervisorsLock.RUnlock()
		return supervisor, true, nil
	}

	supervisorsLock.RUnlock()

	supervisor = &SupervisorData{
		Ref: ref,
	}
	if err := mysql.First(supervisor).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, false, nil
		} else {
			return nil, false, err
		}
	}

	supervisorsLock.Lock()

	supervisorsBySupervisor[supervisor.Ref] = supervisor
	supervisorsByPlayer[supervisor.Player] = supervisor

	supervisorsLock.Unlock()

	return supervisor, true, nil
}

// 根据 Player 查询代理
func QuerySupervisorByPlayer(ref Player) (*SupervisorData, bool, error) {
	supervisorsLock.RLock()

	supervisor, being := supervisorsByPlayer[ref]
	if being {
		supervisorsLock.RUnlock()
		return supervisor, true, nil
	}

	supervisorsLock.RUnlock()

	supervisor = &SupervisorData{}
	if err := mysql.Where("player = ?", ref).First(supervisor).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, false, nil
		} else {
			return nil, false, err
		}
	}

	supervisorsLock.Lock()

	supervisorsBySupervisor[supervisor.Ref] = supervisor
	supervisorsByPlayer[supervisor.Player] = supervisor

	supervisorsLock.Unlock()

	return supervisor, true, nil
}

// 获取代理列表
func QuerySupervisorList() ([]*SupervisorData, error) {
	var supervisors []*SupervisorData
	if err := mysql.Find(&supervisors).Error; err != nil {
		return nil, err
	}

	supervisorsLock.Lock()

	for _, supervisor := range supervisors {
		supervisorsBySupervisor[supervisor.Ref] = supervisor
		supervisorsByPlayer[supervisor.Player] = supervisor
	}

	supervisorsLock.Unlock()

	return supervisors, nil
}

// 刷新缓存
func RefreshSupervisor(supervisor Supervisor) {
	supervisorsLock.Lock()

	supervisorData, being := supervisorsBySupervisor[supervisor]
	if being {
		delete(supervisorsBySupervisor, supervisor)
		delete(supervisorsByPlayer, supervisorData.Player)
	}

	supervisorsLock.Unlock()
}
