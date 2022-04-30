package verify

import (
	"errors"
	"time"

	"github.com/vennekilde/gw2apidb/pkg/gw2api"
	"go.uber.org/zap"
)

type worldSyncError error

type LinkedWorlds map[int][]int

// Errors raised.
var (
	ErrWorldsNotSynced worldSyncError = errors.New("worlds are not synched")
)

var lastEndTime time.Time
var linkedWorlds LinkedWorlds
var isWorldLinksSynced = false

func BeginWorldLinksSyncLoop(gw2API *gw2api.GW2Api) {
	for {
		zap.L().Info("synchronizing linked worlds")
		if err := SynchronizeWorldLinks(gw2API); err != nil {
			zap.L().Error("unable to synchronize matchup", zap.Error(err))
			return
		}

		if !lastEndTime.IsZero() {
			// Sleep until next match
			sleepUntil := time.Until(lastEndTime)
			zap.L().Info("synchronizing linked worlds once matchup is over",
				zap.Duration("synchronizing timer", sleepUntil),
				zap.Time("endtime", lastEndTime))
			// Sleep for at least a minute to not spam the api
			if sleepUntil < time.Minute {
				sleepUntil = time.Minute
			}
			time.Sleep(sleepUntil)
		} else {
			zap.L().Info("synchronizing linked worlds in 5 minutes")
			time.Sleep(time.Minute * 5)
		}
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

	// Sanity check before we go and reset world links before we actually have a new matchup
	if len(matches) > 0 {
		lw := createEmptyLinkedWorldsMap()
		// reset timer to avoid it not being changed by the loop
		lastEndTime = time.Time{}
		foundWorlds := 0
		for _, match := range matches {
			lw.setWorldLinks(match.AllWorlds.Red)
			lw.setWorldLinks(match.AllWorlds.Blue)
			lw.setWorldLinks(match.AllWorlds.Green)
			foundWorlds += len(match.AllWorlds.Red) +
				len(match.AllWorlds.Blue) +
				len(match.AllWorlds.Green)
			if lastEndTime.IsZero() || lastEndTime.After(match.EndTime) {
				lastEndTime = match.EndTime
			}
			zap.L().Info("matchup fetched",
				zap.Any("id", match.ID),
				zap.Any("endtime", match.EndTime),
				zap.Any("reds", match.AllWorlds.Red),
				zap.Any("blues", match.AllWorlds.Blue),
				zap.Any("greens", match.AllWorlds.Green))
		}
		// Only update if we can find all worlds
		if foundWorlds >= len(WorldNames) {
			linkedWorlds = lw
			isWorldLinksSynced = true
			zap.L().Info("Updated linked worlds", zap.Any("linked worlds", linkedWorlds))
		} else {
			zap.L().Warn("not updating linked worlds, did not find all worlds in matchups",
				zap.Int("total worlds", len(WorldNames)),
				zap.Int("found worlds", len(lw)),
			)
		}
	}
	return nil
}

func createEmptyLinkedWorldsMap() LinkedWorlds {
	newLinkedWorlds := make(LinkedWorlds)
	for worldID := range WorldNames {
		newLinkedWorlds[worldID] = []int{}
	}
	return newLinkedWorlds
}

func (lw LinkedWorlds) setWorldLinks(allWorlds []int) {
	for _, worldRefID := range allWorlds {
		links := []int{}
		for _, worldID := range allWorlds {
			if worldID != worldRefID {
				links = append(links, worldID)
			}
		}
		lw[worldRefID] = links
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
func GetAllWorldLinks() LinkedWorlds {
	return linkedWorlds
}
