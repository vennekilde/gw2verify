package api

import (
	"context"
	"time"

	"github.com/uptrace/bun"
)

const (
	None          string = "None"
	PlayForFree   string = "PlayForFree"
	GuildWars2    string = "GuildWars2"
	HeartOfThorns string = "HeartOfThorns"
	PathOfFire    string = "PathOfFire"
	EndOfDragons  string = "EndOfDragons"
)

func (m *Account) BeforeAppendModel(ctx context.Context, query bun.Query) error {
	switch query.(type) {
	case *bun.InsertQuery:
		m.DbUpdated = time.Now()
	case *bun.UpdateQuery:
		m.DbUpdated = time.Now()
	}
	return nil
}

func (m *User) BeforeAppendModel(ctx context.Context, query bun.Query) error {
	switch query.(type) {
	case *bun.InsertQuery:
		m.DbUpdated = time.Now()
	case *bun.UpdateQuery:
		m.DbUpdated = time.Now()
	}
	return nil
}
