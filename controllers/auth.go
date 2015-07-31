package controllers

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/gob"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gorilla/sessions"
	"github.com/intervention-engine/fhir/server"
	"github.com/intervention-engine/ie/models"
	"gopkg.in/mgo.v2/bson"
)

var Store = sessions.NewCookieStore([]byte("somethingsecret"))

func init() {
	gob.Register(bson.ObjectId(""))
}

type request_form struct {
	Session session_request_form
}

type session_request_form struct {
	Identification string
	Password       string
}

type response_form struct {
	Session session_response_form `json:"session"`
}

type session_response_form struct {
	Token string `json:"token"`
}

type error_form struct {
	Error string `json:"error"`
}

type registration_request_form struct {
	Identification string
	Password       string
	Confirmation   string
}

func LoginHandler(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {

	decoder := json.NewDecoder(r.Body)
	var reqform request_form
	err := decoder.Decode(&reqform)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
	}

	//grab the username and password from the form
	username, password := reqform.Session.Identification, reqform.Session.Password

	user, err := models.Login(username, password)

	if err != nil {
		//var errform error_form
		//errform.Error = "invalid credentials"
		rw.Header().Set("Content-Type", "application/json; charset=utf-8")
		//json.NewEncoder(rw).Encode(errform)
		http.Error(rw, "{\"error\":\"invalid credentials\"}", 422)
		return
	}

	if user != nil {
		var respform response_form
		token := generate_token()

		var usersession models.UserSession
		usersession.User = *user
		usersession.Token = token
		usersession.Expiration = time.Now().Add(90 * time.Minute)
		server.Database.C("sessions").Insert(usersession)

		respform.Session.Token = token
		rw.Header().Set("Content-Type", "application/json; charset=utf-8")
		json.NewEncoder(rw).Encode(respform)
	}
}

func generate_token() string {
	rb := make([]byte, 32)
	_, err := rand.Read(rb)

	if err != nil {
		panic(err)
	}

	token := base64.URLEncoding.EncodeToString(rb)
	return token
}

func LogoutHandler(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	client_token := r.Header.Get("Authorization")
	sessionCollection := server.Database.C("sessions")

	err := sessionCollection.Remove(bson.M{"token": client_token})

	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
	}
}

func RegisterHandler(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {

	decoder := json.NewDecoder(r.Body)
	var regform registration_request_form
	err := decoder.Decode(&regform)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
	}

	//grab username, password, and password confirmation from object
	username, password, confirmation := regform.Identification, regform.Password, regform.Confirmation

	count, err := server.Database.C("users").Find(bson.M{"username": username}).Count()
	if count != 0 {
		rw.Header().Set("Content-Type", "application/json; charset=utf-8")
		http.Error(rw, "{\"error\":\"username already exists\"}", 422)
		return
	}

	if password != confirmation {
		rw.Header().Set("Content-Type", "application/json; charset=utf-8")
		http.Error(rw, "{\"error\":\"password and confirmation must match\"}", 422)
		return
	}

	newuser := &models.User{
		Username: username,
		ID:       bson.NewObjectId(),
	}
	newuser.SetPassword(password)

	err = server.Database.C("users").Insert(newuser)
	if err != nil {
		http.Error(rw, "{\"error\":\"problem registering user\"}", http.StatusInternalServerError)
		return
	}
}
