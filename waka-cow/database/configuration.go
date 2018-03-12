package database

import (
	"sync"

	"github.com/liuhan907/waka/waka-cow/proto"
)

// 系统配置
type Configuration struct {
	// 主键
	Ref int32 `gorm:"index;primary_key;AUTO_INCREMENT"`
	// 类型
	// customer_service 客服
	//   val1 姓名
	//   val2 微信
	Type string
	// 值
	Value1 string `gorm:"column:val1;type:text"`
	Value2 string `gorm:"column:val2;type:text"`
	Value3 string `gorm:"column:val3;type:text"`
	Value4 string `gorm:"column:val4;type:text"`
}

// ---------------------------------------------------------------------------------------------------------------------

var (
	lock             sync.RWMutex
	notice           string
	noticeBig        string
	ios              bool
	payURL           string
	registerURL      string
	loginURL         string
	customerServices []*cow_proto.Welcome_Customer
)

// 获取滚动公告
func GetNotice() string {
	lock.RLock()
	defer lock.RUnlock()
	return notice
}

// 获取公告
func GetNoticeBig() string {
	lock.RLock()
	defer lock.RUnlock()
	return noticeBig
}

// 获取是否在 IOS 审核中
func GetIsIOSExamine() bool {
	lock.RLock()
	defer lock.RUnlock()
	return ios
}

// 获取支付链接
func GetPayURL() string {
	lock.RLock()
	defer lock.RUnlock()
	return payURL
}

// 获取注册链接
func GetRegisterURL() string {
	lock.RLock()
	defer lock.RUnlock()
	return registerURL
}

// 获取登录链接
func GetLoginURL() string {
	lock.RLock()
	defer lock.RUnlock()
	return loginURL
}

// 获取客服信息
func GetCustomerServices() []*cow_proto.Welcome_Customer {
	lock.RLock()
	defer lock.RUnlock()
	return customerServices
}

// 刷新配置
func RefreshConfiguration() error {
	v1, err := getNotice()
	if err != nil {
		return err
	}

	v7, err := getNoticeBig()
	if err != nil {
		return err
	}

	v2, err := getIsIOSExamine()
	if err != nil {
		return err
	}

	v3, err := getPayURL()
	if err != nil {
		return err
	}

	v4, err := getRegisterURL()
	if err != nil {
		return err
	}

	v5, err := getLoginURL()
	if err != nil {
		return err
	}

	v6, err := getCustomerServices()
	if err != nil {
		return err
	}

	lock.Lock()

	notice = v1
	noticeBig = v7
	ios = v2
	payURL = v3
	registerURL = v4
	loginURL = v5
	customerServices = v6

	lock.Unlock()

	return nil
}

func getNotice() (string, error) {
	val := &Configuration{}
	if err := mysql.Where("type = ?", "notice").First(&val).Error; err != nil {
		return "", err
	}
	return val.Value1, nil
}

func getNoticeBig() (string, error) {
	val := &Configuration{}
	if err := mysql.Where("type = ?", "big_notice").First(&val).Error; err != nil {
		return "", err
	}
	return val.Value1, nil
}

func getIsIOSExamine() (bool, error) {
	val := &Configuration{}
	if err := mysql.Where("type = ?", "ios").First(&val).Error; err != nil {
		return false, err
	}
	is := false
	if val.Value1 == "true" {
		is = true
	}
	return is, nil
}

func getPayURL() (string, error) {
	val := &Configuration{}
	if err := mysql.Where("type = ?", "pay_url").First(&val).Error; err != nil {
		return "", err
	}
	return val.Value1, nil
}

func getRegisterURL() (string, error) {
	val := &Configuration{}
	if err := mysql.Where("type = ?", "register_url").First(&val).Error; err != nil {
		return "", err
	}
	return val.Value1, nil
}

func getLoginURL() (string, error) {
	val := &Configuration{}
	if err := mysql.Where("type = ?", "login_url").First(&val).Error; err != nil {
		return "", err
	}
	return val.Value1, nil
}

func getCustomerServices() ([]*cow_proto.Welcome_Customer, error) {
	var vals []*Configuration
	if err := mysql.Where("type = ?", "customer_service").Find(&vals).Error; err != nil {
		return nil, err
	}
	var result []*cow_proto.Welcome_Customer
	for _, val := range vals {
		result = append(result, &cow_proto.Welcome_Customer{
			Name:   val.Value1,
			Wechat: val.Value2,
		})
	}
	return result, nil
}
