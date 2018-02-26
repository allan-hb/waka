package conf

import (
	"fmt"
	"io/ioutil"

	"github.com/BurntSushi/toml"
)

type Logger struct {
	LogLevel uint32 `toml:"log_level"`
	LogHeart bool   `toml:"log_heart"`
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

type Gateway struct {
	Listen4 string `toml:"listen4"`
}

type Backend struct {
	Listen4 string `toml:"listen4"`
}

type Hall struct {
	Salt             string `toml:"salt"`
	RegisterDiamonds int32  `toml:"register_diamonds"`
	BindDiamonds     int32  `toml:"bind_diamonds"`
	ShareDiamonds    int32  `toml:"share_diamonds"`
	MinPlayerNumber  int32  `toml:"min_player_number"`
}

type T struct {
	Log      Logger   `toml:"log"`
	Install  Install  `toml:"install"`
	Database Database `toml:"database"`
	Gateway  Gateway  `toml:"gateway"`
	Backend  Backend  `toml:"backend"`
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
