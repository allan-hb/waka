package four

import (
	"math/rand"
	"sort"
	"sync"
	"time"

	"github.com/liuhan907/waka/waka-four/database"
)

type mahjongDistributingT struct {
	Mahjong []string
	Weight  []int32
	Score   []int32
	Pattern []string
}

func Distributing(king database.Player, players []database.Player, roundNumber, victoryRate int32) (collections []map[database.Player][]string) {
	sort.Slice(players, func(i, j int) bool {
		if players[i] == king {
			return true
		}

		if players[j] == king {
			return false
		}

		return players[i] < players[j]
	})

	for _, selected := range buildDistribution(roundNumber, victoryRate) {
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
		for i, player := range players {
			round[player] = distributions[i].Mahjong
		}

		if selected == 1 {
			round[players[0]], round[players[len(players)-1]] = round[players[len(players)-1]], round[players[0]]
		}

		collections = append(collections, round)
	}

	return collections
}

var (
	distributingLock sync.Mutex
	distributingRand = rand.New(rand.NewSource(time.Now().Unix()))
)

func buildDistribution(number, rate int32) []int32 {
	selectedNumber := int32(float64(number*rate)/100 + 0.5)
	distributingLock.Lock()
	selectedNumber += int32(distributingRand.Int()%3 - 1)
	distributingLock.Unlock()

	if selectedNumber < 0 {
		selectedNumber = 0
	}
	if selectedNumber > number {
		selectedNumber = number
	}

	distributionRaw := make([]int32, number)
	for i := int32(0); i < selectedNumber; i++ {
		distributionRaw[i] = 1
	}

	distributingLock.Lock()
	perm := distributingRand.Perm(len(distributionRaw))
	distributingLock.Unlock()

	distribution := make([]int32, number)
	for i, k := range perm {
		distribution[k] = distributionRaw[i]
	}

	return distribution
}
