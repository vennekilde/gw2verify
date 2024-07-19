package server

import (
	"context"
	"database/sql"
	"net/http"
	"time"

	"github.com/MrGunflame/gw2api"
	"github.com/gin-gonic/gin"
	"github.com/vennekilde/gw2verify/v2/internal/api"
	"github.com/vennekilde/gw2verify/v2/internal/orm"
	"github.com/vennekilde/gw2verify/v2/pkg/verify"
)

// (GET /v1/platform/{platform_id}/users/updates)
func (e *Endpoints) GetPlatformUserUpdates(c *gin.Context, platformId api.PlatformId, params api.GetPlatformUserUpdatesParams) {
	ticker := time.NewTicker(120 * time.Second)
	defer func() { ticker.Stop() }()

	ch := e.eventEmitter.GetUserListener(c.GetString("service_id"), &platformId)

	select {
	case event := <-ch:
		c.JSON(http.StatusOK, event)
	case <-ticker.C:
		c.Status(http.StatusRequestTimeout)
	}
}

// (GET /v1/platform/{platform_id}/users/{platform_user_id})
func (e *Endpoints) GetPlatformUser(c *gin.Context, platformId api.PlatformId, platformUserId api.PlatformUserId, params api.GetPlatformUserParams) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var user api.User
	err := orm.QueryGetPlatformUser(orm.DB(), &user, platformId, platformUserId).
		Scan(ctx)
	if err != nil {
		if err == sql.ErrNoRows {
			c.Status(404)
			return
		}
		ThrowReqError(c, err.Error(), nil, http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, &user)
}

// (PUT /v1/platform/{platform_id}/users/{platform_user_id}/apikey)
func (e *Endpoints) PutPlatformUserAPIKey(c *gin.Context, platformId api.PlatformId, platformUserId api.PlatformUserId, params api.PutPlatformUserAPIKeyParams) {
	var reqBody api.APIKeyData

	// decode request
	err := c.Bind(&reqBody)
	if err != nil {
		ThrowReqError(c, err.Error(), err, http.StatusBadRequest)
		return
	}

	if reqBody.Apikey == "" {
		ThrowReqError(c, "apikey is missing", nil, http.StatusBadRequest)
		return
	}

	skipRequirements := false
	if params.SkipRequirements != nil {
		skipRequirements = *params.SkipRequirements
	}

	gw2a := gw2api.New()
	err, userErr := e.syncher.SetAPIKeyByUserService(gw2a, params.World, platformId, platformUserId, reqBody.Primary, reqBody.Apikey, skipRequirements)
	if err != nil {
		ThrowReqError(c, err.Error(), userErr, http.StatusInternalServerError)
		return
	}
	c.Status(http.StatusCreated)
}

// (GET /v1/platform/{platform_id}/users/{platform_user_id}/apikey/name)
func (e *Endpoints) GetPlatformUserAPIKeyName(c *gin.Context, platformId api.PlatformId, platformUserId api.PlatformUserId, params api.GetPlatformUserAPIKeyNameParams) {
	apikeyName := verify.GetAPIKeyName(params.World, platformId, platformUserId)
	RespWithSuccess(c, api.APIKeyName{Name: apikeyName})
}

// (PUT /v1/platform/{platform_id}/users/{platform_user_id}/ban)
func (e *Endpoints) PutPlatformUserBan(c *gin.Context, platformId api.PlatformId, platformUserId api.PlatformUserId) {
	var reqBody api.Ban
	// decode request
	err := c.Bind(&reqBody)
	if err != nil {
		ThrowReqError(c, err.Error(), err, http.StatusBadRequest)
		return
	}

	err = e.banService.BanServiceUser(reqBody.Until, reqBody.Reason, platformId, platformUserId)
	if err != nil {
		ThrowReqError(c, err.Error(), err, http.StatusInternalServerError)
	} else {
		c.Status(http.StatusCreated)
	}
}

// (POST /v1/platform/{platform_id}/users/{platform_user_id}/refresh)
func (e *Endpoints) PostPlatformUserRefresh(c *gin.Context, platformId api.PlatformId, platformUserId api.PlatformUserId) {
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
		c.Status(http.StatusNotFound)
		return
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

	e.GetPlatformUser(c, platformId, platformUserId, api.GetPlatformUserParams{})
}
