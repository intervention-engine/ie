package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/intervention-engine/fhir/server"
	"github.com/intervention-engine/ie/utilities"
	"gopkg.in/mgo.v2/bson"
)

type codeRequestForm struct {
	CodeSystem  string `json:"codesystem"`
	Query       string `json:"query"`
	ResultLimit int    `json:"limit"`
}

func CodeLookup(c *gin.Context) {
	var f codeRequestForm

	if err := c.Bind(&f); err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	codecollection := server.Database.C("codelookup")

	result := []utilities.CodeEntry{}

	codeQuery := bson.M{
		"codeSystem": f.CodeSystem,
		"$or": []interface{}{
			bson.M{"code": bson.RegEx{Pattern: ".*" + f.Query + ".*", Options: "i"}},
			bson.M{"name": bson.RegEx{Pattern: ".*" + f.Query + ".*", Options: "i"}},
		}}

	pipeline := []bson.M{
		{"$match": codeQuery},
		{"$lookup": bson.M{
			"from":         "conditions",
			"localField":   "code",
			"foreignField": "code.coding.code",
			"as":           "_conditions",
		}},
		{"$project": bson.M{
			"_id":        1,
			"code":       1,
			"codeSystem": 1,
			"name":       1,
			"count": bson.M{
				"$size": "$_conditions",
			},
		}},
		{"$sort": bson.D{
			{"count", -1},
			{"code", 1},
		}},
		{"$limit": f.ResultLimit},
	}

	if err := codecollection.Pipe(pipeline).All(&result); err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.Header("Access-Control-Allow-Origin", "*")
	c.JSON(http.StatusOK, result)
}
