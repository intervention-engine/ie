package middleware

import (
	"fmt"
	"net/http"
	"time"

	"github.com/intervention-engine/fhir/server"
	"github.com/intervention-engine/ie/models"
	"github.com/labstack/echo"
	"gopkg.in/mgo.v2/bson"
)

func AuthHandler() echo.MiddlewareFunc {
	return func(hf echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			clientToken := c.Request().Header.Get("Authorization")

			if clientToken == "" {
				return c.String(http.StatusUnauthorized, "No authorization provided")
			}

			sessionCollection := server.Database.C("sessions")
			usersession := models.UserSession{}

			err := sessionCollection.Find(bson.M{"token": clientToken}).One(&usersession)

			if err != nil {
				return c.String(http.StatusUnauthorized, "Unable to validate authorization token")
			} else if usersession.Expiration.Before(time.Now()) {
				return c.String(http.StatusUnauthorized, "Session Expired")
			}
			fmt.Println("User found and token matched")
			return hf(c)
		}
	}
}
