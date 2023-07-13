package verify

import (
	"context"
	"database/sql"
	"strconv"
	"time"

	"github.com/uptrace/bun"
	"github.com/vennekilde/gw2verify/internal/api"
	"github.com/vennekilde/gw2verify/internal/config"
	"github.com/vennekilde/gw2verify/internal/orm"
	"go.uber.org/zap"

	"gitlab.com/MrGunflame/gw2api"
)

const (
	None          string = "None"
	PlayForFree   string = "PlayForFree"
	GuildWars2    string = "GuildWars2"
	HeartOfThorns string = "HeartOfThorns"
	PathOfFire    string = "PathOfFire"
)

type User struct {
	orm.Model     `bun:",extend"`
	bun.BaseModel `bun:"table:users,alias:users"`

	ID int
}

func Status(worldPerspective int, serviceID int, serviceUserID string) (status *api.VerificationStatusOverview, err error) {
	return StatusWithAccounts(worldPerspective, serviceID, serviceUserID, nil)
}

func StatusWithAccounts(worldPerspective int, serviceID int, serviceUserID string, accounts []orm.Account) (status *api.VerificationStatusOverview, err error) {
	// Default, assume unknown denied
	status = &api.VerificationStatusOverview{}
	status.Status = api.ACCESS_DENIED_UNKNOWN
	if serviceUserID == "" {
		return status, nil
	}

	verifyStatus := api.VerificationStatus{}

	//Check if user is linked to a gw2 account
	link, err := orm.GetServiceLink(serviceID, serviceUserID)
	if err != nil {
		if err == sql.ErrNoRows {
			return status.WithStatusIfHigher(verifyStatus.WithStatus(api.ACCESS_DENIED_ACCOUNT_NOT_LINKED)), nil
		}
		return status.WithStatusIfHigher(verifyStatus.WithStatus(api.ACCESS_DENIED_UNKNOWN)), err
	}
	if link.UserID == 0 {
		return status.WithStatusIfHigher(verifyStatus.WithStatus(api.ACCESS_DENIED_ACCOUNT_NOT_LINKED)), nil
	}

	status.ServiceLinks = &[]api.ServiceLink{
		link.ServiceLink,
	}

	//Check verification status of linked accounts
	if len(accounts) == 0 {
		accounts, err = orm.GetUserAccounts(link.UserID)
		if err != nil {
			if err == sql.ErrNoRows {
				return status.WithStatusIfHigher(verifyStatus.WithStatus(api.ACCESS_DENIED_ACCOUNT_NOT_LINKED)), nil
			}
			return status.WithStatusIfHigher(verifyStatus.WithStatus(api.ACCESS_DENIED_UNKNOWN)), err
		}
	}

	statuses := make([]api.VerificationStatus, len(accounts))
	for i := range accounts {
		accStatus := AccountStatus(&accounts[i], worldPerspective)

		// Add status to overview
		statuses[i] = *accStatus

		// Determine if status should be the primary one
		status.WithStatusIfHigher(accStatus)
	}
	status.Statuses = &statuses

	// Check if user has been given temporary access
	_ = checkTemporaryAccesses(worldPerspective, serviceID, serviceUserID, status)

	return status, nil
}

func checkTemporaryAccesses(worldPerspective int, serviceID int, serviceUserID string, status *api.VerificationStatusOverview) (err error) {
	ctx := context.Background()
	// Check for temporary access
	tempAccesses := []TemporaryAccess{}
	err = orm.DB().NewSelect().
		Model(&tempAccesses).
		Where("service_id = ? AND service_user_id = ?", serviceID, serviceUserID).
		Scan(ctx)
	if err != nil {
		// Cannot be identified as a temporary user
		if err == sql.ErrNoRows {
			return nil
		}
		return err
	}

OUTER:
	for i := range tempAccesses {
		tempAccess := &tempAccesses[i]
		// Check if user already has access with this world
		if status.Statuses != nil {
			for _, s := range *status.Statuses {
				if s.World != nil && tempAccess.World == *s.World {
					// Do not check temp status, if user already has access on this world
					if s.Status.AccessGranted() {
						continue OUTER
					}
				}
			}
		}

		err = checkTemporaryAccess(worldPerspective, tempAccess, status)
		if err != nil {
			zap.L().Error("error while checking temporary access data", zap.Any("data", tempAccess), zap.Error(err))
		}
	}

	return nil
}

