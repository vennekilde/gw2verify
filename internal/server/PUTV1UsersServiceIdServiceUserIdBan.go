package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/vennekilde/gw2verify/internal/api"
	"github.com/vennekilde/gw2verify/pkg/verify"
)

// PUTV1UsersServiceIdServiceUserIdBan is the handler for PUT /v1/users/{service_id}/{service_user_id}/ban
// Ban a user's gw2 account from being verified
func (h *RESTHandler) PUTV1UsersServiceIdServiceUserIdBan(c *gin.Context, serviceId int, serviceUserId string) {

	var reqBody api.BanData
	// decode request
	err := c.Bind(&reqBody)
	if err != nil {
		return
	}

	err = verify.BanServiceUser(reqBody.Until, reqBody.Reason, serviceId, serviceUserId)
	if err != nil {
		ThrowReqError(c, err.Error(), err, http.StatusInternalServerError)
	} else {
		c.Status(http.StatusCreated)
	}
}
