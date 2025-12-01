package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// index basic request to /
// nolint: revive // do not check for unused param required by interface
func index(c *gin.Context) {
	responseData := map[string]interface{}{"message": "Welcome to the Paymail Server ✌(◕‿-)✌"}

	c.Header("Content-Type", "application/json")
	c.JSON(http.StatusOK, responseData)
}

// health is a basic request to return a health response
func health(c *gin.Context) {
	c.Status(http.StatusOK)
}
