package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/vennekilde/gw2verify/internal/api"
	"github.com/vennekilde/gw2verify/internal/orm"
	"github.com/vennekilde/gw2verify/pkg/verify"
)

type VerificationStatusExtended struct {
	api.VerificationStatus
	AccountData orm.Account
}

// GETV1UsersServiceIdServiceUserIdVerificationStatus is the handler for GET /v1/users/{service_id}/{service_user_id}/verification/status
// Get a users verification status
func (h *RESTHandler) GETV1UsersServiceIdServiceUserIdVerificationStatus(c *gin.Context, serviceId int, serviceUserId string, params api.GETV1UsersServiceIdServiceUserIdVerificationStatusParams) {
	worldPerspective := HARD_CODED_WORLD_PERSPECTIVE
	if params.World != nil {
		worldPerspective = *params.World
	}

	status, err := verify.Status(worldPerspective, serviceId, serviceUserId)
	if err != nil {
		ThrowReqError(c, err.Error(), err, http.StatusInternalServerError)
		return
	}
	c.JSON(http.StatusOK, &status)
}
