package conf

import (
	"fmt"
	"io/ioutil"

	"github.com/BurntSushi/toml"
)

type Mode struct {
	Mode string `toml:"mode"`
}

type Logger struct {
	Level uint32 `toml:"level"`
}

type Install struct {
	Reset  bool `toml:"reset"`
	Update bool `toml:"update"`
}

type Database struct {
	Host     string `toml:"host"`
	User     string `toml:"user"`
	Password string `toml:"password"`
	Name     string `toml:"name"`
}

type Listen struct {
	Gateway string `toml:"gateway"`
	Backend string `toml:"backend"`
}

type Hall struct {
	Salt          string `toml:"salt"`
	RegisterMoney int32  `toml:"register_money"`
	BindMoney     int32  `toml:"bind_money"`
}

type T struct {
	Mode     Mode     `toml:"mode"`
	Log      Logger   `toml:"log"`
	Install  Install  `toml:"install"`
	Database Database `toml:"database"`
	Gateway  Listen   `toml:"listen"`
	Hall     Hall     `toml:"hall"`
}

var (
	Option T
)

func init() {
	d, err := ioutil.ReadFile("conf.toml")
	if err != nil {
		panic(fmt.Sprintf("read config file \"conf.toml\" failed: %s\n", err.Error()))
	}

	_, err = toml.Decode(string(d), &Option)
	if err != nil {
		panic(fmt.Sprintf("decode config file \"conf.toml\" failed: %s\n", err.Error()))
	}
}
