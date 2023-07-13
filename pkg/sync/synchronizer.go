package sync

import (
	"context"
	"database/sql"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/uptrace/bun"
	"github.com/vennekilde/gw2verify/internal/config"
	"github.com/vennekilde/gw2verify/internal/orm"
	"github.com/vennekilde/gw2verify/pkg/history"
	"github.com/vennekilde/gw2verify/pkg/verify"
	"go.uber.org/zap"

	"gitlab.com/MrGunflame/gw2api"
)

// StartAPISynchronizer starts a synchronization loop that will continuesly fetch the oldest updated API key
// and synchronize it with the gw2 api
func StartAPISynchronizer(gw2API *gw2api.Session) {
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

			tx, err := orm.DB().Begin()
			if err != nil {
				zap.L().Error("unable to begin transaction", zap.Error(err))
				return
			}

			err = SynchronizeNextAPIKey(tx)
			if err != nil {
				tx.Rollback()
				zap.L().Error("unable to sync api key", zap.Error(err))
				return
			}

			err = tx.Commit()
			if err != nil {
				zap.L().Error("unable to commit transaction", zap.Error(err))
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

func SynchronizeNextAPIKey(tx bun.Tx) error {
	// Find next token to sync
	window := config.Config().ExpirationTime
	token, err := orm.FindLastUpdatedAPIKey(window)
	if err != nil {
		return err
	}

	gw2API := gw2api.New()
	// Synchronize all data available with the api key
	acc, err := SynchronizeAPIKey(tx, gw2API, &token)
	if err != nil {
		// Handle failed token
		HandleFailedTokenInfo(&token, acc, err)
		return err
	}

	return nil
}

func HandleFailedTokenInfo(token *orm.TokenInfo, acc *orm.Account, err error) {
	ctx := context.Background()
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

			var stillValidTokens []gw2api.TokenInfo
			orm.DB().NewSelect().
				Model(&stillValidTokens).
				Order("db_updated").
				Where("account_id = ?, last_success >= db_updated - interval '"+strconv.Itoa(config.Config().ExpirationTime)+" seconds' OR last_success IS NULL", acc.ID).
				Limit(1).
				Scan(ctx)

			if len(stillValidTokens) == 0 && !isBanned {
				// delete expired data
				orm.DB().NewDelete().Model(&token).Exec(ctx)
				orm.DB().NewDelete().Model(&storedAcc).Exec(ctx)
			}
		}
	}
}

func SynchronizeUser(tx bun.Tx, gw2API *gw2api.Session, userID int) error {
	window := config.Config().ExpirationTime
	tokens, err := orm.FindUserAPIKeys(userID, window)
	if err != nil {
		return err
	}

	for _, token := range tokens {
		acc, err := SynchronizeAPIKey(tx, gw2API, &token)
		if err != nil {
			zap.L().Error("unable to synchronize user account",
				zap.Any("account", acc),
				zap.Any("token", token),
				zap.Error(err))
		}
	}

	return nil
}

func SynchronizeAPIKey(tx bun.Tx, gw2API *gw2api.Session, token *orm.TokenInfo) (acc *orm.Account, err error) {
	ctx := context.Background()
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
	verify.CheckForVerificationUpdate(storedAcc, gw2Acc)

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

	// update last success
	token.UpdateLastSuccessfulUpdate()

	if config.Config().Debug {
		zap.L().Info("updated account", zap.String("account name", acc.Name))
	}

	return acc, nil
}

func SynchronizeLinkedUser(tx bun.Tx, gw2API *gw2api.Session, serviceID int, serviceUserID string) (err error, userErr error) {
	link, err := orm.GetServiceLink(serviceID, serviceUserID)
	if err != nil {
		return err, nil
	}
	if link.UserID == 0 {
		err = errors.New("no service link found with that user id and service id")
		return err, err
	}

	err = SynchronizeUser(tx, gw2API, link.UserID)
	return nil, err
}