func checkTemporaryAccess(worldPerspective int, tempAccess *TemporaryAccess, status *api.VerificationStatusOverview) (err error) {
	// determine expiration date time
	expires := tempAccess.DbUpdated.Add(time.Duration(config.Config().TemporaryAccessExpirationTime) * time.Second)

	tempStatus := api.VerificationStatus{
		// default to expired
		Status:  api.ACCESS_DENIED_EXPIRED,
		World:   &tempAccess.World,
		Expires: &expires,
		/*ServiceLink: &api.ServiceLink{
			ServiceID:     tempAccess.ServiceID,
			ServiceUserID: tempAccess.ServiceUserID,
		},*/
	}

	defer func() {
		status.WithStatusIfHigher(&tempStatus)
		*status.Statuses = append(*status.Statuses, tempStatus)
	}()

	if time.Now().After(expires) {
		// temporary access has expired
		tempStatus.Status = api.ACCESS_DENIED_EXPIRED
		return nil
	}
	if tempAccess.World == worldPerspective {
		// identified as given temporary home world access
		tempStatus.Status = api.ACCESS_GRANTED_HOME_WORLD_TEMPORARY
		return nil
	}

	// Check if user has linked world temp access
	// Get cached world links
	worldLinks, err := GetWorldLinks(worldPerspective)
	if err != nil {
		// Cannot get world links
		if err == sql.ErrNoRows {
			return nil
		}
		return err
	}

	// check if temporary access is given to a linked world
	for _, world := range worldLinks {
		if world == tempAccess.World {
			tempStatus.Status = api.ACCESS_GRANTED_LINKED_WORLD_TEMPORARY
			return nil
		}
	}

	// not home world or linked
	tempStatus.Status = api.ACCESS_DENIED_INVALID_WORLD

	return nil
}

// AccountStatus checks the verification api.ACCESS status for an account given a world perspective
func AccountStatus(acc *orm.Account, worldPerspective int) (status *api.VerificationStatus) {
	ctx := context.Background()
	status = &api.VerificationStatus{
		Account: &api.Account{},
	}
	if acc == nil || acc.ID == "" {
		return status.WithStatus(api.ACCESS_DENIED_ACCOUNT_NOT_LINKED)
	}

	status.World = &acc.World
	// Convert gw2api account into api account struct
	status.Account.FromGW2API(acc.Account)

	//Ban logic
	banStatus := GetBan(acc)
	if banStatus != nil {
		status.Ban = &banStatus.BanData
		return status.WithStatus(api.ACCESS_DENIED_BANNED)
	}

	//Check if access is expired
	if int(time.Since(acc.DbUpdated).Seconds()) >= config.Config().ExpirationTime {
		//Check if last time token is expired
		tokens := []gw2api.TokenInfo{}
		//For some reason, the Go pg library complains with "pq: got 2 parameters but the statement requires 1"
		//when the param is inside the ' ' closure, so we insert it as part of the string instead
		//It's an int anyway and we trust the source, so it should be fine
		orm.DB().NewSelect().
			Model(&tokens).
			Where("account_id = ? AND last_success >= db_updated - interval '"+strconv.Itoa(config.Config().ExpirationTime)+" seconds'", acc.ID).
			Scan(ctx)
		if len(tokens) <= 0 {
			//No valid api keys found, therefore must be expired
			return status.WithStatus(api.ACCESS_DENIED_EXPIRED)
		}
	}

	if err := processAccountRestrictions(worldPerspective, acc.Account); err != nil {
		desc := err.Error()
		status.Description = &desc
		return status.WithStatus(api.ACCESS_DENIED_REQUIREMENT_NOT_MET)
	}

	if acc.World == worldPerspective {
		return status.WithStatus(api.ACCESS_GRANTED_HOME_WORLD)
	}

	//Get cached world links
	worldLinks, err := GetWorldLinks(worldPerspective)
	if err != nil {
		zap.L().Error("could not get linked worlds", zap.Error(err))
		return status.WithStatus(api.ACCESS_DENIED_UNKNOWN)
	}
	for _, world := range worldLinks {
		if world == acc.World {
			return status.WithStatus(api.ACCESS_GRANTED_LINKED_WORLD)
		}
	}

	return status.WithStatus(api.ACCESS_DENIED_INVALID_WORLD)
}
