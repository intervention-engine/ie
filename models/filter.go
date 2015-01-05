package models

import (
	fhirmodels "github.com/intervention-engine/fhir/models"
	"github.com/intervention-engine/fhir/server"
	"gopkg.in/mgo.v2/bson"
)

type Filter struct {
	Id          string           `json:"id" bson:"_id"`
	Name        string           `json:"name" bson:"name"`
	Description string           `json:"description" bson:"description"`
	Query       fhirmodels.Query `json:"query" bson:"query"`
	Panes       []Pane           `json:"panes" bson:"panes"`
	Url         string           `json:"url" bson:"url"`
}

func (f *Filter) CreateQuery() (string, error) {
	c := server.Database.C("querys")
	query := f.Query
	i := bson.NewObjectId()
	hex := i.Hex()
	query.Id = hex
	err := c.Insert(query)
	return hex, err
}

type Pane struct {
	id    string      `json:"-" bson:"_id"`
	Icon  string      `json:"icon" bson:"icon"`
	Items []EmberItem `json:"items" bson:"items"`
}

type EmberItem struct {
	id         string               `json:"-" bson:"_id"`
	Parameter  fhirmodels.Extension `json:"parameter" bson:"parameter"`
	Active     *bool                `json:"active" bson:"active"`
	FilterType string               `json:"filtertype" bson:"filtertype"`
	Template   string               `json:"template" bson:"template"`
}
