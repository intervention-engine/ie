package controllers

import (
  "net/http"
  "html/template"
  "github.com/intervention-engine/ie/models"
  "github.com/intervention-engine/fhir/server"
  "gopkg.in/mgo.v2/bson"
  "github.com/gorilla/sessions"
  "encoding/gob"
  )

var Store = sessions.NewCookieStore([]byte("somethingsecret"))

func init() {
  gob.Register(bson.ObjectId(""))
}

func LoginHandler(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
  //grab the username and password from the form
  username, password := r.FormValue("username"), r.FormValue("password")

  sess, err := Store.Get(r, "intervention-engine")
  user, err := models.Login(username, password)

  if err != nil {
    sess.AddFlash("Invalid Username/Password")
    sess.Save(r, rw)
    http.Redirect(rw, r, "/login", http.StatusSeeOther)
    return
  }

  if user != nil {
    sess.Values["user"] = user.ID
    sess.Save(r, rw)
  }
  http.Redirect(rw, r, "/index", http.StatusSeeOther)
}

func LoginForm(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
  loginForms, _ := template.ParseFiles("templates/_base.html","templates/login.html")
  sess, err := Store.Get(r, "intervention-engine")
  flashes := sess.Flashes()
  sess.Save(r, rw)
  err = loginForms.Execute(rw, map[string]interface{}{"flashes": flashes})
  if err != nil {
    http.Error(rw, err.Error(), http.StatusInternalServerError)
  }
  sess.Save(r, rw)
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
    ID: bson.NewObjectId(),
  }
  u.SetPassword(password)

  err = server.Database.C("users").Insert(u)
  if err != nil {
    sess.AddFlash("Problem registering user.")
    http.Redirect(rw, r, "/register", http.StatusInternalServerError)
    return
  }

  sess.Values["user"] = u.ID
  sess.AddFlash("Successfully registered user.")
  sess.Save(r, rw)
  http.Redirect(rw, r, "/login", http.StatusSeeOther)
}

func RegisterForm(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
  registerForms, _ := template.ParseFiles("templates/_base.html","templates/register.html")
  sess, err := Store.Get(r, "intervention-engine")
  flashes := sess.Flashes()
  sess.Save(r, rw)
  err = registerForms.Execute(rw, map[string]interface{}{"flashes": flashes})
  if err != nil {
    http.Error(rw, err.Error(), http.StatusInternalServerError)
  }
}

func IndexHandler(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc){
  indexForms, _ := template.ParseFiles("templates/_base.html","templates/index.html")
  sess, err := Store.Get(r, "intervention-engine")
  flashes := sess.Flashes()
  sess.Save(r, rw)
  err = indexForms.Execute(rw, map[string]interface{}{"flashes": flashes})
  if err != nil {
    http.Error(rw, err.Error(), http.StatusInternalServerError)
  }
}
