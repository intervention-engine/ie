package controllers

import (
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
)

func GeneratePieHandler(riskEndpoint string) func(c *gin.Context) {
	f := func(c *gin.Context) {
		idString := c.Param("id")
		piesEndpoint := riskEndpoint + "/pies/" + idString
		resp, err := http.Get(piesEndpoint)

		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		if _, err := io.Copy(c.Writer, resp.Body); err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		c.Status(resp.StatusCode)
		c.Header("Content-Type", resp.Header.Get("Content-Type"))
	}
	return f
}
