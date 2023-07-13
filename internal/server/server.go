package server

import (
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/gin-gonic/gin"
	"github.com/vennekilde/gw2verify/internal/api"
	"go.uber.org/zap"

	middleware "github.com/deepmap/oapi-codegen/pkg/gin-middleware"
)

type RESTServer struct {
	engine  *gin.Engine
	handler *RESTHandler
}

// NewRESTServer returns a RESTServer instance configured to provide serve the verification REST API
func NewRESTServer() *RESTServer {
	r := gin.Default()
	s := &RESTServer{
		engine:  r,
		handler: &RESTHandler{},
	}

	// AuthN middleware for handling JWT tokens
	authNMiddleware := NewTokenMiddleware()

	// Openapi middleware for ensuring requests conform to the openapi spec
	swagger, err := api.GetSwagger()
	if err != nil {
		zap.L().Panic("unable to parse openapi spec", zap.Error(err))
	}
	// Do not validate server hostname
	swagger.Servers = nil

	// Set up an authenticator to check authenticated function. It will allow
	// access to "someScope", but disallow others.
	options := middleware.Options{
		/*ErrorHandler: func(c echo.Context, err *echo.HTTPError) error {
			return c.String(err.Code, "test: "+err.Error())
		},*/
		Options: openapi3filter.Options{
			AuthenticationFunc: authNMiddleware.OpenapiAuthenticator,
		},
		UserData: "hi!",
	}
	openapiValidator := middleware.OapiRequestValidatorWithOptions(swagger, &options)

	// Ensure requests conform to the openapi spec
	r.Use(openapiValidator)

	// Register service endpoint handlers
	api.RegisterHandlersWithOptions(r, s.handler, api.GinServerOptions{})
	return s
}

// Start launches the HTTP server
// Note: this method will block the calling goroutine indefinitely unless an error happens.
func (s *RESTServer) Start() {
	s.engine.Run("127.0.0.1:8080")
}
