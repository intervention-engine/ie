package controllers

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/intervention-engine/fhir/server"
	"gopkg.in/mgo.v2/bson"
)

func NotificationCountHandler(c *gin.Context) {
	pipe := server.Database.C("communicationrequests").Pipe([]bson.M{{"$group": bson.M{"_id": "$subject.referenceid", "count": bson.M{"$sum": 1}}}})
	var results []NotificationCountResult
	if err := pipe.All(&results); err != nil {
		log.Printf("Error getting notification count: %#v", err)
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, results)
}

type NotificationCountResult struct {
	Patient string  `bson:"_id" json:"patient"`
	Count   float64 `bson:"count" json:"count"`
}
