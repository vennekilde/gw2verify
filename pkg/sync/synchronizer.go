package sync

import (
	"context"
	"database/sql"
	"slices"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/pkg/errors"
	"github.com/uptrace/bun"
	"github.com/vennekilde/gw2verify/v2/internal/api"
	"github.com/vennekilde/gw2verify/v2/internal/config"
	"github.com/vennekilde/gw2verify/v2/internal/orm"
	"github.com/vennekilde/gw2verify/v2/pkg/history"
	"github.com/vennekilde/gw2verify/v2/pkg/verify"
	"go.uber.org/zap"

	"github.com/MrGunflame/gw2api"
)

const (
	CustomAchievementIDWvWRank  = -1
	CustomAchievementIDPlayTime = -2
)

const (
	AchievementIDRealmAvenger     = 283
	AchivementIDRealmAvengerIX    = 7912
	AchievementIDSupplySpend      = 306
	AchievementIDDollyEscort      = 285
	AchievementIDDollyKill        = 288
	AchievementIDObjectiveCapture = 303
	AchievementIDObjectiveDefend  = 319
	AchievementIDCampCapture      = 291
	AchievementIDCampDefend       = 310
	AchievementIDTowerCapture     = 297
	AchievementIDTowerDefend      = 322
	AchievementIDKeepCapture      = 300
	AchievementIDKeepDefend       = 316
	AchievementIDCastleCapture    = 294
	AchievementIDCastleDefend     = 313
)

var achievements = []int{
	AchievementIDRealmAvenger,
	AchivementIDRealmAvengerIX,
	AchievementIDSupplySpend,
	AchievementIDDollyEscort,
	AchievementIDDollyKill,
	AchievementIDObjectiveCapture,
	AchievementIDObjectiveDefend,
	AchievementIDCampCapture,
	AchievementIDCampDefend,
	AchievementIDTowerCapture,
	AchievementIDTowerDefend,
	AchievementIDKeepCapture,
	AchievementIDKeepDefend,
	AchievementIDCastleCapture,
	AchievementIDCastleDefend,
}

type Service struct {
	pool sync.Pool
	em   *verify.EventEmitter
}

func NewService(em *verify.EventEmitter) *Service {
	return &Service{
		pool: sync.Pool{
			New: func() interface{} {
				return gw2api.New()
			},
		},
		em: em,
	}
}

func (s *Service) getGW2API() *gw2api.Session {
	return s.pool.Get().(*gw2api.Session)
}

func (s *Service) putGW2API(gw2API *gw2api.Session) {
	s.pool.Put(gw2API)
}

// Start starts a synchronization loop that will continuously fetch the oldest updated API key
// and synchronize it with the gw2 api
func (s *Service) Start() {
	var consecutiveFailureCount atomic.Int32
	var failureCount atomic.Int32
	var successCount atomic.Int32
	var successTimestamp = time.Now()

	var activeJobs atomic.Int32
	var syncJobsSkipped atomic.Int32
	conf := config.Config()
	for {
		func() {
			defer func() {
				if err := recover(); err != nil {
					zap.L().Error("panic occurred:", zap.Any("recovered", err))
				}
			}()

			// Throttle if we experience excessive failures
			if consecutiveFailureCount.Load() >= 10 {
				zap.L().Warn("10 consecutive failures, sleeping for 10 seconds")
				time.Sleep(10 * time.Second)
			}

			if activeJobs.Load() < conf.MaxConcurrentSyncs {
				activeJobs.Add(1)
				go func() {
					err := s.SynchronizeNextAPIKey(orm.DB())
					if err != nil {
						zap.L().Error("unable to sync api key", zap.Error(err))
						consecutiveFailureCount.Add(1)
						failureCount.Add(1)
						return
					}
					consecutiveFailureCount.Store(0)
					successCount.Add(1)
					activeJobs.Add(-1)
				}()
			} else {
				syncJobsSkipped.Add(1)
			}

			// Wait for next sync interval
			time.Sleep(conf.SyncInterval)

			// Print basic performance number every 10th minute
			if time.Since(successTimestamp).Minutes() >= 10 {
				zap.L().Info("statistics past 10 minutes",
					zap.Int32("successes", successCount.Load()),
					zap.Int32("failures", consecutiveFailureCount.Load()),
					zap.Int32("skips", syncJobsSkipped.Load()),
					zap.Int32("active jobs", activeJobs.Load()),
					zap.Int32("current consecutive failures", consecutiveFailureCount.Load()))
				successTimestamp = time.Now()
				successCount.Store(0)
				failureCount.Store(0)
				syncJobsSkipped.Store(0)
			}
		}()
	}
}

