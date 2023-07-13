package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/vennekilde/gw2verify/internal/api"
	"github.com/vennekilde/gw2verify/pkg/history"
)

// POSTV1ChannelsServiceIdChannelStatistics is the handler for POST /v1/channels/{service_id}/{channel}/statistics
// Collect statistics based on the provided parameters and save them for historical purposes
func (h *RESTHandler) POSTV1ChannelsServiceIdChannelStatistics(c *gin.Context, serviceId int, channel string, params api.POSTV1ChannelsServiceIdChannelStatisticsParams) {
	c.Header("Content-Type", "application/json")

	var req api.ChannelMetadata
	err := c.Bind(&req)
	if err != nil {
		return
	}

	worldPerspective := HARD_CODED_WORLD_PERSPECTIVE
	if params.World != nil {
		worldPerspective = *params.World
	}

	if err := history.CollectChannelStatistics(serviceId, channel, worldPerspective, req); err != nil {
		ThrowReqError(c, err.Error(), err, http.StatusInternalServerError)
		return
	}
}
