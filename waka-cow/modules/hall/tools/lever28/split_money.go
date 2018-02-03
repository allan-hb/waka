package lever28

import (
	"errors"
	"math/rand"
	"sort"
	"sync"
	"time"
)

var (
	ErrIllegalMoneyNumber = errors.New("illegal money or number")

	lock sync.Mutex
	rnd  = rand.New(rand.NewSource(time.Now().Unix()))
)

func SplitMoney(money, number int32) ([]int32, error) {
	if money < number {
		return nil, ErrIllegalMoneyNumber
	}

	table := map[int32]bool{}
	for len(table) != int(number-1) {
		lock.Lock()
		point := rnd.Int()%int(money-1) + 1
		lock.Unlock()
		table[int32(point)] = true
	}

	var points []int32
	for v := range table {
		points = append(points, v)
	}

	sort.Slice(points, func(i, j int) bool {
		return points[i] < points[j]
	})

	var result []int32
	for i, v := range points {
		if i == 0 {
			result = append(result, v)
		} else {
			result = append(result, points[i]-points[i-1])
		}
	}
	result = append(result, money-points[len(points)-1])

	return result, nil
}
