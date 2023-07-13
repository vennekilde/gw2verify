package server

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/vennekilde/gw2verify/internal/api"
	"github.com/vennekilde/gw2verify/internal/config"
	"github.com/vennekilde/gw2verify/pkg/verify"
)

// GETV1Configuration is the handler for GET /v1/configuration
// Get a configuration containing relevant information for running a service bot
func (h *RESTHandler) GETV1Configuration(c *gin.Context, params api.GETV1ConfigurationParams) {
	// Fetch linked worlds
	var links verify.LinkedWorlds
	if params.World != nil {
		// Get individual link
		worlds, err := verify.GetWorldLinks(*params.World)
		if err != nil {
			c.Writer.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		links = verify.LinkedWorlds{}
		links[strconv.Itoa(*params.World)] = worlds
	} else {
		// Get all links
		links = verify.GetAllWorldLinks()
	}

	respBody := api.Configuration{
		ExpirationTime:                config.Config().ExpirationTime,
		TemporaryAccessExpirationTime: config.Config().TemporaryAccessExpirationTime,
		WorldLinks:                    links,
	}
	c.JSON(http.StatusOK, respBody)
}
