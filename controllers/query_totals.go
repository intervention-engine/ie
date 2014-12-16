package controllers

import (
	"encoding/json"

	"net/http"

	"github.com/gorilla/mux"
	fhirmodels "github.com/intervention-engine/fhir/models"
	"github.com/intervention-engine/fhir/server"
	"github.com/intervention-engine/ie/middleware"
	"github.com/intervention-engine/ie/models"
	"gopkg.in/mgo.v2/bson"
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
