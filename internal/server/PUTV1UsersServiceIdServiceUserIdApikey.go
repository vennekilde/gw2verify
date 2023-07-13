package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/vennekilde/gw2verify/internal/api"
	"github.com/vennekilde/gw2verify/pkg/verify"
	"gitlab.com/MrGunflame/gw2api"
)

// PUTV1UsersServiceIdServiceUserIdApikey is the handler for PUT /v1/users/{service_id}/{service_user_id}/apikey
// Set a service user's API key
func (h *RESTHandler) PUTV1UsersServiceIdServiceUserIdApikey(c *gin.Context, serviceId int, serviceUserId string, params api.PUTV1UsersServiceIdServiceUserIdApikeyParams) {
	var reqBody api.APIKeyData

	worldPerspective := HARD_CODED_WORLD_PERSPECTIVE
	if params.World != nil {
		worldPerspective = *params.World
	}

	// decode request
	err := c.Bind(&reqBody)
	if err != nil {
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
	err, userErr := verify.SetAPIKeyByUserService(gw2a, worldPerspective, serviceId, serviceUserId, reqBody.Primary, reqBody.Apikey, skipRequirements)
	if err != nil {
		ThrowReqError(c, err.Error(), userErr, http.StatusInternalServerError)
		return
	}
	c.Status(http.StatusCreated)
}
