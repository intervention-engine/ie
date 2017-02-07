package db

import (
	"os"

	mgo "gopkg.in/mgo.v2"
)

var db *mgo.Database

func init() {
	mongoURL := os.Getenv("MONGO_URL")
	if mongoURL == "" {
		mongoURL = "mongodb://localhost:27017"
	}

	session, err := mgo.Dial(mongoURL)

	if err != nil {
		panic(err)
	}

	db = session.DB("fhir")
}

// GetDB db reference
func GetDB() *mgo.Database {
	return db
}
