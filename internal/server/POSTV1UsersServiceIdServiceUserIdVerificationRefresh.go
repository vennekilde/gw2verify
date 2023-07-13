package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/vennekilde/gw2verify/internal/api"
	"github.com/vennekilde/gw2verify/internal/orm"
	"github.com/vennekilde/gw2verify/pkg/sync"
	"github.com/vennekilde/gw2verify/pkg/verify"
	"gitlab.com/MrGunflame/gw2api"
)

// POSTV1UsersServiceIdServiceUserIdVerificationRefresh is the handler for POST /v1/users/{service_id}/{service_user_id}/verification/refresh
// Forces a refresh of the API data and returns the new verification status after the API data has been refreshed. Note this can take a few seconds
func (h *RESTHandler) POSTV1UsersServiceIdServiceUserIdVerificationRefresh(c *gin.Context, serviceId int, serviceUserId string, params api.POSTV1UsersServiceIdServiceUserIdVerificationRefreshParams) {
	worldPerspective := HARD_CODED_WORLD_PERSPECTIVE
	if params.World != nil {
		worldPerspective = *params.World
	}

	tx, err := orm.DB().Begin()
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

	gw2a := gw2api.New()
	// Refresh user status
	err, userErr := sync.SynchronizeLinkedUser(tx, gw2a, serviceId, serviceUserId)
	if err != nil {
		ThrowReqError(c, err.Error(), userErr, http.StatusInternalServerError)
		return
	}
	status, err := verify.Status(worldPerspective, serviceId, serviceUserId)
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

	c.JSON(http.StatusOK, &status)
}
