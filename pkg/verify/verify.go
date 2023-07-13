package verify

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/uptrace/bun"
	"go.uber.org/zap"

	"github.com/vennekilde/gw2verify/internal/api"
	"github.com/vennekilde/gw2verify/internal/config"
	"github.com/vennekilde/gw2verify/internal/orm"
	"github.com/vennekilde/gw2verify/pkg/utils"
	"gitlab.com/MrGunflame/gw2api"
)

// FreeToPlayWvWRankRestriction restricts the minimum required wvw rank before free to play users can verify
var FreeToPlayWvWRankRestriction = 0

type VerifictionStatusListener struct {
	WorldPerspective int
	ServiceID        int
	Listener         chan *api.VerificationStatus
}

var ServicePollListeners map[int]VerifictionStatusListener = make(map[int]VerifictionStatusListener)

// SetAPIKeyByUserService sets an apikey from a user of a specific service
func SetAPIKeyByUserService(gw2API *gw2api.Session, worldPerspective int, serviceID int, serviceUserID string, primary bool, apikey string, ignoreRestrictions bool) (err error, userErr error) {
	ctx := context.Background()

	tx, err := orm.DB().Begin()
	if err != nil {
		return errors.WithStack(err), nil
	}
	committed := false
	defer func() {
		if !committed {
			tx.Rollback()
		}
	}()

	user, err := GetOrInsertUser(tx, serviceID, serviceUserID)
	if err != nil {
		return errors.WithStack(err), nil
	}

	err = SetOrReplaceServiceLink(tx, serviceID, serviceUserID, primary, user.ID)
	if err != nil {
		return errors.WithStack(err), nil
	}

	//Strip spaces
	apikey = utils.StripWhitespace(apikey)

	// Prepare api client
	gw2API.WithAccessToken(apikey)

	// Fetch token and account from gw2 api
	gw2Token, err := gw2API.Tokeninfo()
	if err != nil {
		return errors.WithStack(err), err
	}
	acc, err := gw2API.Account()
	if err != nil {
		return errors.WithStack(err), err
	}

	// Check if apikey & account pass restrictions
	if ignoreRestrictions == false {
		// @DISABLED
		// Ensure no one is already linked with this account
		//var storedLink ServiceLink
		//if err = orm.DB().NewSelect().Model(&storedLink).Where( "service_id = ? AND account_id = ? AND is_primary = true", serviceID, acc.ID).Error; err != nil && err != sql.ErrNoRows {
		//	return err, nil
		//}
		//if storedLink.ServiceUserID != "" && storedLink.ServiceUserID != serviceUserID {
		//	return fmt.Errorf("Could not set APIKey '%s' for user %s on service %d. Already linked with another user with id %s", apikey, serviceUserID, serviceID, storedLink.ServiceUserID), errors.New("Account already linked with another user. Contact an admin if you need to transfer access to your current discord user")
		//}

		// Additional restrictions
		err = processRestrictions(gw2API, worldPerspective, acc, gw2Token, serviceID, serviceUserID)
		if err != nil {
			return err, err
		}
	}

	var storedAcc orm.Account
	err = orm.DB().NewSelect().
		Model(&storedAcc).
		Where(`"id" = ?`, acc.ID).
		Scan(ctx)
	if err != nil && err != sql.ErrNoRows {
		return err, nil
	}

	// Check if a notification should be published
	CheckForVerificationUpdate(storedAcc, acc)

	// Persist token info
	token := orm.TokenInfo{
		TokenInfo:   gw2Token,
		AccountID:   acc.ID,
		APIKey:      apikey,
		LastSuccess: time.Now(),
	}

	err = token.Persist(tx)
	if err != nil {
		return err, nil
	}

	// Persist account info
	ormAcc := orm.Account{Account: acc, UserID: user.ID}
	err = ormAcc.Persist(tx)
	if err != nil {
		return errors.WithStack(err), nil
	}

	err = tx.Commit()
	if err != nil {
		return errors.WithStack(err), nil
	}
	committed = true

	zap.L().Info("stored APIKey", zap.String("APIKey", apikey), zap.String("Account ID", acc.ID))
	return nil, userErr
}

func processRestrictions(gw2api *gw2api.Session, worldPerspective int, acc gw2api.Account, token gw2api.TokenInfo, serviceID int, serviceUserID string) (err error) {
	if config.Config().SkipRestrictions {
		return nil
	}
	if err := processAPIKeyRestrictions(worldPerspective, acc, token, serviceID, serviceUserID); err != nil {
		return err
	}
	if err := processAccountRestrictions(worldPerspective, acc); err != nil {
		return err
	}
	if err := processCharacterRestrictions(gw2api, acc); err != nil {
		return err
	}
	return nil
}

