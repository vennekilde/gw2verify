package verify

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"github.com/uptrace/bun"
	"github.com/vennekilde/gw2verify/v2/internal/api"
)

// GrantEphemeralWorldAssignment temporarily mark the account as being from the given world
// This will grant the user temporary access, if they are set to be a from a world that would normally grant access
func GrantEphemeralWorldAssignment(tx bun.Tx, userID int64, world int, until time.Time) (err error, userErr error) {
	ctx := context.Background()
	temporaryAccess := api.EphemeralAssociation{
		UserID: userID,
		World:  &world,
		Until:  &until,
	}

	_, err = tx.NewInsert().
		Model(&temporaryAccess).
		Exec(ctx)
	return errors.WithStack(err), nil
}
