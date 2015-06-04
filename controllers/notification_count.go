package controllers

import (
	"encoding/json"
	"github.com/intervention-engine/fhir/server"
	"gopkg.in/mgo.v2/bson"
	"log"
	"net/http"
)

func NotificationCountHandler(rw http.ResponseWriter, r *http.Request) {
	pipe := server.Database.C("communicationrequests").Pipe([]bson.M{{"$group": bson.M{"_id": "$subject.referenceid", "count": bson.M{"$sum": 1}}}})
	var results []NotificationCountResult
	err := pipe.All(&results)
	if err != nil {
		log.Printf("Error getting notification count: %#v", err)
		http.Error(rw, err.Error(), http.StatusInternalServerError)
	} else {
		json.NewEncoder(rw).Encode(results)
	}
}

type NotificationCountResult struct {
	Patient string  `bson:"_id" json:"patient"`
	Count   float64 `bson:"count" json:"count"`
}
