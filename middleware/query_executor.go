package middleware

import (
	"log"
	"net/http"

	"github.com/gorilla/context"
	fhirmodels "github.com/intervention-engine/fhir/models"
	"github.com/intervention-engine/fhir/server"
	"github.com/intervention-engine/ie/models"
	"gopkg.in/mgo.v2/bson"
)

func QueryExecutionHandler(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	next(rw, r)

	resourceType := context.Get(r, "Resource")
	actionType := context.Get(r, "Action")
	if resourceType == "Query" && actionType == "create" {
		query := context.Get(r, "Query").(*fhirmodels.Query)
		go QueryRunner(query)
	}
}

func QueryRunner(query *fhirmodels.Query) {
	pipeline := models.NewPipeline(query)
	qr, err := pipeline.ExecuteCount(server.Database)
	result := fhirmodels.QueryResponseComponent{}
	if err != nil {
		result.Outcome = "error"
		log.Println(err.Error())
	} else {
		result.Outcome = "ok"
		result.Total = float64(qr.Total)
	}
	query.Response = &result
	c := server.Database.C("querys")
	c.Update(bson.M{"_id": query.Id}, query)
}
