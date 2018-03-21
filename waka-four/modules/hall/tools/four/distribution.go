package four

import (
	"math"
	"math/rand"
	"sort"
	"sync"
	"time"

	"github.com/liuhan907/waka/waka-four/database"
	"gopkg.in/ahmetb/go-linq.v3"
)

type Player struct {
	Player database.Player
	Weight int32
}

func Distributing(players []Player, roundNumber int32) (collections []map[database.Player][]string) {
	for _, selected := range distributeFutureRing(players, roundNumber) {
		var distributions []*mahjongDistributingT
		for _, mahjong := range Acquire4(len(players)) {
			b, w, s, p, err := SearchBestPattern(mahjong)
			if err != nil {
				panic(err)
			}

			distributions = append(distributions, &mahjongDistributingT{
				Mahjong: append(append([]string{}, b[0]...), b[1]...),
				Weight:  w,
				Score:   s,
				Pattern: p,
			})
		}

		sort.Slice(distributions, func(i, j int) bool {
			var s1 int32
			var s2 int32

			if distributions[i].Weight[0] > distributions[j].Weight[0] {
				s1 += distributions[i].Score[0]
				s2 -= distributions[i].Score[0]
			} else if distributions[i].Weight[0] < distributions[j].Weight[0] {
				s1 -= distributions[j].Score[0]
				s2 += distributions[j].Score[0]
			}

			if distributions[i].Weight[1] > distributions[j].Weight[1] {
				s1 += distributions[i].Score[1]
				s2 -= distributions[i].Score[1]
			} else if distributions[i].Weight[1] < distributions[j].Weight[1] {
				s1 -= distributions[j].Score[1]
				s2 += distributions[j].Score[1]
			}

			return s1 < s2
		})

		round := make(map[database.Player][]string, len(players))
		for _, player := range players {
			round[player.Player] = distributions[selected[player.Player]].Mahjong
		}

		collections = append(collections, round)
	}

	return collections
}

type mahjongDistributingT struct {
	Mahjong []string
	Weight  []int32
	Score   []int32
	Pattern []string
}

var (
	getModes func(n int) [][]int32
)

func init() {
	modes := make(map[int][][]int32, 7)
	for i := 2; i < 9; i++ {
		var origin []int32
		linq.Range(0, i).Select(func(in interface{}) interface{} { return int32(in.(int)) }).ToSlice(&origin)
		modes[i-2] = permutation(origin)
	}
	getModes = func(n int) [][]int32 {
		return modes[n-2]
	}
}

var (
	dealLock sync.Mutex
	dealRand = rand.New(rand.NewSource(time.Now().Unix()))
)

func distributeFutureRing(players []Player, roundNumber int32) (futureMap []map[database.Player]int32) {
	available := make([][]int32, 0, 8000)
	modes := getModes(len(players))
	numberMap := distributeNumberRing(players, roundNumber)

	form := make([][]int32, roundNumber)
	for i := range form {
		available = available[:0]
		for _, mode := range modes {
			can := true
			for k, c := range mode {
				if numberMap[k][c] == 0 {
					can = false
					break
				}
			}
			if can {
				available = append(available, mode)
			}
		}
		if len(available) > 0 {
			dealLock.Lock()
			selected := available[dealRand.Int63()%int64(len(available))]
			dealLock.Unlock()
			form[i] = selected
			for k, c := range selected {
				numberMap[k][c]--
			}
		}
	}

	dealLock.Lock()
	dealRand.Shuffle(len(form), func(i, j int) {
		form[i], form[j] = form[j], form[i]
	})
	dealLock.Unlock()

	for _, v := range form {
		m := make(map[database.Player]int32, len(players))
		for k, c := range v {
			m[players[k].Player] = c
		}
		futureMap = append(futureMap, m)
	}

	return futureMap
}

func distributeNumberRing(players []Player, roundNumber int32) (numberMap [][]int32) {
	numberMap = make([][]int32, len(players))
	for k := range players {
		numberMap[k] = make([]int32, len(players))
	}

	weights := int32(linq.From(players).Select(func(in interface{}) interface{} { return int64(in.(Player).Weight) }).SumInts())
	for i := int32(len(players) - 1); i > 0; i-- {
		for k := range players {
			playerR := float64(players[k].Weight) / float64(weights)
			remainderR := float64(roundNumber-int32(linq.From(numberMap[k]).SumInts())) / float64(roundNumber)
			n := int32(playerR*remainderR*float64(roundNumber) + 0.5)
			numberMap[k][i] = n
		}

		for {
			for {
				number := int32(linq.From(numberMap).Select(func(in interface{}) interface{} { return int64((in.([]int32))[i]) }).SumInts())
				if number >= roundNumber {
					break
				}

				var player int
				var max int32
				for k, v := range numberMap {
					if t := roundNumber - int32(linq.From(v).SumInts()); t > max {
						player = k
						max = t
					}
				}
				numberMap[player][i]++
			}

			for {
				number := int32(linq.From(numberMap).Select(func(in interface{}) interface{} { return int64((in.([]int32))[i]) }).SumInts())
				if number <= roundNumber {
					break
				}

				var player int
				var min int32 = math.MaxInt32
				for k, v := range numberMap {
					if t := roundNumber - int32(linq.From(v).SumInts()); t < min {
						player = k
						min = t
					}
				}
				numberMap[player][i]--
			}

			number := int32(linq.From(numberMap).Select(func(in interface{}) interface{} { return int64((in.([]int32))[i]) }).SumInts())
			if number == roundNumber {
				break
			}
		}
	}

	for _, v := range numberMap {
		v[0] = roundNumber - int32(linq.From(v).SumInts())
	}

	return numberMap
}

func permutation(origin []int32) (permutations [][]int32) {
	if len(origin) == 1 {
		return [][]int32{{origin[0]}}
	}

	for i := range origin {
		first := origin[i]
		last := append(append([]int32{}, origin[:i]...), origin[i+1:]...)
		for _, p := range permutation(last) {
			permutations = append(permutations, append(append([]int32{}, first), p...))
		}
	}

	return permutations
}
