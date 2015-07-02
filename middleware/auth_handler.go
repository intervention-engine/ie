package middleware

import (
	"fmt"
	"github.com/intervention-engine/fhir/server"
	"github.com/intervention-engine/ie/models"
	"gopkg.in/mgo.v2/bson"
	"net/http"
	"time"
)

func AuthHandler(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	client_token := r.Header.Get("Authorization")

	sessionCollection := server.Database.C("sessions")
	usersession := models.UserSession{}

	err := sessionCollection.Find(bson.M{"token": client_token}).One(&usersession)

	if err != nil {
		http.Error(rw, err.Error(), http.StatusUnauthorized)
	} else if usersession.Expiration.Before(time.Now()) {
		http.Error(rw, "Session Expired", http.StatusUnauthorized)
	} else {
		fmt.Println("User found and token matched")
		next(rw, r)
	}

}
