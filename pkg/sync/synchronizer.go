package sync

import (
	"context"
	"database/sql"
	"slices"
	"strconv"
	"strings"
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
	AchivementIDRealmAvenger   = 283
	AchivementIDRealmAvengerIX = 7912
)

type Service struct {
	gw2API *gw2api.Session
	em     *verify.EventEmitter
}

func NewService(gw2API *gw2api.Session, em *verify.EventEmitter) *Service {
	return &Service{
		gw2API: gw2API,
		em:     em,
	}
}

// Start starts a synchronization loop that will continuously fetch the oldest updated API key
// and synchronize it with the gw2 api
func (s *Service) Start() {
	var failureCount int
	var successCount = 0
	var successTimestamp = time.Now()
	for {
		func() {
			defer func() {
				if err := recover(); err != nil {
					zap.L().Error("panic occurred:", zap.Any("recovered", err))
				}
			}()

			// Throttle if we experience excessive failures
			if failureCount >= 10 {
				zap.L().Warn("10 consecutive failures, sleeping for 10 seconds")
				time.Sleep(10 * time.Second)
			}
			// if successful, this will be reset to zero
			failureCount++

			err := s.SynchronizeNextAPIKey(orm.DB())
			if err != nil {
				zap.L().Error("unable to sync api key", zap.Error(err))
				return
			}

			// reset failure counter
			failureCount = 0

			// Print basic performance number every 10th minute
			successCount++
			if time.Since(successTimestamp).Minutes() >= 10 {
				zap.L().Info("statistics past 10 minutes",
					zap.Int("successes", successCount),
					zap.Int("current consecutive failures", failureCount))
				successTimestamp = time.Now()
				successCount = 0
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
			time.Sleep(5 * time.Second)
			return nil
		}
		return err
	}

	// Synchronize all data available with the api key
	acc, err := s.SynchronizeAPIKey(tx, s.gw2API, &token)
	if err != nil {
		// Handle failed token
		err = s.HandleFailedTokenInfo(&token, acc, err)
		time.Sleep(5 * time.Second)
		return err
	}

	return nil
}

func (s *Service) HandleFailedTokenInfo(token *orm.TokenInfo, acc *orm.Account, err error) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	token.UpdateLastAttemptedUpdate()

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
	if config.Config().DeleteDataAfterDay != nil {
		// Is the data old?
		expTime := *config.Config().DeleteDataAfterDay * (time.Hour * 24)
		if time.Since(token.LastSuccess) > expTime {
			var isBanned bool
			var storedAcc orm.Account
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

func (s *Service) SynchronizeAPIKey(tx bun.IDB, gw2API *gw2api.Session, token *orm.TokenInfo) (acc *orm.Account, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	// Fetch newest account data from gw2 api
	gw2Acc, err := gw2API.WithAccessToken(token.APIKey).Account()
	if err != nil {
		return acc, errors.WithStack(err)
	}
	acc = &orm.Account{
		Account: gw2Acc,
	}

	// Fetch persisted account data from database
	storedAcc := orm.Account{}
	err = tx.NewSelect().Model(&storedAcc).Where(`"id" = ?`, acc.ID).Scan(ctx)
	if err != nil && err != sql.ErrNoRows {
		return acc, errors.WithStack(err)
	}
	acc.UserID = storedAcc.UserID

	// Store any account changes in the history
	history.CollectAccount(storedAcc, gw2Acc)

	// Notify listeners if needed of verification changes (if any)
	var newAcc, oldAcc api.Account
	newAcc.FromGW2API(gw2Acc)
	oldAcc.FromGW2API(storedAcc.Account)
	if s.em.ShouldEmitAccount(&oldAcc, &newAcc) {
		var user api.User
		err = orm.QueryGetUser(tx, &user, acc.UserID).
			Scan(ctx)
		if err != nil {
			return acc, errors.WithStack(err)
		}
		s.em.Emit(&user)
	}

	// persist any changes made to the account data
	err = acc.Persist(tx)
	if err != nil {
		return acc, err
	}

	//Check if token metadata is missing
	if token.AccountID == "" || len(token.Permissions) <= 0 {
		token.AccountID = acc.ID
		token.LastSuccess = time.Now()

		//Retrieve token from api and persist it
		token.TokenInfo, err = gw2API.Tokeninfo()
		if err != nil {
			return acc, errors.WithStack(err)
		}

		err = token.Persist(tx)
		if err != nil {
			return acc, err
		}
	}

	// Synchronize activity from achivements
	var kills int
	if !slices.ContainsFunc(token.Permissions, func(val string) bool { return strings.Contains(val, "progression") }) {
		goto skipActivity
	}

	{ // Fetch achivements
		achivements, err := gw2API.AccountAchievements(AchivementIDRealmAvenger, AchivementIDRealmAvengerIX)
		if err != nil && err.Error() != "all ids provided are invalid" {
			zap.L().Error("unable to fetch account achivements", zap.Error(err))
			goto skipActivity
		}

		if len(achivements) > 0 {
			for _, achivement := range achivements {
				if achivement.Current > kills {
					kills = achivement.Current
				}
			}
		}
	}

	// Update activity
	err = history.UpdateActivity(tx, acc.ID, acc.WvWRank, kills)
	if err != nil {
		return acc, errors.WithStack(err)
	}
skipActivity:

	// update last success
	token.UpdateLastSuccessfulUpdate()

	if config.Config().Debug {
		zap.L().Info("updated account", zap.String("account name", acc.Name))
	}

	return acc, nil
}
