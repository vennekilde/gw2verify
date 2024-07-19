package verify

import (
	"context"
	"database/sql"
	"time"

	"github.com/pkg/errors"
	"github.com/uptrace/bun"
	"github.com/vennekilde/gw2verify/v2/internal/api"
	"github.com/vennekilde/gw2verify/v2/internal/orm"
	"go.uber.org/zap"
)

func GetOrInsertUser(tx bun.Tx, platformID int, platformUserID string) (*api.User, error) {
	user := api.User{}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := tx.NewSelect().
		Model(&user).
		Join("INNER JOIN platform_links ON \"user\".id = platform_links.user_id").
		Where("platform_id = ? AND platform_user_id = ?", platformID, platformUserID).
		Scan(ctx)
	if err != nil && err != sql.ErrNoRows {
		return nil, errors.WithStack(err)
	}

	if err == sql.ErrNoRows {
		_, err = tx.NewInsert().
			Model(&user).
			Returning("*").
			Exec(ctx)
		if err != nil {
			return nil, errors.WithStack(err)
		}
	}

	return &user, nil
}

// SetOrReplacePlatformLink creates or replaces a service link between a service user and an account
func SetOrReplacePlatformLink(idb bun.IDB, platformID int, platformUserID string, primary bool, userID int64) (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	link := orm.PlatformLink{}

	committed := false
	tx, err := idb.BeginTx(ctx, nil)
	if err != nil {
		return errors.WithStack(err)
	}
	defer func() {
		if !committed {
			err := tx.Rollback()
			if err != nil {
				zap.L().Error("unable to rollback transaction", zap.Error(err))
			}
		}
	}()

	// Delete existing primary links if set
	if primary {
		result, err := tx.NewDelete().
			Model(&link).
			Where(`platform_id = ? AND user_id = ? AND "primary" = TRUE`, platformID, userID).
			Exec(ctx)
		if err != nil {
			return errors.WithStack(err)
		}
		rowsAffected, err := result.RowsAffected()
		if err != nil {
			return errors.WithStack(err)
		}
		if rowsAffected > 0 {
			zap.L().Info("removed rows while replacing service link",
				zap.Int64("affected rows", rowsAffected),
				zap.Int("platformID", platformID),
				zap.String("platformUserID", platformUserID),
				zap.Int64("userID", userID))
		}
	}

	link.PlatformUserID = platformUserID
	link.PlatformID = platformID
	link.UserID = userID
	link.Primary = primary
	//link.ServiceUserDisplayName = ""
	_, err = tx.NewInsert().
		Model(&link).
		On("CONFLICT (platform_id, platform_user_id) DO UPDATE").
		Exec(ctx)
	if err != nil {
		return errors.Errorf("could not persist platform link: User %s on platform %d. Error: %#v", platformUserID, platformID, err)
	}

	err = tx.Commit()
	if err != nil {
		return errors.WithStack(err)
	}
	committed = true

	zap.L().Info("stored platform link", zap.Any("link", link))
	return err
}
