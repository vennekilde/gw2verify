package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/vennekilde/gw2verify/internal/api"
)

// GETV1UsersServiceIdServiceUserIdProperties is the handler for GET /v1/users/{service_id}/{service_user_id}/properties
// Get all user properties
func (h *RESTHandler) GETV1UsersServiceIdServiceUserIdProperties(c *gin.Context, serviceId int, serviceUserId string) {
	var respBody []api.Property
	c.JSON(http.StatusOK, respBody)
}
