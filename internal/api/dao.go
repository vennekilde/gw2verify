package api

import (
	"context"

	"github.com/pkg/errors"
	"github.com/uptrace/bun"
)

func (acc *Account) Persist(tx bun.IDB) (err error) {
	ctx := context.Background()
	query := tx.NewInsert().
		Model(acc).
		On(`CONFLICT ("id") DO UPDATE`)

	if acc.WvWGuildID == nil {
		// Exclude column from update
		query.ExcludeColumn("wvw_guild_id")
	}

	_, err = query.Exec(ctx)
	return errors.WithStack(err)
}
