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

	query := codecollection.Find(bson.M{
		"codeSystem": f.CodeSystem,
		"$or": []interface{}{
			bson.M{"code": bson.RegEx{Pattern: ".*" + f.Query + ".*", Options: "i"}},
			bson.M{"name": bson.RegEx{Pattern: ".*" + f.Query + ".*", Options: "i"}},
		}}).Limit(f.ResultLimit)

	if err := query.All(&result); err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.Header("Access-Control-Allow-Origin", "*")
	c.JSON(http.StatusOK, result)
}
