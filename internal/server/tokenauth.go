package server

import (
	"context"
	"errors"
	"net/http"

	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/gin-gonic/gin"
	"github.com/vennekilde/gw2verify/internal/config"
	"go.uber.org/zap"
)

type TokenMiddleware struct {
}

// NewTokenMiddleware returns a new instance of the TokenMiddleware handler
func NewTokenMiddleware() *TokenMiddleware {
	return &TokenMiddleware{}
}

// TokenRequestValidator validates that the Token provided in the request Authorization header is valid
func (m *TokenMiddleware) TokenRequestValidator(c *gin.Context) {
	bearer := c.GetHeader("Authorization")
	authenticated := m.checkBearer(bearer)
	if !authenticated {
		zap.L().Warn("unable to verify token from request",
			zap.String("request uri", c.Request.RequestURI),
			zap.String("remote addr", c.Request.RemoteAddr),
			zap.String("bearer", bearer))
		c.AbortWithStatus(http.StatusForbidden)
	}
}

// TokenRequestValidator validates that the Token provided in the request Authorization header is valid
func (m *TokenMiddleware) OpenapiAuthenticator(c context.Context, input *openapi3filter.AuthenticationInput) error {
	bearer := input.RequestValidationInput.Request.Header.Get("Authorization")
	authenticated := m.checkBearer(bearer)
	if !authenticated {
		zap.L().Warn("unable to verify token from request",
			zap.String("request uri", input.RequestValidationInput.Request.RequestURI),
			zap.String("remote addr", input.RequestValidationInput.Request.RemoteAddr),
			zap.String("bearer", bearer))
		return errors.New("unauthorized")
	}
	return nil
}

// TokenRequestValidator validates that the Token provided in the request Authorization header is valid
func (m *TokenMiddleware) checkBearer(bearer string) bool {
	accessToken := config.Config().RESTAuthToken
	return accessToken == bearer
}
