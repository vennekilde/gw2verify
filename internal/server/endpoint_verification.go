package server

import (
	"context"
	"net/http"
	"time"

	"github.com/MrGunflame/gw2api"
	"github.com/gin-gonic/gin"
	"github.com/vennekilde/gw2verify/v2/internal/api"
	"github.com/vennekilde/gw2verify/v2/internal/config"
	"github.com/vennekilde/gw2verify/v2/internal/orm"
	"github.com/vennekilde/gw2verify/v2/pkg/history"
	"github.com/vennekilde/gw2verify/v2/pkg/sync"
	"github.com/vennekilde/gw2verify/v2/pkg/verify"
)

type VerificationEndpoint struct {
	verification *verify.Verification
	worlds       *verify.Worlds
	statistics   *history.Statistics
	eventEmitter *verify.EventEmitter
	syncher      *sync.Service
	banService   *verify.BanService
}

func NewVerificationEndpoint(verification *verify.Verification, worlds *verify.Worlds, statistics *history.Statistics, eventEmitter *verify.EventEmitter, syncher *sync.Service, banService *verify.BanService) *VerificationEndpoint {
	return &VerificationEndpoint{
		verification: verification,
		worlds:       worlds,
		statistics:   statistics,
		eventEmitter: eventEmitter,
		syncher:      syncher,
		banService:   banService,
	}
}

// (GET /v1/verification/platform/{platform_id}/users/updates)
func (e *VerificationEndpoint) GetVerificationPlatformUserUpdates(c *gin.Context, platformId api.PlatformId, params api.GetVerificationPlatformUserUpdatesParams) {
	ticker := time.NewTicker(120 * time.Second)
	defer func() { ticker.Stop() }()

	ch := e.eventEmitter.GetStatusListener(c.GetString("service_id"), &platformId, params.World)

	select {
	case event := <-ch:
		c.JSON(http.StatusOK, event)
	case <-ticker.C:
		c.Status(http.StatusRequestTimeout)
	}
}

// (GET /v1/verification/platform/{platform_id}/users/{platform_user_id})
func (e *VerificationEndpoint) GetVerificationPlatformUserStatus(c *gin.Context, platformId api.PlatformId, platformUserId api.PlatformUserId, params api.GetVerificationPlatformUserStatusParams) {
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(10*time.Second))
	defer cancel()

	var user api.User
	err := orm.QueryGetPlatformUser(orm.DB(), &user, platformId, platformUserId).
		//ColumnExpr("bans.*, ephemeral_associations.world").
		Scan(ctx)
	if err != nil {
		ThrowReqError(c, err.Error(), nil, http.StatusInternalServerError)
		return
	}

	status := api.VerificationStatus{
		Ban:    verify.GetActiveBan(user.Bans),
		Status: e.verification.Status(params.World, &user),
	}
	c.JSON(http.StatusOK, &status)
}

// (POST /v1/verification/platform/{platform_id}/users/{platform_user_id}/refresh)
func (e *VerificationEndpoint) PostVerificationPlatformUserRefresh(c *gin.Context, platformId api.PlatformId, platformUserId api.PlatformUserId, params api.PostVerificationPlatformUserRefreshParams) {
	tx, err := orm.DB().BeginTx(c, nil)
	if err != nil {
		ThrowReqError(c, err.Error(), nil, http.StatusInternalServerError)
		return
	}
	committed := false
	defer func() {
		if !committed {
			tx.Rollback()
		}
	}()

	link, err := orm.GetPlatformLink(platformId, platformUserId)
	if err != nil {
		ThrowReqError(c, err.Error(), err, http.StatusInternalServerError)
		return
	}
	if link.UserID == 0 {
		ThrowReqError(c, err.Error(), err, http.StatusNotFound)
	}

	gw2API := gw2api.New()
	err = e.syncher.SynchronizeUser(tx, gw2API, link.UserID)
	if err != nil {
		ThrowReqError(c, err.Error(), err, http.StatusInternalServerError)
		return
	}

	err = tx.Commit()
	if err != nil {
		ThrowReqError(c, err.Error(), nil, http.StatusInternalServerError)
		return
	}
	committed = true

	e.GetVerificationPlatformUserStatus(c, platformId, platformUserId, api.GetVerificationPlatformUserStatusParams{
		World: params.World,
	})
}

// (PUT /v1/verification/platform/{platform_id}/users/{platform_user_id}/temporary)
func (e *VerificationEndpoint) PutVerificationPlatformUserTemporary(c *gin.Context, platformId api.PlatformId, platformUserId api.PlatformUserId, params api.PutVerificationPlatformUserTemporaryParams) {
	var reqBody api.EphemeralAssociation

	tx, err := orm.DB().BeginTx(c, nil)
	if err != nil {
		ThrowReqError(c, err.Error(), nil, http.StatusInternalServerError)
		return
	}
	committed := false
	defer func() {
		if !committed {
			tx.Rollback()
		}
	}()

	// decode request
	err = c.Bind(&reqBody)
	if err != nil {
		ThrowReqError(c, err.Error(), err, http.StatusBadRequest)
		return
	}

	var world int
	if reqBody.World != nil && *reqBody.World > 0 {
		world = *reqBody.World
	} else if reqBody.AccessType != nil {
		if *reqBody.AccessType == api.HOME_WORLD {
			// Grant Home World temporary access
			world = params.World
		} else if *reqBody.AccessType == api.LINKED_WORLD {
			// Grant Linked World temporary access
			worldLinks, err := e.worlds.GetWorldLinks(params.World)
			if err != nil {
				ThrowReqError(c, "unable to get world links", nil, http.StatusInternalServerError)
				return
			}
			if len(worldLinks) > 0 {
				world = worldLinks[0]
			} else {
				// Not linked with another world, so cannot temporary grant linked world access
				// @TODO Consider just setting the user to home world temporary in this case
				ThrowReqError(c, "Currently not linked with any other servers", nil, http.StatusBadRequest)
				return
			}
		} else {
			ThrowReqError(c, "Invalid AccessType", nil, http.StatusBadRequest)
			return
		}
	} else {
		ThrowReqError(c, "Missing world or access_type", nil, http.StatusBadRequest)
		return
	}

	var until time.Time
	if reqBody.Until != nil {
		until = *reqBody.Until
	} else {
		until = time.Now().Add(time.Duration(config.Config().TemporaryAccessExpirationTime) * time.Second)
	}

	user, err := verify.GetOrInsertUser(tx, platformId, platformUserId)
	if err != nil {
		ThrowReqError(c, err.Error(), err, http.StatusInternalServerError)
		return
	}
	err = verify.SetOrReplacePlatformLink(tx, platformId, platformUserId, true, user.Id)
	if err != nil {
		ThrowReqError(c, err.Error(), err, http.StatusInternalServerError)
		return
	}

	err, userErr := verify.GrantEphemeralWorldAssignment(tx, user.Id, world, until)
	if err != nil {
		ThrowReqError(c, err.Error(), userErr, http.StatusInternalServerError)
		return
	}

	err = tx.Commit()
	if err != nil {
		ThrowReqError(c, err.Error(), nil, http.StatusInternalServerError)
		return
	}
	committed = true

	respBody := config.Config().TemporaryAccessExpirationTime
	c.JSON(200, &respBody)
}
