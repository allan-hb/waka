package database

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/liuhan907/waka/waka-cow/conf"
)

var (
	log = logrus.WithFields(logrus.Fields{
		"pid":    os.Getpid(),
		"module": "cow.database",
	})

	mysql *gorm.DB
)

func init() {
	db, err := gorm.Open("mysql", fmt.Sprintf(`%s:%s@tcp(%s)/%s?charset=utf8mb4&parseTime=True&loc=Local`,
		conf.Option.Database.User, conf.Option.Database.Password, conf.Option.Database.Host, conf.Option.Database.Name))
	if err != nil {
		log.Fatalln(err)
	}

	mysql = db

	mysql.LogMode(true)

	tables := []interface{}{
		new(PlayerData),
		new(FreezeData),
		new(SupervisorData),
		new(SupervisorRoomBlacklist),
		new(CowRoomPurchaseHistory),
		new(CowGoldRoomPurchaseHistory),
		new(GomokuPurchaseHistory),
		new(BonusHistory),
		new(CowWarHistory),
		new(GomokuWarHistory),
		new(Lever28WarHistory),
		new(RedWarHistory),
		new(Configuration),
	}
	if conf.Option.Install.Reset {
		if err := mysql.DropTableIfExists(tables...).Error; err != nil {
			log.Panic(err)
		}
		if err := mysql.CreateTable(tables...).Error; err != nil {
			log.Panic(err)
		}
		if err := mysql.Exec("alter table players AUTO_INCREMENT = 100000;").Error; err != nil {
			log.Panic(err)
		}
		if err := mysql.Create(&PlayerData{
			Nickname:  "__system",
			CreatedAt: time.Now(),
		}).Error; err != nil {
			log.Panic(err)
		}
	}
	if conf.Option.Install.Update {
		if err := mysql.AutoMigrate(tables...).Error; err != nil {
			log.Panic(err)
		}
		systemPlayerCount := 0
		if err := mysql.Model(new(PlayerData)).Where("ref = ?", 100000).Count(&systemPlayerCount).Error; err != nil {
			log.Panic(err)
		}
		if systemPlayerCount == 0 {
			if err := mysql.Create(&PlayerData{
				Ref:       100000,
				Nickname:  "__system",
				Vip:       time.Now(),
				CreatedAt: time.Now(),
			}).Error; err != nil {
				log.Panic(err)
			}
		}
	}

	recoverFreezeMoneyAfterLast()
	RefreshConfiguration()
	RefreshSupervisorRoomBlacklist()
	QuerySupervisorList()
}

type Int32SliceSQLField []int32

func (i32 Int32SliceSQLField) Value() (driver.Value, error) {
	bytes, err := json.Marshal(i32)
	return string(bytes), err
}

func (i32 *Int32SliceSQLField) Scan(input interface{}) error {
	switch value := input.(type) {
	case string:
		return json.Unmarshal([]byte(value), i32)
	case []byte:
		return json.Unmarshal(value, i32)
	default:
		return errors.New("not supported")
	}
}
