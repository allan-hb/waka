package tools

import (
	"math/rand"
	"sync"
	"time"
)

var (
	lock sync.Mutex
	r    = rand.New(rand.NewSource(time.Now().Unix()))
)

// 数字池
type NumberPool struct {
	random bool
	front  []int32
	behind []int32
}

// 获取
func (my *NumberPool) Acquire() (int32, bool) {
	if my.random {
		if len(my.front) == 0 {
			my.front, my.behind = my.behind, my.front
		}
	}

	if len(my.front) == 0 {
		return -1, false
	}

	c := my.front[len(my.front)-1]
	my.front = my.front[:len(my.front)-1]

	return c, true
}

// 返还
func (my *NumberPool) Return(c int32) {
	if my.random {
		my.behind = append(my.behind, c)
	} else {
		my.front = append(my.front, c)
	}
}

// 创建一个新的数字池
func NewNumberPool(start, number int32, random bool) *NumberPool {
	pool := &NumberPool{
		random: random,
		front:  nil,
		behind: make([]int32, 0, number),
	}

	front := make([]int32, number)
	for i := range front {
		front[i] = start + int32(i)
	}

	if random {
		shuffle := make([]int32, len(front))

		lock.Lock()
		perm := r.Perm(len(shuffle))
		lock.Unlock()

		for i, v := range perm {
			shuffle[v] = front[i]
		}

		front = shuffle
	}

	pool.front = front

	return pool
}
