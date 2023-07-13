package server

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/vennekilde/gw2verify/internal/api"
	"github.com/vennekilde/gw2verify/pkg/verify"
)

// GETV1UpdatesServiceIdSubscribe is the handler for GET /v1/updates/{service_id}/subscribe
// Long polling rest endpoint for receiving verification updates
func (h *RESTHandler) GETV1UpdatesServiceIdSubscribe(c *gin.Context, serviceId int, params api.GETV1UpdatesServiceIdSubscribeParams) {
	ticker := time.NewTicker(120 * time.Second)
	defer func() { ticker.Stop() }()

	worldPerspective := HARD_CODED_WORLD_PERSPECTIVE
	if params.World != nil {
		worldPerspective = *params.World
	}

	statusListener := verify.ServicePollListeners[serviceId]
	if statusListener.Listener == nil {
		statusListener = verify.VerifictionStatusListener{
			ServiceID:        serviceId,
			WorldPerspective: worldPerspective,
			Listener:         make(chan *api.VerificationStatus),
		}
		verify.ServicePollListeners[serviceId] = statusListener
	}

	select {
	case event := <-statusListener.Listener:
		c.JSON(http.StatusOK, event)
	case <-ticker.C:
		c.Status(http.StatusRequestTimeout)
	}
	statusListener.Listener = nil
}
