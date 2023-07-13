package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/vennekilde/gw2verify/internal/api"
	"github.com/vennekilde/gw2verify/internal/config"
	"github.com/vennekilde/gw2verify/pkg/verify"
)

// PUTV1UsersServiceIdServiceUserIdVerificationTemporary is the handler for PUT /v1/users/{service_id}/{service_user_id}/verification/temporary
// Grant a user temporary world relation. Additionally, the "temp_expired" property will be removed from the user's properties
func (h *RESTHandler) PUTV1UsersServiceIdServiceUserIdVerificationTemporary(c *gin.Context, serviceId int, serviceUserId string, params api.PUTV1UsersServiceIdServiceUserIdVerificationTemporaryParams) {
	var reqBody api.TemporaryData
	worldPerspective := HARD_CODED_WORLD_PERSPECTIVE
	if params.World != nil {
		worldPerspective = *params.World
	}

	// decode request
	err := c.Bind(&reqBody)
	if err != nil {
		return
	}

	var world int
	if reqBody.World != nil && *reqBody.World > 0 {
		world = *reqBody.World
	} else if reqBody.AccessType != nil {
		if *reqBody.AccessType == api.HOMEWORLD {
			// Grant Home World temporary access
			world = worldPerspective
		} else if *reqBody.AccessType == api.LINKEDWORLD {
			// Grant Linked World temporary access
			worldLinks, err := verify.GetWorldLinks(worldPerspective)
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

	err, userErr := verify.GrantTemporaryWorldAssignment(serviceId, serviceUserId, world)
	if err != nil {
		ThrowReqError(c, err.Error(), userErr, http.StatusInternalServerError)
		return
	}

	respBody := config.Config().TemporaryAccessExpirationTime
	c.JSON(200, &respBody)
}
