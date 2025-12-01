package server

import (
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
)

// Handlers are used to isolate loading the routes (used for testing)
func Handlers(configuration *Configuration) *gin.Engine {
	engine := gin.New()
	engine.Use(gin.LoggerWithWriter(configuration.Logger), gin.Recovery())

	configuration.RegisterBasicRoutes(engine)
	configuration.RegisterRoutes(engine)

	return engine
}

// RegisterBasicRoutes register the basic routes to the http router
func (c *Configuration) RegisterBasicRoutes(engine *gin.Engine) {
	// Skip if not set
	if c.BasicRoutes == nil {
		return
	}

	// Set the main index page (navigating to slash)
	if c.BasicRoutes.AddIndexRoute {
		engine.GET("/", index)
		// router.OPTIONS("/", router.SetCrossOriginHeaders) // Disabled for security
	}

	// Set the health request (used for load balancers)
	if c.BasicRoutes.AddHealthRoute {
		engine.GET("/health", health)
		engine.OPTIONS("/health", health)
		engine.HEAD("/health", health)
	}
}

// RegisterRoutes register all the available paymail routes to the http router
func (c *Configuration) RegisterRoutes(engine *gin.Engine) {
	engine.GET("/.well-known/"+c.ServiceName, c.showCapabilities) // service discovery

	for _, cap := range c.callableCapabilities {
		c.registerRoute(engine, cap)
	}

	for _, nestedCap := range c.nestedCapabilities {
		for _, cap := range nestedCap {
			c.registerRoute(engine, cap)
		}
	}
}

func (c *Configuration) registerRoute(engine *gin.Engine, cap CallableCapability) {
	routerPath := c.templateToRouterPath(cap.Path)
	engine.Handle(
		cap.Method,
		routerPath,
		cap.Handler,
	)
}

func (c *Configuration) templateToRouterPath(template string) string {
	template = strings.ReplaceAll(template, PaymailAddressTemplate, _routerParam(PaymailAddressParamName))
	template = strings.ReplaceAll(template, PubKeyTemplate, _routerParam(PubKeyParamName))
	return fmt.Sprintf("/%s/%s/%s", c.APIVersion, c.ServiceName, strings.TrimPrefix(template, "/"))
}

func _routerParam(name string) string {
	return ":" + name
}
