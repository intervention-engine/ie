package db

import (
	"os"

	mgo "gopkg.in/mgo.v2"
)

var db *mgo.Database

// SetupDBConnection gets a Mongo session and a database reference
// Session needs to be closed by caller
func SetupDBConnection(db string) (*mgo.Session, *mgo.Database) {
	mongoURL := os.Getenv("MONGO_URL")
	if mongoURL == "" {
		mongoURL = "mongodb://localhost:27017"
	}

	s, err := mgo.Dial(mongoURL)

	if err != nil {
		panic(err)
	}

	return s, connectDB(s, db)
}

func connectDB(s *mgo.Session, dbName string) *mgo.Database {
	if db == nil {
		db = s.DB(dbName)
	}

	return db
}
