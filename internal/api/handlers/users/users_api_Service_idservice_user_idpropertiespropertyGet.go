// THIS FILE IS SAFE TO EDIT. It will not be overwritten when rerunning go-raml.
package users

import (
	"encoding/json"
	"net/http"

	"github.com/vennekilde/gw2verify/internal/api/types"
	"github.com/vennekilde/gw2verify/internal/apiservice"
)

// Service_idservice_user_idpropertiespropertyGet is the handler for GET /users/{service_id}/{service_user_id}/properties/{property}
// Get a user property
func (api UsersAPI) Service_idservice_user_idpropertiespropertyGet(w http.ResponseWriter, r *http.Request) { // name := req.FormValue("name")
	if apiservice.Permitted(w, r) == false {
		return
	}
	w.Header().Set("Content-Type", "application/json")
	var respBody types.Property
	json.NewEncoder(w).Encode(&respBody)
}
