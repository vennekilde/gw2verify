package verify

import (
	"time"

	"github.com/vennekilde/gw2verify/v2/internal/api"
	"go.uber.org/zap"
)

type Verification struct {
	worlds *Worlds
}

func NewVerification(worlds *Worlds) *Verification {
	return &Verification{
		worlds: worlds,
	}
}

func (v *Verification) Status(worldPerspective int, user *api.User) api.Status {
	// Default, assume unknown denied
	status := api.ACCESS_DENIED_UNKNOWN

	activeBan := GetActiveBan(user.Bans)
	if activeBan != nil {
		return api.ACCESS_DENIED_BANNED
	}

	for _, acc := range user.Accounts {
		accStatus := v.AccountStatus(worldPerspective, &acc)
		if accStatus.Priority() > status.Priority() {
			status = accStatus
		}
	}

	if status.AccessDenied() {
		// Check if user has been given temporary access
		for _, assoc := range user.EphemeralAssociations {
			if assoc.Until == nil || assoc.World == nil || time.Now().After(*assoc.Until) {
				continue
			}
			ephStatus := v.AccountWorldStatus(worldPerspective, *assoc.World, true)
			if ephStatus.Priority() > status.Priority() {
				status = ephStatus
			}
		}
	}

	return status
}

// AccountStatus checks the verification api.ACCESS status for an account given a world perspective
func (v *Verification) AccountStatus(worldPerspective int, acc *api.Account) api.Status {
	if acc == nil || acc.ID == "" {
		return api.ACCESS_DENIED_ACCOUNT_NOT_LINKED
	}

	//Check if access is expired
	if acc.Expired != nil && *acc.Expired {
		return api.ACCESS_DENIED_EXPIRED
	}

	return v.AccountWorldStatus(worldPerspective, acc.World, false)
}

func (v *Verification) AccountWorldStatus(worldPerspective int, world int, ephemeral bool) api.Status {
	if world == worldPerspective {
		if ephemeral {
			return api.ACCESS_GRANTED_HOME_WORLD_TEMPORARY
		}
		return api.ACCESS_GRANTED_HOME_WORLD
	}

	//Get cached world links
	worldLinks, err := v.worlds.GetWorldLinks(worldPerspective)
	if err != nil {
		zap.L().Error("could not get linked worlds", zap.Error(err))
		return api.ACCESS_DENIED_UNKNOWN
	}
	for _, worldLink := range worldLinks {
		if worldLink == world {
			if ephemeral {
				return api.ACCESS_GRANTED_LINKED_WORLD_TEMPORARY
			}
			return api.ACCESS_GRANTED_LINKED_WORLD
		}
	}

	return api.ACCESS_DENIED_INVALID_WORLD
}
