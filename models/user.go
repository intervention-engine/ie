package models

import (
	"github.com/intervention-engine/fhir/server"
	"golang.org/x/crypto/bcrypt"
	"gopkg.in/mgo.v2/bson"
	"time"
)

type User struct {
	ID       bson.ObjectId `bson:"_id,omitempty"`
	Username string
	Password []byte
}

type UserSession struct {
	User       User
	Token      string
	Expiration time.Time
}

func (u *User) SetPassword(password string) {
	hashedpass, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		panic(err)
	}

	u.Password = hashedpass
}

func Login(username, password string) (u *User, err error) {
	err = server.Database.C("users").Find(bson.M{"username": username}).One(&u)
	if err != nil {
		return
	}

	err = bcrypt.CompareHashAndPassword(u.Password, []byte(password))
	if err != nil {
		u = nil
	}

	return u, err
}
