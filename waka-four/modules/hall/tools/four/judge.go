package four

import (
	"errors"
	"sort"
	"strconv"

	"strings"

	"github.com/liuhan907/waka/waka-four/modules/hall/tools"
	"gopkg.in/ahmetb/go-linq.v3"
)

var (
	ErrIllegalMahjong = errors.New("illegal mahjong number")
)

// 至尊
// 对天
// 对地
// 对银
// 对狐
// 长对
// 短对
// 烂对
// 天杠
// 地杠
// 天九王
// 天牌 1 - 9
// 地牌 1 - 9
// 银牌 1 - 9
// 狐牌 1 - 9
// 长牌 1 - 9
// 短牌 1 - 9
// 烂   1 - 9
// 零点
func GetPattern(mahjongs []string) (weight int32, score int32, pattern string, e error) {
	if len(mahjongs) != 2 {
		return 0, 0, "", ErrIllegalMahjong
	}

	sort.Strings(mahjongs)

	highest := int32(0)
	double := int32(0)
	value := int32(0)
	rank := int32(0)

	switch {
	case
		strings.HasPrefix(mahjongs[0], "bamboo_3") && strings.HasPrefix(mahjongs[1], "character_6"):
		highest = 1
		score = 13
		pattern = "至尊"

	case
		strings.HasPrefix(mahjongs[0], "white") && strings.HasPrefix(mahjongs[1], "white"):
		double = 9
		score = 12
		pattern = "对天"
	case
		strings.HasPrefix(mahjongs[0], "dot_2") && strings.HasPrefix(mahjongs[1], "dot_2"):
		double = 8
		score = 12
		pattern = "对地"
	case
		strings.HasPrefix(mahjongs[0], "dot_8") && strings.HasPrefix(mahjongs[1], "dot_8"):
		double = 7
		score = 12
		pattern = "对银"
	case
		strings.HasPrefix(mahjongs[0], "bamboo_4") && strings.HasPrefix(mahjongs[1], "bamboo_4"):
		double = 6
		score = 12
		pattern = "对狐"
	case
		strings.HasPrefix(mahjongs[0], "dot_4") && strings.HasPrefix(mahjongs[1], "dot_4"),
		strings.HasPrefix(mahjongs[0], "bamboo_6") && strings.HasPrefix(mahjongs[1], "bamboo_6"),
		strings.HasPrefix(mahjongs[0], "green") && strings.HasPrefix(mahjongs[1], "green"):
		double = 5
		score = 12
		pattern = "长对"
	case
		strings.HasPrefix(mahjongs[0], "red") && strings.HasPrefix(mahjongs[1], "red"),
		strings.HasPrefix(mahjongs[0], "bamboo_7") && strings.HasPrefix(mahjongs[1], "bamboo_7"),
		strings.HasPrefix(mahjongs[0], "east") && strings.HasPrefix(mahjongs[1], "east"),
		strings.HasPrefix(mahjongs[0], "dot_6") && strings.HasPrefix(mahjongs[1], "dot_6"):
		double = 4
		score = 12
		pattern = "短对"
	case strings.HasPrefix(mahjongs[0], "dot_7") && strings.HasPrefix(mahjongs[1], "dot_7"),
		strings.HasPrefix(mahjongs[0], "bamboo_8") && strings.HasPrefix(mahjongs[1], "bamboo_8"),
		strings.HasPrefix(mahjongs[0], "bamboo_5") && strings.HasPrefix(mahjongs[1], "bamboo_5"),
		strings.HasPrefix(mahjongs[0], "bamboo_9") && strings.HasPrefix(mahjongs[1], "bamboo_9"):
		double = 3
		score = 12
		pattern = "烂对"

	case
		strings.HasPrefix(mahjongs[0], "bamboo_9") && strings.HasPrefix(mahjongs[1], "white"):
		value = 11
		score = 11
		pattern = "天九王"
	case
		strings.HasPrefix(mahjongs[0], "dot_8") && strings.HasPrefix(mahjongs[1], "white"):
		value = 10
		rank = 9
		score = 10
		pattern = "天杠"
	case
		strings.HasPrefix(mahjongs[0], "bamboo_8") && strings.HasPrefix(mahjongs[1], "white"):
		value = 10
		rank = 8
		score = 10
		pattern = "天杠"
	case
		strings.HasPrefix(mahjongs[0], "dot_8") && strings.HasPrefix(mahjongs[1], "dot_2"):
		value = 10
		rank = 7
		score = 10
		pattern = "地杠"
	case
		strings.HasPrefix(mahjongs[0], "bamboo_8") && strings.HasPrefix(mahjongs[1], "dot_2"):
		value = 10
		rank = 6
		score = 10
		pattern = "地杠"

	default:
		mw1 := MahjongWeight[mahjongs[0]]
		mw2 := MahjongWeight[mahjongs[1]]

		mw := mw1 + mw2

		if mw > 10 {
			mw -= 10
		}

		if mod := mw % 10; mod != 0 {
			value = mw
			score = mw
			rank, pattern = GetRanks(mahjongs)
			pattern += strconv.FormatInt(int64(mod), 10)
		} else {
			pattern = "零点"
		}
	}

	weight = highest*10000 + double*1000 + value*10 + rank

	return weight, score, pattern, nil
}

func SearchBestPattern(mahjong []string) (bests [][]string, weights []int32, scores []int32, patterns []string, e error) {
	if len(mahjong) != 4 {
		return nil, nil, nil, nil, ErrIllegalMahjong
	}

	bests = make([][]string, 2)
	weights = make([]int32, 2)
	scores = make([]int32, 2)
	patterns = make([]string, 2)

	for _, mahjong := range tools.Combination(mahjong, 2) {
		w, s, p, _ := GetPattern(mahjong)
		if w > weights[1] {
			bests[1] = mahjong
			weights[1] = w
			scores[1] = s
			patterns[1] = p
		}
	}

	linq.From(mahjong).WhereT(func(x string) bool {
		return x != bests[1][0] && x != bests[1][1]
	}).ToSlice(&bests[0])

	weights[0], scores[0], patterns[0], _ = GetPattern(bests[0])

	return bests, weights, scores, patterns, nil
}
