package controllers

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/gob"
	"net/http"
	"time"

	"github.com/gorilla/sessions"
	"github.com/intervention-engine/fhir/server"
	"github.com/intervention-engine/ie/models"
	"github.com/labstack/echo"
	"gopkg.in/mgo.v2/bson"
)

var Store = sessions.NewCookieStore([]byte("somethingsecret"))

func init() {
	gob.Register(bson.ObjectId(""))
}

type requestForm struct {
	Session sessionRequestForm
}

type sessionRequestForm struct {
	Identification string
	Password       string
}

type responseForm struct {
	Session sessionResponseForm `json:"session"`
}

type sessionResponseForm struct {
	Token string `json:"token"`
}

type errorForm struct {
	Error string `json:"error"`
}

type registrationRequestForm struct {
	Identification string
	Password       string
	Confirmation   string
}

func LoginHandler(c *echo.Context) error {
	var reqform requestForm
	err := c.Bind(&reqform)
	if err != nil {
		return err
	}

	//grab the username and password from the form
	username, password := reqform.Session.Identification, reqform.Session.Password

	user, err := models.Login(username, password)

	if err != nil {
		//var errform error_form
		//errform.Error = "invalid credentials"
		c.Response().Header().Set("Content-Type", "application/json; charset=utf-8")
		ef := errorForm{Error: "Invalid Credentials"}
		return c.JSON(http.StatusUnauthorized, ef)
	}

	if user != nil {
		var respform responseForm
		token := generateToken()

		var usersession models.UserSession
		usersession.User = *user
		usersession.Token = token
		usersession.Expiration = time.Now().Add(90 * time.Minute)
		server.Database.C("sessions").Insert(usersession)

		respform.Session.Token = token
		return c.JSON(http.StatusOK, respform)
	}
	return c.String(http.StatusInternalServerError, "Somehow reached here")
}

func generateToken() string {
	rb := make([]byte, 32)
	_, err := rand.Read(rb)

	if err != nil {
		panic(err)
	}

	token := base64.URLEncoding.EncodeToString(rb)
	return token
}

func LogoutHandler(c *echo.Context) error {
	clientToken := c.Request().Header.Get("Authorization")
	sessionCollection := server.Database.C("sessions")

	err := sessionCollection.Remove(bson.M{"token": clientToken})

	if err != nil {
		return err
	}
	return nil
}

func RegisterHandler(c *echo.Context) error {
	var regform registrationRequestForm
	err := c.Bind(&regform)
	if err != nil {
		return err
	}

	//grab username, password, and password confirmation from object
	username, password, confirmation := regform.Identification, regform.Password, regform.Confirmation

	count, err := server.Database.C("users").Find(bson.M{"username": username}).Count()
	if count != 0 {
		return c.JSON(http.StatusConflict, errorForm{Error: "Username already exists"})
	}

	if password != confirmation {
		return c.JSON(http.StatusBadRequest, errorForm{Error: "Password and confirmation must match"})
	}

	newuser := &models.User{
		Username: username,
		ID:       bson.NewObjectId(),
	}
	newuser.SetPassword(password)

	err = server.Database.C("users").Insert(newuser)
	if err != nil {
		return err
	}
	return c.String(http.StatusOK, "User registered")
}
