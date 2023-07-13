package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// @TODO fix hardcoded later
var HARD_CODED_WORLD_PERSPECTIVE = 2007

// RESTHandler is responsible for handling all REST requests related to the service
type RESTHandler struct {
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
