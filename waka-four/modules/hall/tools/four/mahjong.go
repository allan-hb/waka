package four

import (
	"math/rand"
	"sync"
	"time"
)

var (
	// 六万 三条                                13
	// 白板 白板                                12
	// 二筒 二筒                                12
	// 八筒 八筒                                12
	// 四条 四条                                12
	// 六条 六条 四筒 四筒 发   发               12
	// 七条 七条 东风 东风 红中 红中 六筒 六筒     12
	// 七筒 七筒 八条 八条 五条 五条 九条 九条     12
	//
	// 6   六万   character_6
	// 2   二筒   dot_2
	// 4   四筒   dot_4
	// 6   六筒   dot_6
	// 7   七筒   dot_7
	// 8   八筒   dot_8
	// 3   三条   bamboo_3
	// 4   四条   bamboo_4
	// 5   五条   bamboo_5
	// 6   六条   bamboo_6
	// 7   七条   bamboo_7
	// 8   八条   bamboo_8
	// 9   九条   bamboo_9
	// 10  红中   red
	// 10  发     green
	// 11  东风   east
	// 12  白板   white
	//
	// 至尊 > 其它
	// 对子 > 非对子
	// 相同对子比另一道，赢者全赢
	// 非对子比点数          x + y
	//   超过 10 减 10      x + y - 10
	//     特殊
	//       八条 | 八筒 白板 10
	//       八条 | 八筒 二筒 10
	//   其它 相加为 X0 都算 0 点
	// 点数相同看排位，上位赢下位
	//
	// 麻将
	Mahjong = []string{
		"character_6",
		"dot_2", "dot_2", "dot_4", "dot_4", "dot_6", "dot_6", "dot_7", "dot_7", "dot_8", "dot_8",
		"bamboo_3", "bamboo_4", "bamboo_4", "bamboo_5", "bamboo_5", "bamboo_6", "bamboo_6", "bamboo_7", "bamboo_7", "bamboo_8", "bamboo_8", "bamboo_9", "bamboo_9",
		"red", "red",
		"green", "green",
		"east", "east",
		"white", "white",
	}

	MahjongName = map[string]string{
		"character_6": "六万",
		"dot_2":       "二筒",
		"dot_4":       "四筒",
		"dot_6":       "六筒",
		"dot_7":       "七筒",
		"dot_8":       "八筒",
		"bamboo_3":    "三条",
		"bamboo_4":    "四条",
		"bamboo_5":    "五条",
		"bamboo_6":    "六条",
		"bamboo_7":    "七条",
		"bamboo_8":    "八条",
		"bamboo_9":    "九条",
		"red":         "红中",
		"green":       "发",
		"east":        "东风",
		"white":       "白板",
	}

	MahjongWeight = map[string]int32{
		"character_6": 6,
		"dot_2":       2,
		"dot_4":       4,
		"dot_6":       6,
		"dot_7":       7,
		"dot_8":       8,
		"bamboo_3":    3,
		"bamboo_4":    4,
		"bamboo_5":    5,
		"bamboo_6":    6,
		"bamboo_7":    7,
		"bamboo_8":    8,
		"bamboo_9":    9,
		"red":         10,
		"green":       10,
		"east":        1,
		"white":       2,
	}
)

func GetRank(mahjong string) (int32, string) {
	switch mahjong {
	case "white":
		return 9, "天牌"
	case "dot_2":
		return 8, "地牌"
	case "dot_8":
		return 7, "银牌"
	case "bamboo_4":
		return 6, "狐牌"
	case "bamboo_6", "dot_4", "green":
		return 5, "长牌"
	case "east", "bamboo_7", "red", "dot_6":
		return 4, "短牌"
	case "bamboo_5", "bamboo_8", "bamboo_9", "dot_7",
		"character_6", "bamboo_3":
		return 3, "烂牌"
	default:
		return 2, "unknown"
	}
}

func GetRanks(mahjongs []string) (rank int32, pattern string) {
	for _, mahjong := range mahjongs {
		if r, p := GetRank(mahjong); r > rank {
			rank, pattern = r, p
		}
	}
	return rank, pattern
}

var (
	devLock = sync.Mutex{}
	dev     = rand.New(rand.NewSource(time.Now().Unix()))
)

// 获取指定数量的4 张牌
func Acquire4(group int) [][]string {
	pool := make([]string, len(Mahjong))
	copy(pool, Mahjong)

	devLock.Lock()
	dev.Shuffle(len(pool), func(i, j int) {
		pool[i], pool[j] = pool[j], pool[i]
	})
	devLock.Unlock()

	if group*4 > len(pool) {
		panic("acquire too more mahjong numbers")
	}

	var result [][]string
	for i := 0; i < group; i++ {
		copied := append([]string{}, pool[:4]...)
		result = append(result, copied)
		pool = pool[4:]
	}
	return result
}
