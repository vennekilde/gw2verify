package verify

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"github.com/uptrace/bun"
	"github.com/vennekilde/gw2verify/internal/orm"
)

// TemporaryAccess represents a user that has been granted temporary access from a given world
type TemporaryAccess struct {
	orm.Model     `bun:",extend"`
	bun.BaseModel `bun:"table:temporary_accesses,alias:temporary_access"`

	ServiceID     int
	ServiceUserID string
	World         int
}

// GrantTemporaryWorldAssignment temporarily mark the account as being from the given world
// This will grant the user temporary access, if they are set to be a from a world that would normally grant access
func GrantTemporaryWorldAssignment(serviceID int, serviceUserID string, world int) (err error, userErr error) {
	ctx := context.Background()
	temporaryAccess := TemporaryAccess{
		ServiceUserID: serviceUserID,
		ServiceID:     serviceID,
		World:         world,
	}
	temporaryAccess.DbUpdated = time.Now().UTC()

	_, err = orm.DB().NewInsert().
		Model(&temporaryAccess).
		On("CONFLICT (service_id, service_user_id, world) DO UPDATE").
		Exec(ctx)
	return errors.WithStack(err), nil
}
