package lever28

import (
	"errors"
	"sort"
)

var (
	ErrIllegal = errors.New("illegal")
)

func GetMahjongType(mahjong []int32, banker bool) (w int32, err error) {
	if len(mahjong) != 2 {
		return 0, ErrIllegal
	}

	sort.Slice(mahjong, func(i, j int) bool {
		return mahjong[i] < mahjong[j]
	})

	judge := func(a, b, r int32) {
		if mahjong[0] == a && mahjong[1] == b {
			w = r
		}
	}

	// 计算对子
	judge(0, 0, 1090)
	judge(9, 9, 1080)
	judge(8, 8, 1070)
	judge(7, 7, 1060)
	judge(6, 6, 1050)
	judge(5, 5, 1040)
	judge(4, 4, 1030)
	judge(3, 3, 1020)
	judge(2, 2, 1010)
	judge(1, 1, 1000)

	// 计算二八杠
	judge(2, 8, 990)

	// 计算单点
	// 9
	judge(0, 9, 899)
	judge(1, 8, 898)
	judge(2, 7, 897)
	judge(3, 6, 896)
	judge(4, 5, 895)
	// 8
	judge(0, 8, 889)
	judge(1, 7, 888)
	judge(2, 6, 887)
	judge(3, 5, 886)
	// 7
	judge(0, 7, 879)
	judge(8, 9, 878)
	judge(1, 6, 877)
	judge(2, 5, 876)
	judge(3, 4, 875)
	// 6
	judge(0, 6, 869)
	judge(7, 9, 868)
	judge(1, 5, 867)
	judge(2, 4, 866)
	// 5
	judge(0, 5, 859)
	judge(6, 9, 858)
	judge(7, 8, 857)
	judge(1, 4, 856)
	judge(2, 3, 855)
	// 4
	judge(0, 4, 849)
	judge(5, 9, 848)
	judge(6, 8, 847)
	judge(1, 3, 846)
	// 3
	judge(0, 3, 839)
	judge(4, 9, 838)
	judge(5, 8, 837)
	judge(6, 7, 836)
	judge(1, 2, 835)
	// 2
	judge(0, 2, 829)
	judge(3, 9, 828)
	judge(4, 8, 827)
	judge(5, 7, 826)
	// 1
	judge(0, 1, 819)
	judge(2, 9, 818)
	judge(3, 8, 817)
	judge(4, 7, 816)
	judge(5, 6, 815)

	if w == 0 {
		w = 800
		if banker {
			w += 1
		}
	}

	return w, nil
}
