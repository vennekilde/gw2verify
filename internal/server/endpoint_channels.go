package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/vennekilde/gw2verify/v2/internal/api"
)

// (POST /v1/channels/{platform_id}/{channel}/statistics)
func (e *Endpoints) PostChannelPlatformStatistics(c *gin.Context, platformId api.PlatformId, channel string, params api.PostChannelPlatformStatisticsParams) {
	c.Header("Content-Type", "application/json")

	var req api.ChannelMetadata
	err := c.Bind(&req)
	if err != nil {
		ThrowReqError(c, err.Error(), err, http.StatusBadRequest)
		return
	}

	if err := e.statistics.WorldStatistics(platformId, channel, params.World, req); err != nil {
		ThrowReqError(c, err.Error(), err, http.StatusInternalServerError)
		return
	}
}
