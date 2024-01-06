package server

import (
	"context"
	"errors"
	"net/http"
	"time"

	middleware "github.com/deepmap/oapi-codegen/pkg/gin-middleware"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/gin-gonic/gin"
	"github.com/patrickmn/go-cache"
	"github.com/vennekilde/gw2verify/v2/internal/orm"
	"go.uber.org/zap"
)

type Service struct {
	Uuid   string
	Name   string
	ApiKey string
}

type TokenMiddleware struct {
	serviceCache *cache.Cache
}

// NewTokenMiddleware returns a new instance of the TokenMiddleware handler
func NewTokenMiddleware() *TokenMiddleware {
	c := cache.New(5*time.Minute, 10*time.Minute)
	return &TokenMiddleware{
		serviceCache: c,
	}
}

// TokenRequestValidator validates that the Token provided in the request Authorization header is valid
func (m *TokenMiddleware) TokenRequestValidator(c *gin.Context) {
	bearer := c.GetHeader("Authorization")
	authenticated, serviceID := m.checkBearer(bearer)
	if !authenticated {
		zap.L().Warn("unable to verify token from request",
			zap.String("request uri", c.Request.RequestURI),
			zap.String("remote addr", c.Request.RemoteAddr),
			zap.String("bearer", bearer))
		c.AbortWithStatus(http.StatusForbidden)
	}
	gctx := c.Value(middleware.GinContextKey).(*gin.Context)
	gctx.Set("service_id", serviceID)
}

// TokenRequestValidator validates that the Token provided in the request Authorization header is valid
func (m *TokenMiddleware) OpenapiAuthenticator(c context.Context, input *openapi3filter.AuthenticationInput) error {
	bearer := input.RequestValidationInput.Request.Header.Get("Authorization")
	authenticated, serviceID := m.checkBearer(bearer)
	if !authenticated {
		zap.L().Warn("unable to verify token from request",
			zap.String("request uri", input.RequestValidationInput.Request.RequestURI),
			zap.String("remote addr", input.RequestValidationInput.Request.RemoteAddr),
			zap.String("bearer", bearer))
		return errors.New("unauthorized")
	}
	gctx := c.Value(middleware.GinContextKey).(*gin.Context)
	gctx.Set("service_id", serviceID)
	return nil
}

// TokenRequestValidator validates that the Token provided in the request Authorization header is valid
func (m *TokenMiddleware) checkBearer(bearer string) (permitted bool, serviceID string) {
	ctx := context.TODO()

	// Remove bearer prefix
	if len(bearer) <= 7 {
		return false, serviceID
	}
	bearer = bearer[7:]

	// Check if we have cached the bearer
	serviceIDInt, ok := m.serviceCache.Get(bearer)
	if ok {
		serviceID = serviceIDInt.(string)
		return true, serviceID
	}

	var service Service
	err := orm.DB().NewSelect().
		Model(&service).
		Where("api_key = ?", bearer).
		Scan(ctx)
	if err != nil {
		zap.L().Error("unable to verify token", zap.String("bearer", bearer), zap.Error(err))
		return false, serviceID
	}

	if service.Uuid != "" {
		m.serviceCache.Add(bearer, service.Uuid, cache.DefaultExpiration)
	}

	return service.ApiKey == bearer, service.Uuid
}
