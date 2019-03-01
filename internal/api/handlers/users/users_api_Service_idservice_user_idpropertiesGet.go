// THIS FILE IS SAFE TO EDIT. It will not be overwritten when rerunning go-raml.
package users

import (
	"encoding/json"
	"github.com/vennekilde/gw2verify/internal/api/types"
	"net/http"
)

// Service_idservice_user_idpropertiesGet is the handler for GET /users/{service_id}/{service_user_id}/properties
// Get all user properties
func (api UsersAPI) Service_idservice_user_idpropertiesGet(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var respBody []types.Property
	json.NewEncoder(w).Encode(&respBody)
}
