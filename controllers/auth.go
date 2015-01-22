package controllers

import (
  "net/http"
  "html/template"
  "github.com/intervention-engine/ie/models"
  "github.com/intervention-engine/fhir/server"
  "gopkg.in/mgo.v2/bson"
  "github.com/gorilla/sessions"
  )

var Store = sessions.NewCookieStore([]byte("somethingsecret"))

func LoginHandler(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
  //grab the username and password from the form
  username, password := r.FormValue("username"), r.FormValue("password")

  user, err := models.Login(username, password)
  sess, err := Store.Get(r, "intervention-engine")

  if err != nil {
    sess.AddFlash("Invalid Username/Password")
  }

  if user != nil {
    sess.Values["user"] = user.ID
    sess.Save(r, rw)
  }
  http.Redirect(rw, r, "/login", http.StatusSeeOther)
}

func LoginForm(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
  loginForms, _ := template.ParseFiles("templates/_base.html","templates/login.html")
  sess, err := Store.Get(r, "intervention-engine")
  err = loginForms.Execute(rw, map[string]interface{}{"session": sess})
  if err != nil {
    http.Error(rw, err.Error(), http.StatusInternalServerError)
  }
}

func LogoutHandler(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
  sess, err := Store.Get(r, "intervention-engine")
  delete(sess.Values, "user")
  if err != nil {
    http.Error(rw, err.Error(), http.StatusInternalServerError)
  }
  sess.Save(r, rw)
  http.Redirect(rw, r, "/", http.StatusSeeOther)
}

func RegisterHandler(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
  username, password := r.FormValue("username"), r.FormValue("password")
  sess, err := Store.Get(r, "intervention-engine")

  u := &models.User{
    Username: username,
    ID: bson.NewObjectId(),
  }
  u.SetPassword(password)

  err = server.Database.C("users").Insert(u)
  if err != nil {
    sess.AddFlash("Problem registering user.")
    RegisterForm(rw, r, next)
  }

  sess.Values["user"] = u.ID
  sess.Save(r, rw)
  http.Redirect(rw, r, "/register", http.StatusSeeOther)
}

func RegisterForm(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
  registerForms, _ := template.ParseFiles("templates/_base.html","templates/register.html")
  err := registerForms.Execute(rw, nil)
  if err != nil {
    http.Error(rw, err.Error(), http.StatusInternalServerError)
  }
}
