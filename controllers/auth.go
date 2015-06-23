package controllers

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"github.com/gorilla/sessions"
	"github.com/intervention-engine/fhir/server"
	"github.com/intervention-engine/ie/models"
	"gopkg.in/mgo.v2/bson"
	"html/template"
	"net/http"
	"time"
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
	username, password, confirm := r.FormValue("username"), r.FormValue("password"), r.FormValue("confirm")
	sess, err := Store.Get(r, "intervention-engine")

	if password != confirm {
		sess.AddFlash("Password and confirmation must match.")
		sess.Save(r, rw)
		http.Redirect(rw, r, "/register", http.StatusSeeOther)
		return
	}

	u := &models.User{
		Username: username,
		ID:       bson.NewObjectId(),
	}
	u.SetPassword(password)

	err = server.Database.C("users").Insert(u)
	if err != nil {
		sess.AddFlash("Problem registering user.")
		http.Redirect(rw, r, "/register", http.StatusInternalServerError)
		return
	}

	sess.Values["user"] = u.ID
	sess.Save(r, rw)
	fmt.Fprintf(rw, "Success registering user.")
}

func RegisterForm(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	registerForms, _ := template.ParseFiles("templates/_base.html", "templates/register.html")
	sess, err := Store.Get(r, "intervention-engine")
	flashes := sess.Flashes()
	sess.Save(r, rw)
	err = registerForms.Execute(rw, map[string]interface{}{"flashes": flashes})
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
	}
}
