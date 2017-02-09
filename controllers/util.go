package controllers

import (
	"io/ioutil"

	"github.com/gin-gonic/gin"
)

func handleMongoError(c *gin.Context, err error) {
	if err != nil {
		c.AbortWithError(500, err)
	}
}

func getJSONBody(c *gin.Context) string {

	defer c.Request.Body.Close()
	body, err := ioutil.ReadAll(c.Request.Body)

	if err != nil {
		c.AbortWithError(400, err)
	}

	return string(body)

}
