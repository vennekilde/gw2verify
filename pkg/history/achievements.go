package history

import (
	"context"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/uptrace/bun"
)

type Achievement struct {
	ID          int64     `bun:",pk"`
	AccountID   string    `json:"account_id"`
	Timestamp   time.Time `json:"timestamp"`
	Achievement int       `json:"achievement"`
	Value       int       `json:"value"`
}

// Equivalent checks if two activities contain the same stats and ignores the timestamp
func (a Achievement) Equivalent(b Achievement) bool {
	return strings.EqualFold(a.AccountID, b.AccountID) && a.Achievement == b.Achievement && a.Value == b.Value
}

// UpdateAchievement updates the achievement of a user
func UpdateAchievement(tx bun.IDB, accountID string, achievementID int, value int) error {
	ctx := context.Background()

	// Get last two achievements
	var achievements []Achievement
	err := tx.NewSelect().
		Model(&achievements).
		Where("account_id = ? AND achievement = ?", accountID, achievementID).
		Order("timestamp DESC").
		Limit(2).
		Scan(ctx)
	if err != nil {
		return errors.WithStack(err)
	}

	// Insert activity
	achievement := Achievement{
		AccountID:   accountID,
		Timestamp:   time.Now(),
		Achievement: achievementID,
		Value:       value,
	}

	// Check if the last two activities are the same statswise
	updateLast := len(achievements) == 2 && achievements[0].Equivalent(achievements[1]) && achievements[0].Equivalent(achievement)

	// If the last two activities are the same, update the last activity
	if updateLast {
		achievement.ID = achievements[0].ID
		_, err := tx.NewUpdate().
			Model(&achievement).
			Where("id = ?", achievement.ID).
			Exec(ctx)
		if err != nil {
			return errors.WithStack(err)
		}
	} else {
		_, err := tx.NewInsert().
			Model(&achievement).
			ExcludeColumn("id"). // Exclude ID to allow for auto increment
			Exec(ctx)
		if err != nil {
			return errors.WithStack(err)
		}
	}
	return nil
}
