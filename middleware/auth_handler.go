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

func AuthHandler(c *echo.Context) error {
	clientToken := c.Request().Header.Get("Authorization")

	sessionCollection := server.Database.C("sessions")
	usersession := models.UserSession{}

	err := sessionCollection.Find(bson.M{"token": clientToken}).One(&usersession)

	if err != nil {
		return err
	} else if usersession.Expiration.Before(time.Now()) {
		return c.String(http.StatusUnauthorized, "Session Expired")
	}
	fmt.Println("User found and token matched")
	return nil
}
