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
	if resourceType == "Group" && actionType == "create" {
		group := context.Get(r, "Group").(*fhirmodels.Group)
		go GroupRunner(group)
	}
}

func GroupRunner(group *fhirmodels.Group) {
	pipeline := models.NewPipeline(group)
	qr, err := pipeline.ExecuteCount(server.Database)
	if err != nil {
		log.Println(err.Error())
	} else {
		total := uint32(qr.Total)
		group.Quantity = &total
		c := server.Database.C("groups")
		c.Update(bson.M{"_id": group.Id}, group)
	}

}
