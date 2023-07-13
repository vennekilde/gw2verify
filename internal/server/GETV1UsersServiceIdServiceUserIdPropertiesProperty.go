package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/vennekilde/gw2verify/internal/api"
)

// GETV1UsersServiceIdServiceUserIdPropertiesProperty is the handler for GET /v1/users/{service_id}/{service_user_id}/properties/{property}
// Get a user property
func (h *RESTHandler) GETV1UsersServiceIdServiceUserIdPropertiesProperty(c *gin.Context, serviceId int, serviceUserId string, property string, params api.GETV1UsersServiceIdServiceUserIdPropertiesPropertyParams) {
	// name := req.FormValue("name")
	var respBody api.Property
	c.JSON(http.StatusOK, &respBody)
}
