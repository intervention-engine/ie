package controllers

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/context"
	"github.com/gorilla/mux"
	"github.com/intervention-engine/fhir/server"
	"github.com/intervention-engine/ie/middleware"
	"github.com/intervention-engine/ie/models"
	"gopkg.in/mgo.v2/bson"
)

func FilterIndexHandler(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	var result []models.Filter
	c := server.Database.C("filters")
	iter := c.Find(nil).Limit(100).Iter()
	err := iter.All(&result)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
	}

	log.Println("Setting filter search context")
	context.Set(r, "Filter", result)
	context.Set(r, "Resource", "Filter")
	context.Set(r, "Action", "search")

	rw.Header().Set("Content-Type", "application/json; charset=utf-8")
	rw.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(rw).Encode(result)
}

func FilterShowHandler(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {

	var id bson.ObjectId

	idString := mux.Vars(r)["id"]
	if bson.IsObjectIdHex(idString) {
		id = bson.ObjectIdHex(idString)
	} else {
		http.Error(rw, "Invalid id", http.StatusBadRequest)
	}

	c := server.Database.C("filters")

	result := models.Filter{}
	err := c.Find(bson.M{"_id": id.Hex()}).One(&result)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Println("Setting filter read context")
	context.Set(r, "Filter", result)
	context.Set(r, "Resource", "Filter")
	context.Set(r, "Action", "read")

	rw.Header().Set("Content-Type", "application/json; charset=utf-8")
	rw.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(rw).Encode(result)
}

func FilterCreateHandler(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	decoder := json.NewDecoder(r.Body)
	filter := &models.Filter{}
	err := decoder.Decode(filter)
	if err != nil {
		log.Println("Error on decoding filter")
		http.Error(rw, err.Error(), http.StatusInternalServerError)
	}

	host, err := os.Hostname()
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
	}

	c := server.Database.C("filters")
	i := bson.NewObjectId()
	filter.Id = i.Hex()
	queryId, err := filter.CreateQuery(server.Database)
	if err == nil {
		filter.Url = "http://" + host + ":3001/Query/" + queryId
		query := filter.Query
		query.Id = queryId
		go middleware.QueryRunner(&query)
	}

	err = c.Insert(filter)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
	}

	log.Println("Setting filter create context")
	context.Set(r, "Filter", filter)
	context.Set(r, "Resource", "Filter")
	context.Set(r, "Action", "create")

	rw.Header().Add("Location", "http://"+host+":3001/Filter/"+i.Hex())
	json.NewEncoder(rw).Encode(filter)

}

func FilterUpdateHandler(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {

	var id bson.ObjectId

	idString := mux.Vars(r)["id"]
	if bson.IsObjectIdHex(idString) {
		id = bson.ObjectIdHex(idString)
	} else {
		http.Error(rw, "Invalid id", http.StatusBadRequest)
	}

	decoder := json.NewDecoder(r.Body)
	filter := &models.Filter{}
	err := decoder.Decode(filter)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
	}

	c := server.Database.C("filters")
	filter.Id = id.Hex()
	err = c.Update(bson.M{"_id": id.Hex()}, filter)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
	}

	log.Println("Setting filter update context")
	context.Set(r, "Filter", filter)
	context.Set(r, "Resource", "Filter")
	context.Set(r, "Action", "update")
}

func FilterDeleteHandler(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	var id bson.ObjectId

	idString := mux.Vars(r)["id"]
	if bson.IsObjectIdHex(idString) {
		id = bson.ObjectIdHex(idString)
	} else {
		http.Error(rw, "Invalid id", http.StatusBadRequest)
	}

	c := server.Database.C("filters")

	err := c.Remove(bson.M{"_id": id.Hex()})
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Println("Setting filter delete context")
	context.Set(r, "Filter", id.Hex())
	context.Set(r, "Resource", "Filter")
	context.Set(r, "Action", "delete")
}