func (s *Service) SynchronizeNextAPIKey(tx bun.IDB) error {
	// Find next token to sync
	window := config.Config().ExpirationTime
	token, err := orm.FindLastUpdatedAPIKey(window)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		return err
	}

	err = token.UpdateLastAttemptedUpdate()
	if err != nil {
		return err
	}

	gw2API := s.getGW2API()
	defer s.putGW2API(gw2API)
	// Synchronize all data available with the api key
	acc, err := s.SynchronizeAPIKey(tx, gw2API, &token)
	if err != nil {
		// Handle failed token
		err = s.HandleFailedTokenInfo(&token, acc, err)
		return err
	}

	return nil
}

func (s *Service) HandleFailedTokenInfo(token *orm.TokenInfo, acc *api.Account, err error) error {
	ctx := context.Background()

	if acc != nil {
		zap.L().Error("could not synchronize apikey",
			zap.String("apikey", token.APIKey),
			zap.String("account name", acc.Name),
			zap.Error(err))
	} else {
		// Show error if in debug mode, or if error is not just an error, stating it is an invalid key
		showErr := !strings.Contains(err.Error(), "invalid key") || config.Config().Debug
		if showErr {
			zap.L().Error("could not synchronize apikey",
				zap.String("apikey", token.APIKey),
				zap.Error(err))
		}
	}

	// Should we delete old data?
	if config.Config().DeleteDataAfter != nil {
		// Is the data old?
		expTime := *config.Config().DeleteDataAfter
		if time.Since(token.LastSuccess) > expTime {
			var isBanned bool
			var storedAcc api.Account
			// Do we have acc data for the user?
			if acc != nil {
				// Check if user is banned before deleting data
				if token.AccountID != "" {
					_ = orm.DB().NewSelect().
						Model(&storedAcc).
						Where(`"id" = ?`, acc.ID).
						Scan(ctx)
					if storedAcc.ID != "" {
						isBanned = verify.GetBan(&storedAcc) != nil
					}
				}
			}

			count, err := orm.DB().NewSelect().
				Model(&orm.TokenInfo{}).
				Where("account_id = ?, last_success >= db_updated - interval '"+strconv.Itoa(config.Config().ExpirationTime)+" seconds' OR last_success IS NULL", acc.ID).
				Count(ctx)
			if err != nil {
				return err
			}

			if count == 0 && !isBanned {
				// delete expired data
				orm.DB().NewDelete().Model(&token).Exec(ctx)
				orm.DB().NewDelete().Model(&storedAcc).Exec(ctx)
			}
		}
	}
	return nil
}

func (s *Service) SynchronizeUser(tx bun.IDB, gw2API *gw2api.Session, userID int64) error {
	window := config.Config().ExpirationTime
	tokens, err := orm.FindUserAPIKeys(userID, window)
	if err != nil {
		return err
	}

	for _, token := range tokens {
		acc, err := s.SynchronizeAPIKey(tx, gw2API, &token)
		if err != nil {
			zap.L().Error("unable to synchronize user account",
				zap.Any("account", acc),
				zap.Any("token", token),
				zap.Error(err))
		}
	}

	return nil
}

