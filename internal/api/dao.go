package api

import (
	"context"

	"github.com/pkg/errors"
	"github.com/uptrace/bun"
)

func (acc *Account) Persist(tx bun.IDB) (err error) {
	ctx := context.Background()
	_, err = tx.NewInsert().
		Model(acc).
		On(`CONFLICT ("id") DO UPDATE`).
		Exec(ctx)
	return errors.WithStack(err)
}
