package cow

import "fmt"

var (
	rateMap = []map[string]int32{
		{
			"straight_flush": 11,
			"boom":           10,
			"5small":         9,
			"5flower":        8,
			"flush":          7,
			"straight":       6,
			"full_house":     5,
			"nn":             4,
			"n9":             3,
			"n8":             2,
			"n7":             2,
			"n6":             1,
			"n5":             1,
			"n4":             1,
			"n3":             1,
			"n2":             1,
			"n1":             1,
			"n0":             1,
		},
		{
			"straight_flush": 17,
			"boom":           16,
			"5small":         15,
			"5flower":        14,
			"flush":          13,
			"straight":       12,
			"full_house":     11,
			"nn":             10,
			"n9":             9,
			"n8":             8,
			"n7":             7,
			"n6":             6,
			"n5":             5,
			"n4":             4,
			"n3":             3,
			"n2":             2,
			"n1":             1,
			"n0":             1,
		},
	}
)

// boom 炸弹
// full_house 葫芦
// nn 牛牛
// n[1-9] 牛1 - 牛9
// n0 无牛
// straight_flush 同花顺
// flush 同花
// straight 顺子
// 计算模式
func GetPokersPattern(pokers []string, mode int32, additional bool) (weight int32, pattern string, rate int32, e error) {
	if !isLegalPokers(pokers) {
		return -1, "", -1, ErrIllegalPokers
	}

	maxValue, err := getMaxPokerValues(pokers)
	if err != nil {
		return -1, "", -1, err
	}

	maxSuit, err := getMaxPokerSuit(pokers)
	if err != nil {
		return -1, "", -1, err
	}

	if additional {
		straightFlush, err := isStraightFlush(pokers)
		if err != nil {
			return -1, "", -1, err
		}
		if straightFlush {
			return 90000 + 0 + maxValue*10 + maxSuit, "straight_flush", rateMap[mode]["straight_flush"], nil
		}

		boom, err := isBoom(pokers)
		if err != nil {
			return -1, "", -1, err
		}
		if boom {
			return 80000 + 0 + maxValue*10 + maxSuit, "boom", rateMap[mode]["boom"], nil
		}

		fiveSmallCow, err := isFiveSmallCow(pokers)
		if err != nil {
			return -1, "", -1, err
		}
		if fiveSmallCow {
			return 70000 + 0 + maxValue*10 + maxSuit, "5small", rateMap[mode]["5small"], nil
		}

		fiveFlowerCow, err := isFiveFlowerCow(pokers)
		if err != nil {
			return -1, "", -1, err
		}
		if fiveFlowerCow {
			return 60000 + 0 + maxValue*10 + maxSuit, "5flower", rateMap[mode]["5flower"], nil
		}

		flush, err := isFlush(pokers)
		if err != nil {
			return -1, "", -1, err
		}
		if flush {
			return 50000 + 0 + maxValue*10 + maxSuit, "flush", rateMap[mode]["flush"], nil
		}

		straight, err := isStraight(pokers)
		if err != nil {
			return -1, "", -1, err
		}
		if straight {
			return 40000 + 0 + maxValue*10 + maxSuit, "straight", rateMap[mode]["straight"], nil
		}

		fullHouse, err := isFullHouse(pokers)
		if err != nil {
			return -1, "", -1, err
		}
		if fullHouse {
			return 30000 + 0 + maxValue*10 + maxSuit, "full_house", rateMap[mode]["straight_flush"], nil
		}
	}

	cow, cowNumber, err := isCow(pokers)
	if err != nil {
		return -1, "", -1, err
	}
	if cow {
		if cowNumber == 0 {
			return 20000 + 0 + maxValue*10 + maxSuit, "nn", rateMap[mode]["nn"], nil
		} else {
			return 10000 + cowNumber*1000 + maxValue*10 + maxSuit, fmt.Sprintf("n%v", cowNumber), rateMap[mode][fmt.Sprintf("n%v", cowNumber)], nil
		}
	} else {
		return 0 + 0 + maxValue*10 + maxSuit, "n0", 1, nil
	}
}

// 搜索最佳模式
func SearchBestPokerPattern(pokers []string, mode int32, additional bool) (best []string, weight int32, pattern string, rate int32, e error) {
	if !isLegalPokers(pokers) {
		return nil, -1, "", -1, ErrIllegalPokers
	}

	for _, v := range permutations(pokers) {
		w, p, r, err := GetPokersPattern(v, mode, additional)
		if err != nil {
			return nil, -1, "", -1, err
		}
		if w > weight {
			best, weight, pattern, rate = v, w, p, r
		}
	}
	return
}

func permutations(slice []string) [][]string {
	var helper func([]string, int)
	res := [][]string{}

	helper = func(slice []string, n int) {
		if n == 1 {
			tmp := make([]string, len(slice))
			copy(tmp, slice)
			res = append(res, tmp)
		} else {
			for i := 0; i < n; i++ {
				helper(slice, n-1)
				if n%2 == 1 {
					tmp := slice[i]
					slice[i] = slice[n-1]
					slice[n-1] = tmp
				} else {
					tmp := slice[0]
					slice[0] = slice[n-1]
					slice[n-1] = tmp
				}
			}
		}
	}
	helper(slice, len(slice))
	return res
}
