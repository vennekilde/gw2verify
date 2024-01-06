package server

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/vennekilde/gw2verify/v2/internal/api"
	"github.com/vennekilde/gw2verify/v2/internal/config"
	"github.com/vennekilde/gw2verify/v2/pkg/verify"
	"go.uber.org/zap"
)

// Endpoints is responsible for handling all REST requests related to the service
type Endpoints struct {
	*VerificationEndpoint
}

func NewEndpoints(verificationEndpoint *VerificationEndpoint) *Endpoints {
	return &Endpoints{
		VerificationEndpoint: verificationEndpoint,
	}
}

func ThrowReqError(c *gin.Context, errorMsg string, userErr error, statusCode int) {
	jsonErr := make(map[string]string)
	jsonErr["error"] = errorMsg
	if userErr != nil {
		jsonErr["safe-display-error"] = userErr.Error()
	}
	zap.L().Warn("error while processing request",
		zap.String("request uri", c.Request.RequestURI),
		zap.String("remote addr", c.Request.RemoteAddr),
		zap.String("error", errorMsg))
	c.JSON(statusCode, &jsonErr)
}

func RespWithSuccess(c *gin.Context, respBody interface{}) {
	c.JSON(http.StatusOK, &respBody)
}

// h *RESTHandler api.ServerInterface

// GetV1Configuration is the handler for GET /v1/configuration
// Get a configuration containing relevant information for running a service bot
func (e *Endpoints) GetV1Configuration(c *gin.Context, params api.GetV1ConfigurationParams) {
	// Fetch linked worlds
	var links verify.LinkedWorlds
	if params.World != nil {
		// Get individual link
		worlds, err := e.worlds.GetWorldLinks(*params.World)
		if err != nil {
			c.Writer.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		links = verify.LinkedWorlds{}
		links[strconv.Itoa(*params.World)] = worlds
	} else {
		// Get all links
		links = e.worlds.GetAllWorldLinks()
	}

	respBody := api.Configuration{
		ExpirationTime:                config.Config().ExpirationTime,
		TemporaryAccessExpirationTime: config.Config().TemporaryAccessExpirationTime,
		WorldLinks:                    links,
	}
	c.JSON(http.StatusOK, respBody)
}
