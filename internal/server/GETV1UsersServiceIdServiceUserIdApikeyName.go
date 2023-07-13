package server

import (
	"github.com/gin-gonic/gin"
	"github.com/vennekilde/gw2verify/internal/api"
	"github.com/vennekilde/gw2verify/pkg/verify"
)

// GETV1UsersServiceIdServiceUserIdApikeyName is the handler for GET /v1/users/{service_id}/{service_user_id}/apikey/name
// Get a service user's apikey name they are required to use if apikey name restriction is enforced
func (h *RESTHandler) GETV1UsersServiceIdServiceUserIdApikeyName(c *gin.Context, serviceId int, serviceUserId string, params api.GETV1UsersServiceIdServiceUserIdApikeyNameParams) {
	worldPerspective := HARD_CODED_WORLD_PERSPECTIVE
	if params.World != nil {
		worldPerspective = *params.World
	}

	apikeyName := verify.GetAPIKeyName(worldPerspective, serviceId, serviceUserId)
	RespWithSuccess(c, api.APIKeyName{Name: apikeyName})
}
