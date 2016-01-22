package middleware

import (
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
			// For now, only authenticate on GET.  Need to lock it down more later, but will need to update
			// the generate and upload tools at the same time.
			if c.Request().Method != "GET" {
				return hf(c)
			}

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
			return hf(c)
		}
	}
}
