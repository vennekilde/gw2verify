package sync

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/MrGunflame/gw2api"
	"github.com/vennekilde/gw2verify/v2/internal/api"
	"github.com/vennekilde/gw2verify/v2/internal/config"
	"github.com/vennekilde/gw2verify/v2/internal/orm"
	"github.com/vennekilde/gw2verify/v2/pkg/utils"
	"github.com/vennekilde/gw2verify/v2/pkg/verify"
)

// FreeToPlayWvWRankRestriction restricts the minimum required wvw rank before free to play users can verify
var FreeToPlayWvWRankRestriction = 0

// SetAPIKeyByUserService sets an apikey from a user of a specific service
func (s *Service) SetAPIKeyByUserService(gw2API *gw2api.Session, worldPerspective *int, platformID int, platformUserID string, primary bool, apikey string, ignoreRestrictions bool) (err error, userErr error) {
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

	user, err := verify.GetOrInsertUser(tx, platformID, platformUserID)
	if err != nil {
		return errors.WithStack(err), nil
	}

	err = verify.SetOrReplacePlatformLink(tx, platformID, platformUserID, primary, user.Id)
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
	if !ignoreRestrictions {
		// @DISABLED
		// Ensure no one is already linked with this account
		//var storedLink PlatformLink
		//if err = orm.DB().NewSelect().Model(&storedLink).Where( "platform_id = ? AND account_id = ? AND is_primary = true", platformID, acc.ID).Error; err != nil && err != sql.ErrNoRows {
		//	return err, nil
		//}
		//if storedLink.PlatformUserID != "" && storedLink.PlatformUserID != platformUserID {
		//	return fmt.Errorf("Could not set APIKey '%s' for user %s on service %d. Already linked with another user with id %s", apikey, platformUserID, platformID, storedLink.PlatformUserID), errors.New("Account already linked with another user. Contact an admin if you need to transfer access to your current discord user")
		//}

		// Additional restrictions
		err = s.processRestrictions(gw2API, worldPerspective, acc, gw2Token, platformID, platformUserID)
		if err != nil {
			return err, err
		}
	}

	var newAcc, oldAcc api.Account
	err = orm.DB().NewSelect().
		Model(&oldAcc).
		Where(`"id" = ?`, acc.ID).
		Scan(ctx)
	if err != nil && err != sql.ErrNoRows {
		return err, nil
	}

	// Notify listeners if needed of verification changes (if any)
	newAcc.FromGW2API(acc)
	if s.em.ShouldEmitAccount(&oldAcc, &newAcc) {
		err = orm.QueryGetUser(tx, user, user.Id).
			Model(user).
			Scan(ctx)
		if err != nil {
			return err, nil
		}
		s.em.Emit(user)
	}

	// Persist account info
	newAcc.UserID = user.Id
	err = newAcc.Persist(tx)
	if err != nil {
		return errors.WithStack(err), nil
	}

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

	err = tx.Commit()
	if err != nil {
		return errors.WithStack(err), nil
	}
	committed = true

	zap.L().Info("stored APIKey", zap.String("APIKey", apikey), zap.String("Account ID", acc.ID))
	return nil, userErr
}

func (s *Service) processRestrictions(gw2api *gw2api.Session, worldPerspective *int, acc gw2api.Account, token gw2api.TokenInfo, platformID int, platformUserID string) (err error) {
	if config.Config().SkipRestrictions {
		return nil
	}
	if err := s.processAPIKeyRestrictions(worldPerspective, acc, token, platformID, platformUserID); err != nil {
		return err
	}
	if err := s.processAccountRestrictions(acc.Access, acc.World); err != nil {
		return err
	}
	if err := s.processCharacterRestrictions(gw2api, acc); err != nil {
		return err
	}
	return nil
}

func (s *Service) processAPIKeyRestrictions(worldPerspective *int, acc gw2api.Account, token gw2api.TokenInfo, platformID int, platformUserID string) (err error) {
	//Check if api key is named correctly
	apiKeyCode := verify.GetAPIKeyCode(platformID, platformUserID)
	if !strings.Contains(strings.ToUpper(token.Name), apiKeyCode) {
		return fmt.Errorf("APIKey name incorrect. You need to name your api key \"%s\" instead of \"%s\"", verify.GetAPIKeyName(worldPerspective, platformID, platformUserID), token.Name)
	}

	//Check if api key has the correct permissions
	requiredPermissions := []string{"progression", "characters", "wvw"}
	missingPermissions := make([]string, 0, len(requiredPermissions))
	for _, perm := range requiredPermissions {
		if !Contains(token.Permissions, perm) {
			missingPermissions = append(missingPermissions, perm)
		}
	}

	if len(missingPermissions) > 0 {
		return fmt.Errorf("missing apikey permissions: [%s]. Please create a new api key with the following permissions enabled: [%s]", strings.Join(missingPermissions, ", "), strings.Join(requiredPermissions, ", "))
	}

	return err
}

func (s *Service) processAccountRestrictions(access []string, rank int) (err error) {
	freeToPlay := IsFreeToPlay(access)

	//FreeToPlay restrictions
	if freeToPlay {
		//Check if FreeToPlay player meets WvW rank requirements
		if rank < FreeToPlayWvWRankRestriction {
			return fmt.Errorf("You need to have WvW rank %d to verify yourself. This is required for all FreeToPlay accounts\nCurrent WvW rank: %d", FreeToPlayWvWRankRestriction, rank)
		}
	}
	return err
}

func (s *Service) processCharacterRestrictions(gw2api *gw2api.Session, acc gw2api.Account) (err error) {
	freeToPlay := IsFreeToPlay(acc.Access)

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
