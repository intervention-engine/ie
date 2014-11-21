package middleware

import (
	"log"
	"net/http"

	"github.com/gorilla/context"
	fhirmodels "gitlab.mitre.org/intervention-engine/fhir/models"
	"gitlab.mitre.org/intervention-engine/fhir/server"
	"gitlab.mitre.org/intervention-engine/ie/models"
	"gopkg.in/mgo.v2/bson"
)

type QueryResult struct {
	Total int `json:"total", bson:"total"`
}

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
	pipeline := models.CreateFactPipeline(query)
	factCollection := server.Database.C("facts")
	qr := &QueryResult{}
	err := factCollection.Pipe(pipeline).One(qr)
	result := fhirmodels.QueryResponseComponent{}
	if err != nil {
		result.Outcome = "error"
		log.Println(err.Error())
	} else {
		result.Outcome = "ok"
		result.Total = float64(qr.Total)
	}
	query.Response = result
	c := server.Database.C("querys")
	c.Update(bson.M{"_id": query.Id}, query)
}
