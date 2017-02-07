package controllers

import "github.com/gin-gonic/gin"

// Controller interface for all controllers in IE use for Routing
type Controller interface {
	All(c *gin.Context)
	Create(c *gin.Context)
	Read(c *gin.Context)
	Update(c *gin.Context)
	Delete(c *gin.Context)
}

// RegisterController a controller with a router
func RegisterController(route string, e *gin.RouterGroup, c Controller) {
	e.GET(route, c.All)
	e.POST(route, c.Create)

	var instanceRoute = route + "/:id"

	e.PUT(instanceRoute, c.Update)
	e.GET(instanceRoute, c.Read)
	e.DELETE(instanceRoute, c.Delete)
}
