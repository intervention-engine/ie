package controllers

import (
	"encoding/json"
	//"github.com/intervention-engine/fhir/server"
	"github.com/intervention-engine/ie/models"
	//"gopkg.in/mgo.v2/bson"
	//"html/template"
	"net/http"
)

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
		var errform error_form
		errform.Error = "invalid credentials"
		rw.Header().Set("Content-Type", "application/json; charset=utf-8")
		json.NewEncoder(rw).Encode(errform)
		return
	}

	if user != nil {
		var respform response_form
		respform.Session.Token = "secret token!"
		rw.Header().Set("Content-Type", "application/json; charset=utf-8")
		json.NewEncoder(rw).Encode(respform)
	}
}

// func LogoutHandler(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
// 	sess, err := Store.Get(r, "intervention-engine")
// 	delete(sess.Values, "user")
// 	sess.AddFlash("You have been logged out.")
// 	if err != nil {
// 		http.Error(rw, err.Error(), http.StatusInternalServerError)
// 	}
// 	sess.Save(r, rw)
// 	http.Redirect(rw, r, "/login", http.StatusSeeOther)
// }
//
// func RegisterHandler(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
// 	username, password, confirm := r.FormValue("username"), r.FormValue("password"), r.FormValue("confirm")
// 	sess, err := Store.Get(r, "intervention-engine")
//
// 	if password != confirm {
// 		sess.AddFlash("Password and confirmation must match.")
// 		sess.Save(r, rw)
// 		http.Redirect(rw, r, "/register", http.StatusSeeOther)
// 		return
// 	}
//
// 	u := &models.User{
// 		Username: username,
// 		ID:       bson.NewObjectId(),
// 	}
// 	u.SetPassword(password)
//
// 	err = server.Database.C("users").Insert(u)
// 	if err != nil {
// 		sess.AddFlash("Problem registering user.")
// 		http.Redirect(rw, r, "/register", http.StatusInternalServerError)
// 		return
// 	}
//
// 	sess.Values["user"] = u.ID
// 	sess.AddFlash("Successfully registered user.")
// 	sess.Save(r, rw)
// 	http.Redirect(rw, r, "/login", http.StatusSeeOther)
// }
//
// func RegisterForm(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
// 	registerForms, _ := template.ParseFiles("templates/_base.html", "templates/register.html")
// 	sess, err := Store.Get(r, "intervention-engine")
// 	flashes := sess.Flashes()
// 	sess.Save(r, rw)
// 	err = registerForms.Execute(rw, map[string]interface{}{"flashes": flashes})
// 	if err != nil {
// 		http.Error(rw, err.Error(), http.StatusInternalServerError)
// 	}
// }