func (s *Service) SynchronizeAPIKey(tx bun.IDB, gw2API *gw2api.Session, token *orm.TokenInfo) (newAcc *api.Account, err error) {
	ctx := context.Background()
	// Fetch newest account data from gw2 api
	gw2API = gw2API.WithAccessToken(token.APIKey)
	gw2Acc, err := gw2API.Account()
	if err != nil {
		return newAcc, errors.WithStack(err)
	}
	newAcc = &api.Account{}
	newAcc.FromGW2API(gw2Acc)

	// Fetch persisted account data from database
	oldAcc := api.Account{}
	err = tx.NewSelect().Model(&oldAcc).Where(`"id" = ?`, gw2Acc.ID).Scan(ctx)
	if err != nil && err != sql.ErrNoRows {
		return newAcc, errors.WithStack(err)
	}
	newAcc.UserID = oldAcc.UserID

	// Store any account changes in the history
	history.CollectAccount(oldAcc, gw2Acc)

	// Notify listeners if needed of verification changes (if any)
	if s.em.ShouldEmitAccount(&oldAcc, newAcc) {
		var user api.User
		err = orm.QueryGetUser(tx, &user, newAcc.UserID).
			Scan(ctx)
		if err != nil {
			return newAcc, errors.WithStack(err)
		}
		s.em.Emit(&user)
	}

	// Synchronize WvW data
	err = synchronizeAccountWvW(gw2API, newAcc, token.Permissions)
	if err != nil {
		return newAcc, errors.WithStack(err)
	}

	// persist any changes made to the account data
	err = newAcc.Persist(tx)
	if err != nil {
		return newAcc, err
	}

	//Check if token metadata is missing
	if token.AccountID == "" || len(token.Permissions) <= 0 {
		token.AccountID = newAcc.ID
		token.LastSuccess = time.Now()

		//Retrieve token from api and persist it
		token.TokenInfo, err = gw2API.Tokeninfo()
		if err != nil {
			return newAcc, errors.WithStack(err)
		}

		err = token.Persist(tx)
		if err != nil {
			return newAcc, err
		}
	}

	// Synchronize account achievements
	err = s.synchronizeAccountAchievements(tx, gw2API, token, newAcc)
	if err != nil {
		return newAcc, errors.WithStack(err)
	}

	// update last success
	err = token.UpdateLastSuccessfulUpdate()
	if err != nil {
		return newAcc, err
	}

	if config.Config().Debug {
		zap.L().Info("updated account", zap.String("account name", newAcc.Name))
	}

	return newAcc, nil
}

func synchronizeAccountWvW(gw2API *gw2api.Session, acc *api.Account, permissions []string) error {
	if !slices.ContainsFunc(permissions, func(val string) bool { return strings.Contains(val, "wvw") }) {
		return nil
	}

	accWvW, err := gw2API.AccountWvW()
	if err != nil {
		return errors.WithStack(err)
	}

	// Update WvW Team and Guild
	if accWvW.Team != 0 {
		acc.WvWTeamID = accWvW.Team
	}
	if accWvW.Guild == "" {
		// Differentiate between unassigned and no data
		accWvW.Guild = "unassigned"
	}
	acc.WvWGuildID = &accWvW.Guild
	return nil
}

func (s *Service) synchronizeAccountAchievements(tx bun.IDB, gw2API *gw2api.Session, token *orm.TokenInfo, acc *api.Account) error {
	// Synchronize achivements
	if slices.ContainsFunc(token.Permissions, func(val string) bool { return strings.Contains(val, "progression") }) {
		achivements, err := gw2API.AccountAchievements(achievements...)
		if err != nil && err.Error() != "all ids provided are invalid" {
			zap.L().Error("unable to fetch account achivements", zap.Error(err))
		} else {
			var kills int
			for _, achivement := range achivements {
				// Update Realm Avenger with highest kill count
				// It seems that they are not updated in order, so we need to keep track of the highest kill count
				if achivement.ID == AchievementIDRealmAvenger || achivement.ID == AchivementIDRealmAvengerIX {
					if achivement.Current > kills {
						kills = achivement.Current
					}
					continue
				}
				err = history.UpdateAchievement(tx, acc.ID, achivement.ID, achivement.Current)
				if err != nil {
					zap.L().Error("unable to update account achivement", zap.Error(err), zap.Any("achivement", achivement))
				}
			}
			err = history.UpdateAchievement(tx, acc.ID, AchievementIDRealmAvenger, kills)
			if err != nil {
				zap.L().Error("unable to update account achivement realm avenger", zap.Error(err))
			}
		}

		// Update WvW rank with fake achievement id
		err = history.UpdateAchievement(tx, acc.ID, CustomAchievementIDWvWRank, acc.WvWRank)
		if err != nil {
			zap.L().Error("unable to update account achivement wvw rank", zap.Error(err))
		}

		// Update playtime with fake achievement id
		err = history.UpdateAchievement(tx, acc.ID, CustomAchievementIDPlayTime, int(acc.Age))
		if err != nil {
			zap.L().Error("unable to update account achivement playtime", zap.Error(err))
		}
	}
	return nil
}
