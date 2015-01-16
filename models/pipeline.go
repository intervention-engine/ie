package models

import (
	"github.com/intervention-engine/fhir/models"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"time"
)

type Pipeline struct {
	MongoPipeline []bson.M
}

type QueryResult struct {
	Total int `json:"total", bson:"total"`
}

type PipelineProducer func(q *models.Query) (p Pipeline)

func NewPipeline(q *models.Query) Pipeline {
	pipeline := Pipeline{}
	pipeline.MongoPipeline = []bson.M{{"$group": bson.M{"_id": "$patientid", "gender": bson.M{"$max": "$gender"}, "birthdate": bson.M{"$max": "$birthdate"}, "entries": bson.M{"$push": bson.M{"startdate": "$startdate", "enddate": "$enddate", "codes": "$codes", "type": "$type"}}}}}
	for _, extension := range q.Parameter {
		switch extension.Url {
		case "http://interventionengine.org/patientgender":
			pipeline.MongoPipeline = append(pipeline.MongoPipeline, bson.M{"$match": bson.M{"gender": extension.ValueString}})
		case "http://interventionengine.org/patientage":
			lowAgeDate, highAgeDate := ageRangeToTime(extension.ValueRange)
			pipeline.MongoPipeline = append(pipeline.MongoPipeline, bson.M{"$match": bson.M{"birthdate.time": bson.M{"$gte": highAgeDate, "$lte": lowAgeDate}}})
		case "http://interventionengine.org/conditioncode":
			// Hack for now assuming that all codable concepts contain a single code
			conditionCode := extension.ValueCodeableConcept.Coding[0].Code
			conditionSystem := extension.ValueCodeableConcept.Coding[0].System
			pipeline.MongoPipeline = append(pipeline.MongoPipeline, bson.M{"$match": bson.M{"entries.type": "Condition", "entries.codes.coding.code": conditionCode, "entries.codes.coding.system": conditionSystem}})
		}
	}

	return pipeline
}

func NewPersonPipeline(q *models.Query) Pipeline {
	pipeline := NewPipeline(q)

	pipeline.MongoPipeline = append(pipeline.MongoPipeline, bson.M{"$group": bson.M{"_id": nil, "total": bson.M{"$sum": 1}}})
	return pipeline
}

func NewConditionPipeline(q *models.Query) Pipeline {
	pipeline := NewPipeline(q)

	pipeline.MongoPipeline = append(pipeline.MongoPipeline, bson.M{"$unwind": "$entries"})
	pipeline.MongoPipeline = append(pipeline.MongoPipeline, bson.M{"$match": bson.M{"entries.type": "Condition"}})
	pipeline.MongoPipeline = append(pipeline.MongoPipeline, bson.M{"$group": bson.M{"_id": "entries.codes.coding.code", "total": bson.M{"$sum": 1}}})
	pipeline.MongoPipeline = append(pipeline.MongoPipeline, bson.M{"$group": bson.M{"_id": nil, "total": bson.M{"$sum": 1}}})
	return pipeline
}

func NewEncounterPipeline(q *models.Query) Pipeline {
	pipeline := NewPipeline(q)

	pipeline.MongoPipeline = append(pipeline.MongoPipeline, bson.M{"$unwind": "$entries"})
	pipeline.MongoPipeline = append(pipeline.MongoPipeline, bson.M{"$match": bson.M{"entries.type": "Encounter"}})
	pipeline.MongoPipeline = append(pipeline.MongoPipeline, bson.M{"$group": bson.M{"_id": nil, "total": bson.M{"$sum": 1}}})
	return pipeline
}

func (p *Pipeline) Execute(db *mgo.Database) (QueryResult, error) {
	factCollection := db.C("facts")
	qr := QueryResult{}
	err := factCollection.Pipe(p.MongoPipeline).One(&qr)
	return qr, err
}

func ageRangeToTime(ageRange models.Range) (lowAgeDate, highAgeDate time.Time) {
	lowAgeDate = time.Date(time.Now().Year()-int(ageRange.Low.Value), time.Now().Month(), time.Now().Day(), 0, 0, 0, 0, time.UTC)
	highAgeDate = time.Date(time.Now().Year()-int(ageRange.High.Value), time.Now().Month(), time.Now().Day(), 0, 0, 0, 0, time.UTC)
	return
}
