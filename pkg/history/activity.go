package history

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"github.com/uptrace/bun"
)

type Activity struct {
	ID        int64     `bun:",pk"`
	UserID    int64     `json:"user_id"`
	Timestamp time.Time `json:"timestamp"`
	Rank      int       `json:"rank"`
	Kills     int       `json:"kills"`
}

// Equivalent checks if two activities contain the same stats and ignores the timestamp
func (a Activity) Equivalent(b Activity) bool {
	return a.UserID == b.UserID && a.Rank == b.Rank && a.Kills == b.Kills
}

// UpdateActivity updates the activity of a user
func UpdateActivity(tx bun.IDB, userID int64, rank int, kills int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Get last two activities
	var activities []Activity
	err := tx.NewSelect().
		Model(&activities).
		Where("user_id = ?", userID).
		Order("timestamp DESC").
		Limit(2).
		Scan(ctx)
	if err != nil {
		return errors.WithStack(err)
	}

	// Insert activity
	activity := Activity{
		UserID:    userID,
		Rank:      rank,
		Kills:     kills,
		Timestamp: time.Now(),
	}

	// Check if the last two activities are the same statswise
	updateLast := len(activities) == 2 && activities[0].Equivalent(activities[1]) && activities[0].Equivalent(activity)

	// If the last two activities are the same, update the last activity
	if updateLast {
		activity.ID = activities[0].ID
		_, err := tx.NewUpdate().
			Model(&activity).
			Where("id = ?", activities[0].ID).
			Exec(ctx)
		if err != nil {
			return errors.WithStack(err)
		}
	} else {
		_, err := tx.NewInsert().
			Model(&activity).
			ExcludeColumn("id"). // Exclude ID to allow for auto increment
			Exec(ctx)
		if err != nil {
			return errors.WithStack(err)
		}
	}
	return nil
}