func processAPIKeyRestrictions(worldPerspective int, acc gw2api.Account, token gw2api.TokenInfo, serviceID int, serviceUserID string) (err error) {
	//Check if api key is named correctly
	apiKeyCode := GetAPIKeyCode(serviceID, serviceUserID)
	if strings.Contains(strings.ToUpper(token.Name), apiKeyCode) == false {
		return fmt.Errorf("APIKey name incorrect. You need to name your api key \"%s\" instead of \"%s\"", GetAPIKeyName(worldPerspective, serviceID, serviceUserID), token.Name)
	}

	//freeToPlay := IsFreeToPlay(acc)

	//FreeToPlay restrictions
	//if freeToPlay {
	//Ensure progression permission is present
	hasProgression := Contains(token.Permissions, "progression")

	//Ensure characters permission is present
	hasCharacters := Contains(token.Permissions, "characters")

	if !hasProgression || !hasCharacters {
		return errors.New("missing apikey permission \"characters\" and/or \"progression\"")
	}
	//}

	return err
}

func processAccountRestrictions(worldPerspective int, acc gw2api.Account) (err error) {
	freeToPlay := IsFreeToPlay(acc)

	//FreeToPlay restrictions
	if freeToPlay {
		//Check if FreeToPlay player meets WvW rank requirements
		if acc.WvWRank < FreeToPlayWvWRankRestriction {
			return fmt.Errorf("You need to have WvW rank %d to verify yourself. This is required for all FreeToPlay accounts\nCurrent WvW rank: %d", FreeToPlayWvWRankRestriction, acc.WvWRank)
		}
	}
	return err
}

func processCharacterRestrictions(gw2api *gw2api.Session, acc gw2api.Account) (err error) {
	freeToPlay := IsFreeToPlay(acc)

	//FreeToPlay restrictions
	if freeToPlay {
		//Fetch character names
		charNames, err := gw2api.Characters()
		if err != nil {
			return err
		}

		highestLevel := 0
		for _, name := range charNames {
			//Fetch characters
			char, err := gw2api.CharacterCore(name)
			if err != nil {
				return err
			}

			if char.Level > highestLevel {
				highestLevel = char.Level
			}
		}

		//FreeToPlay level 80 retrictions
		if highestLevel < 80 {
			return fmt.Errorf("You need to be level 80 to verify yourself. This is required for all FreeToPlay accounts\nHighest character level found: %d", highestLevel)
		}

	}
	return err
}

// CheckForVerificationUpdate checks if the verification status for the user has changed since last synchronization
// If it has, it will send out a notification to all registered listeners
func CheckForVerificationUpdate(storedAcc orm.Account, acc gw2api.Account) (err error) {
	// Clone
	newAcc := storedAcc
	newAcc.Account = acc
	if storedAcc.World != acc.World || int(time.Since(storedAcc.DbUpdated).Seconds()) >= config.Config().ExpirationTime {
		err = OnVerificationUpdate(newAcc)
	}
	return err
}

// OnVerificationUpdate sends out a verification notification to all registered listeners
func OnVerificationUpdate(acc orm.Account) (err error) {
	links, err := orm.GetUserServiceLinks(acc.UserID)
	if err != nil {
		return err
	}

	for i, link := range links {
		serviceListener := ServicePollListeners[link.ServiceID]
		if serviceListener.Listener != nil {
			acc.DbUpdated = time.Now().UTC()
			status := AccountStatus(&acc, serviceListener.WorldPerspective)
			status.ServiceLink = &links[i].ServiceLink
			serviceListener.Listener <- status
		}
	}

	return err
}

func GetOrInsertUser(tx bun.Tx, serviceID int, serviceUserID string) (*User, error) {
	user := User{}
	ctx := context.Background()

	err := tx.NewSelect().
		Model(&user).
		Join("INNER JOIN service_links ON users.id = service_links.user_id").
		Where("service_id = ? AND service_user_id = ?", serviceID, serviceUserID).
		Scan(ctx)
	if err != nil && err != sql.ErrNoRows {
		return nil, errors.WithStack(err)
	}

	if err == sql.ErrNoRows {
		_, err = tx.NewInsert().
			Model(&user).
			Returning("*").
			Exec(ctx)
		if err != nil {
			return nil, errors.WithStack(err)
		}
	}

	return &user, nil
}

// SetOrReplaceServiceLink creates or replaces a service link between a service user and an account
func SetOrReplaceServiceLink(tx bun.Tx, serviceID int, serviceUserID string, primary bool, userID int) (err error) {
	ctx := context.Background()
	link := orm.ServiceLink{}

	// Delete existing primary links if set
	if primary {
		result, err := tx.NewDelete().
			Model(&link).
			Where(`service_id = ? AND user_id = ? AND "primary" = TRUE`, serviceID, userID).
			Exec(ctx)
		if err != nil {
			return err
		}
		rowsAffected, err := result.RowsAffected()
		if err != nil {
			return err
		}
		if rowsAffected > 0 {
			zap.L().Info("removed rows while replacing service link",
				zap.Int64("affected rows", rowsAffected),
				zap.Int("serviceID", serviceID),
				zap.String("serviceUserID", serviceUserID),
				zap.Int("userID", userID))
		}
	}

	link.ServiceUserID = serviceUserID
	link.ServiceID = serviceID
	link.UserID = userID
	link.Primary = primary
	//link.ServiceUserDisplayName = ""
	_, err = tx.NewInsert().
		Model(&link).
		On("CONFLICT (service_id, service_user_id) DO UPDATE").
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("could not persist service link: User %s on service %d. Error: %#v", serviceUserID, serviceID, err)
	}

	zap.L().Info("stored service link", zap.Any("link", link))
	return err
}
