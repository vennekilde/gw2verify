package verify

import (
	"errors"
	"time"

	"github.com/vennekilde/gw2apidb/pkg/gw2api"
	"go.uber.org/zap"
)

type worldSyncError error

// Errors raised.
var (
	ErrWorldsNotSynced worldSyncError = errors.New("worlds are not synched")
)

var lastEndTime time.Time
var linkedWorlds map[int][]int
var isWorldLinksSynced = false

func BeginWorldLinksSyncLoop(gw2API *gw2api.GW2Api) {
	for {
		if !lastEndTime.IsZero() {
			// Sleep until next match
			sleepUntil := time.Until(lastEndTime)
			zap.L().Info("synchronizing linked worlds once matchup is over", zap.Duration("duration left", sleepUntil), zap.Time("endtime", lastEndTime))
			time.Sleep(sleepUntil)
		}

		zap.L().Info("synchronizing linked worlds")
		if err := SynchronizeWorldLinks(gw2API); err != nil {
			zap.L().Error("unable to synchronize matchup", zap.Error(err))
		}
		time.Sleep(time.Minute * 5)
	}
}

func SynchronizeWorldLinks(gw2API *gw2api.GW2Api) error {
	matchIds, err := gw2API.Matches()
	if err != nil {
		return err
	}
	matches, err := gw2API.MatchIds(matchIds...)
	if err != nil {
		return err
	}
	resetWorldLinks()
	for _, match := range matches {
		setWorldLinks(match.AllWorlds.Red)
		setWorldLinks(match.AllWorlds.Blue)
		setWorldLinks(match.AllWorlds.Green)
		if lastEndTime.After(match.EndTime) {
			lastEndTime = match.EndTime
		}
		isWorldLinksSynced = true
		zap.L().Info("matchup fetched", zap.Any("matchup", match))
	}
	return nil
}

func resetWorldLinks() {
	isWorldLinksSynced = false
	linkedWorlds = make(map[int][]int)
	for worldID := range WorldNames {
		linkedWorlds[worldID] = []int{}
	}
}

func setWorldLinks(allWorlds []int) {
	for _, worldRefID := range allWorlds {
		links := []int{}
		for _, worldID := range allWorlds {
			if worldID != worldRefID {
				links = append(links, worldID)
			}
		}
		linkedWorlds[worldRefID] = links
	}
}

func matchHasWorld(match gw2api.Match, worldID int) bool {
	for _, world := range match.AllWorlds.Red {
		if world == worldID {
			return true
		}
	}
	for _, world := range match.AllWorlds.Blue {
		if world == worldID {
			return true
		}
	}
	for _, world := range match.AllWorlds.Green {
		if world == worldID {
			return true
		}
	}
	return false
}

func IsWorldLinksSynchronized() bool {
	return isWorldLinksSynced
}

func GetWorldLinks(worldPerspective int) (links []int, err error) {
	if !IsWorldLinksSynchronized() {
		return links, ErrWorldsNotSynced
	}
	return linkedWorlds[worldPerspective], err
}
func GetAllWorldLinks() map[int][]int {
	return linkedWorlds
}
