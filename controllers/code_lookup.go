package controllers

import (
	"net/http"

	"github.com/intervention-engine/fhir/server"
	"github.com/intervention-engine/ie/utilities"
	"github.com/labstack/echo"
	"gopkg.in/mgo.v2/bson"
)

type codeRequestForm struct {
	CodeSystem  string `json:"codesystem"`
	Query       string `json:"query"`
	ResultLimit int    `json:"limit"`
}

func CodeLookup(c *echo.Context) error {
	var f codeRequestForm

	err := c.Bind(&f)
	if err != nil {
		return err
	}

	codecollection := server.Database.C("codelookup")

	result := []utilities.CodeEntry{}

	query := codecollection.Find(bson.M{
		"codeSystem": f.CodeSystem,
		"$or": []interface{}{
			bson.M{"code": bson.RegEx{Pattern: ".*" + f.Query + ".*", Options: "i"}},
			bson.M{"name": bson.RegEx{Pattern: ".*" + f.Query + ".*", Options: "i"}},
		}}).Limit(f.ResultLimit)

	err = query.All(&result)
	if err != nil {
		return err
	}

	c.Response().Header().Set("Access-Control-Allow-Origin", "*")
	return c.JSON(http.StatusOK, result)
}
