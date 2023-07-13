package server

import (
	"github.com/gin-gonic/gin"
	"github.com/vennekilde/gw2verify/internal/api"
)

// PUTV1UsersServiceIdServiceUserIdProperties is the handler for PUT /v1/users/{service_id}/{service_user_id}/properties
// Set a user property
func (h *RESTHandler) PUTV1UsersServiceIdServiceUserIdProperties(c *gin.Context, serviceId int, serviceUserId string, params api.PUTV1UsersServiceIdServiceUserIdPropertiesParams) { // name := req.FormValue("name")// value := req.FormValue("value")
	//URL Params
	// params := mux.Vars(r)
	// serviceID := params["service_id"]
	// serviceUserId := params["service_user_id"]
	// serviceId, err := strconv.Atoi(serviceID)
	// if err != nil {
	// 	w.WriteHeader(http.StatusBadRequest)
	// 	return
	// }

	// value := r.FormValue("value")
}
