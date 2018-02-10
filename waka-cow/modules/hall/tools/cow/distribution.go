package cow

import (
	"math/rand"
	"sort"
	"sync"
	"time"

	"github.com/liuhan907/waka/waka-cow/database"
)

type distributingT struct {
	Pokers []string
	Weight int32
}

func DistributingOnce(king database.Player, players []database.Player, mode int32) (collection map[database.Player][]string) {
	sort.Slice(players, func(i, j int) bool {
		if players[i] == king {
			return true
		}

		if players[j] == king {
			return false
		}

		return players[i] < players[j]
	})

	var distributions []*distributingT
	for _, pokers := range Acquire5(len(players)) {
		b, w, _, _, err := SearchBestPokerPattern(pokers, mode)
		if err != nil {
			panic(err)
		}

		distributions = append(distributions, &distributingT{
			Pokers: b,
			Weight: w,
		})
	}

	sort.Slice(distributions, func(i, j int) bool {
		return distributions[i].Weight < distributions[j].Weight
	})

	round := make(map[database.Player][]string, len(players))
	for i, player := range players {
		round[player] = distributions[i].Pokers
	}

	round[players[0]], round[players[len(players)-1]] = round[players[len(players)-1]], round[players[0]]

	return round
}

func Distributing(king database.Player, players []database.Player, roundNumber, victoryRate int32, mode int32) (collections []map[database.Player][]string) {
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
		var distributions []*distributingT
		for _, pokers := range Acquire5(len(players)) {
			b, w, _, _, err := SearchBestPokerPattern(pokers, mode)
			if err != nil {
				panic(err)
			}

			distributions = append(distributions, &distributingT{
				Pokers: b,
				Weight: w,
			})
		}

		sort.Slice(distributions, func(i, j int) bool {
			return distributions[i].Weight < distributions[j].Weight
		})

		round := make(map[database.Player][]string, len(players))
		for i, player := range players {
			round[player] = distributions[i].Pokers
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
