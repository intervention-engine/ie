package middleware

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/intervention-engine/fhir/server"
	"github.com/intervention-engine/ie/models"
	"gopkg.in/mgo.v2/bson"
)

func AuthHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		// For now, only authenticate on GET.  Need to lock it down more later, but will need to update
		// the generate and upload tools at the same time.
		if c.Request.Method != "GET" {
			c.Next()
			return
		}

		clientToken := c.Request.Header.Get("Authorization")

		if clientToken == "" {
			c.AbortWithError(http.StatusUnauthorized, errors.New("No authorization provided"))
			return
		}

		sessionCollection := server.Database.C("sessions")
		usersession := models.UserSession{}

		if err := sessionCollection.Find(bson.M{"token": clientToken}).One(&usersession); err != nil {
			c.AbortWithError(http.StatusUnauthorized, errors.New("Unable to validate authorization token"))
			return
		} else if usersession.Expiration.Before(time.Now()) {
			c.AbortWithError(http.StatusUnauthorized, errors.New("Session Expired"))
			return
		}
		c.Next()
	}
}
