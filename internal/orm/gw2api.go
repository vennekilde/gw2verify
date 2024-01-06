package orm

import (
	"context"
	"strconv"
	"time"

	"github.com/MrGunflame/gw2api"
	"github.com/pkg/errors"
	"github.com/uptrace/bun"
	"github.com/vennekilde/gw2verify/v2/internal/api"
)

type Model struct {
	DbCreated time.Time `bun:",nullzero,notnull,default:current_timestamp,scanonly"`
	DbUpdated time.Time `bun:",nullzero,notnull,default:current_timestamp"`
}

func (m *Model) BeforeAppendModel(ctx context.Context, query bun.Query) error {
	switch query.(type) {
	case *bun.InsertQuery:
		m.DbUpdated = time.Now()
	case *bun.UpdateQuery:
		m.DbUpdated = time.Now()
	}
	return nil
}

type TokenInfo struct {
	Model            `bun:",extend"`
	gw2api.TokenInfo `bun:",extend"`

	LastSuccess time.Time
	APIKey      string `bun:"api_key"`
	AccountID   string
}

func (token *TokenInfo) Persist(tx bun.IDB) (err error) {
	ctx := context.Background()
	_, err = tx.NewInsert().
		Model(token).
		On(`CONFLICT ("id") DO UPDATE`).
		Exec(ctx)
	return errors.WithStack(err)
}

func (token *TokenInfo) UpdateLastAttemptedUpdate() (err error) {
	ctx := context.Background()
	_, err = DB().NewUpdate().
		Model(token).
		Where(`"id" = ?`, token.ID).
		Set("db_updated = ?", time.Now()).
		Exec(ctx)
	return errors.WithStack(err)
}

func (token *TokenInfo) UpdateLastSuccessfulUpdate() (err error) {
	ctx := context.Background()
	_, err = DB().NewUpdate().Model(token).
		Set("db_updated = ?", time.Now().UTC()).
		Set("last_success = ?", time.Now().UTC()).
		Where(`"id" = ?`, token.ID).
		Exec(ctx)
	return errors.WithStack(err)
}

func FindLastUpdatedAPIKey(ignoreOlderThan int) (token TokenInfo, err error) {
	ctx := context.Background()
	err = DB().NewSelect().
		Model(&token).
		Order("db_updated").
		Where("last_success >= db_updated - interval '" + strconv.Itoa(ignoreOlderThan) + " seconds' OR last_success IS NULL").
		Limit(1).
		Scan(ctx)
	return token, errors.WithStack(err)
}

func FindUserAPIKeys(userID int64, ignoreOlderThan int) (tokens []TokenInfo, err error) {
	ctx := context.Background()
	err = DB().NewSelect().
		Model(&tokens).
		Order("db_updated").
		Join("INNER JOIN accounts ON accounts.id = account_id AND accounts.user_id = ?", userID).
		Where("last_success >= token_info.db_updated - interval '" + strconv.Itoa(ignoreOlderThan) + " seconds' OR last_success IS NULL").
		Scan(ctx)
	return tokens, errors.WithStack(err)
}

type Account struct {
	Model          `bun:",extend"`
	gw2api.Account `bun:",extend"`
	UserID         int64
}

func (acc *Account) Persist(tx bun.IDB) (err error) {
	ctx := context.Background()
	_, err = tx.NewInsert().
		Model(acc).
		On(`CONFLICT ("id") DO UPDATE`).
		Exec(ctx)
	return errors.WithStack(err)
}

func GetUserAccounts(userID int64) (accounts []Account, err error) {
	ctx := context.Background()
	err = DB().NewSelect().
		Model(&accounts).
		Where("user_id = ?", userID).
		Scan(ctx)
	return accounts, errors.WithStack(err)
}

type PlatformLink struct {
	Model            `bun:",extend"`
	api.PlatformLink `bun:",extend"`
}

func GetPlatformLink(platformID int, platformUserId string) (link PlatformLink, err error) {
	ctx := context.Background()
	err = DB().NewSelect().
		Model(&link).
		Where("platform_id = ? AND platform_user_id = ?", platformID, platformUserId).
		Scan(ctx)
	return link, errors.WithStack(err)
}

func GetUserPlatformLinks(userID int64) (links []PlatformLink, err error) {
	ctx := context.Background()
	err = DB().NewSelect().
		Model(&links).
		Where("user_id = ?", userID).
		Scan(ctx)
	return links, errors.WithStack(err)
}

// Ban contains information on length and why an account was banned
type Ban struct {
	Model         `bun:",extend"`
	api.Ban       `bun:",extend"`
	bun.BaseModel `bun:"table:bans,alias:bans"`
}
