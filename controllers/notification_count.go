package controllers

import (
	"log"
	"net/http"

	"github.com/intervention-engine/fhir/server"
	"github.com/labstack/echo"
	"gopkg.in/mgo.v2/bson"
)

func NotificationCountHandler(c *echo.Context) error {
	pipe := server.Database.C("communicationrequests").Pipe([]bson.M{{"$group": bson.M{"_id": "$subject.referenceid", "count": bson.M{"$sum": 1}}}})
	var results []NotificationCountResult
	err := pipe.All(&results)
	if err != nil {
		log.Printf("Error getting notification count: %#v", err)
		return err
	}
	return c.JSON(http.StatusOK, results)
}

type NotificationCountResult struct {
	Patient string  `bson:"_id" json:"patient"`
	Count   float64 `bson:"count" json:"count"`
}
