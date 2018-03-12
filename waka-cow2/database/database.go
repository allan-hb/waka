package database

import (
	"fmt"
	"os"
	"time"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/sirupsen/logrus"

	"github.com/liuhan907/waka/waka-cow2/conf"
)

var (
	log = logrus.WithFields(logrus.Fields{
		"pid":    os.Getpid(),
		"module": "cow2.database",
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
		new(CowOrderRoomPurchaseHistory), new(CowPayForAnotherRoomPurchaseHistory),
		new(CowWarHistory),
		new(Configuration),
		new(FriendData), new(AskData),
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
		if err := mysql.Model(new(PlayerData)).Where("id = ?", 100000).Count(&systemPlayerCount).Error; err != nil {
			log.Panic(err)
		}
		if systemPlayerCount == 0 {
			if err := mysql.Create(&PlayerData{
				Id:        100000,
				Nickname:  "__system",
				CreatedAt: time.Now(),
				SharedAt:  time.Date(2018, 1, 1, 0, 0, 0, 0, time.Now().Location()),
			}).Error; err != nil {
				log.Panic(err)
			}
		}
	}

	RefreshConfiguration()
}
