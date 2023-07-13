package verify

import (
	"errors"
	"strconv"
	"time"

	"github.com/vennekilde/gw2verify/internal/api"
	"gitlab.com/MrGunflame/gw2api"
	"go.uber.org/zap"
)

type worldSyncError error

type LinkedWorlds map[string]api.WorldLinks

// Errors raised.
var (
	ErrWorldsNotSynced worldSyncError = errors.New("worlds are not synched")
)

var lastEndTime time.Time
var linkedWorlds LinkedWorlds
var isWorldLinksSynced = false

func BeginWorldLinksSyncLoop(gw2API *gw2api.Session) {
	for {
		zap.L().Info("synchronizing linked worlds")
		if err := SynchronizeWorldLinks(gw2API); err != nil {
			zap.L().Error("unable to synchronize matchup", zap.Error(err))
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

func SynchronizeWorldLinks(gw2API *gw2api.Session) error {
	matches, err := gw2API.WvWMatches()
	if err != nil {
		return err
	}

	// Sanity check before we go and reset world links before we actually have a new matchup
	if len(matches) > 0 {
		lw := createEmptyLinkedWorldsMap()
		// reset timer to avoid it not being changed by the loop
		lowestEndTime := time.Time{}
		foundWorlds := 0
		for _, match := range matches {
			zap.L().Info("matchup fetched",
				zap.Any("id", match.ID),
				zap.Any("endtime", match.EndTime),
				zap.Any("reds", match.AllWorlds["red"]),
				zap.Any("blues", match.AllWorlds["blue"]),
				zap.Any("greens", match.AllWorlds["green"]))

			// Persist world link
			lw.setWorldLinks(match.AllWorlds["red"])
			lw.setWorldLinks(match.AllWorlds["blue"])
			lw.setWorldLinks(match.AllWorlds["green"])
			// bump found world counter
			foundWorlds += len(match.AllWorlds["red"]) +
				len(match.AllWorlds["blue"]) +
				len(match.AllWorlds["green"])

			// Parse match end time
			matchEndTime, err := time.Parse(time.RFC3339, match.EndTime)
			if err != nil {
				zap.L().Error("unable to parse matchup end time", zap.Error(err))
				continue
			}

			if lowestEndTime.IsZero() || lowestEndTime.After(matchEndTime) {
				lowestEndTime = matchEndTime
			}
		}
		// Only update if we can find all worlds
		if foundWorlds >= len(WorldNames) {
			setMatchupLinks(lw, lowestEndTime)
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

func setMatchupLinks(lw LinkedWorlds, lowestEndTime time.Time) {
	linkedWorlds = lw
	lastEndTime = lowestEndTime
	isWorldLinksSynced = true
}

func createEmptyLinkedWorldsMap() LinkedWorlds {
	newLinkedWorlds := make(LinkedWorlds)
	for worldID := range WorldNames {
		newLinkedWorlds[strconv.Itoa(worldID)] = []int{}
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
		lw[strconv.Itoa(worldRefID)] = links
	}
}

func matchHasWorld(match gw2api.WvWMatch, worldID int) bool {
	for _, world := range match.AllWorlds["red"] {
		if world == worldID {
			return true
		}
	}
	for _, world := range match.AllWorlds["blue"] {
		if world == worldID {
			return true
		}
	}
	for _, world := range match.AllWorlds["green"] {
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
	return linkedWorlds[strconv.Itoa(worldPerspective)], err
}
func GetAllWorldLinks() LinkedWorlds {
	return linkedWorlds
}
