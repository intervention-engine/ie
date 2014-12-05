package controllers

import (
	"encoding/json"

	"github.com/gorilla/mux"
	fhirmodels "gitlab.mitre.org/intervention-engine/fhir/models"
	"gitlab.mitre.org/intervention-engine/fhir/server"
	"gitlab.mitre.org/intervention-engine/ie/middleware"
	"gitlab.mitre.org/intervention-engine/ie/models"
	"gopkg.in/mgo.v2/bson"
	"net/http"
)

func ConditionTotalHandler(rw http.ResponseWriter, r *http.Request) {
	query := loadQuery(rw, r)
	pipeline := models.CreateConditionPipeline(query)
	factCollection := server.Database.C("facts")
	qr := &middleware.QueryResult{}
	factCollection.Pipe(pipeline).One(qr)
	json.NewEncoder(rw).Encode(qr)
}

func EncounterTotalHandler(rw http.ResponseWriter, r *http.Request) {
	query := loadQuery(rw, r)
	pipeline := models.CreateEncounterPipeline(query)
	factCollection := server.Database.C("facts")
	qr := &middleware.QueryResult{}
	factCollection.Pipe(pipeline).One(qr)
	json.NewEncoder(rw).Encode(qr)
}

func loadQuery(rw http.ResponseWriter, r *http.Request) *fhirmodels.Query {
	var id bson.ObjectId

	idString := mux.Vars(r)["id"]
	if bson.IsObjectIdHex(idString) {
		id = bson.ObjectIdHex(idString)
	} else {
		http.Error(rw, "Invalid id", http.StatusBadRequest)
	}

	c := server.Database.C("querys")

	query := &fhirmodels.Query{}
	err := c.Find(bson.M{"_id": id.Hex()}).One(query)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return nil
	}
	return query
}
