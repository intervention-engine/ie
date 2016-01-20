package controllers

import (
	"io"
	"net/http"

	"github.com/labstack/echo"
)

func GeneratePieHandler(riskEndpoint string) func(c *echo.Context) error {
	f := func(c *echo.Context) error {
		idString := c.Param("id")
		piesEndpoint := riskEndpoint + "/pies/" + idString
		resp, err := http.Get(piesEndpoint)

		if err != nil {
			return err
		}

		_, err = io.Copy(c.Response(), resp.Body)
		return err
	}
	return f
}
