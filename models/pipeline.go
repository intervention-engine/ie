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

type MatchingStage struct {
	AndStatements bson.M
	OrStatements  []bson.M
}

func NewMS() *MatchingStage {
	ms := &MatchingStage{}
	ms.AndStatements = bson.M{}
	ms.OrStatements = []bson.M{}
	return ms
}

func (m *MatchingStage) AddAndStatement(key string, value interface{}) {
	m.AndStatements[key] = value
}

func (m *MatchingStage) AddOrStatement(statement bson.M) {
	m.OrStatements = append(m.OrStatements, statement)
}

func (m *MatchingStage) AddType(typeString string) {
	m.AndStatements["entries.type"] = typeString
}

func (m *MatchingStage) AddCodableConecpt(cc models.CodeableConcept) {
	if len(cc.Coding) == 1 {
		code := cc.Coding[0].Code
		system := cc.Coding[0].System
		m.AddAndStatement("entries.codes.coding.code", code)
		m.AddAndStatement("entries.codes.coding.system", system)
	} else {
		for _, coding := range cc.Coding {
			code := coding.Code
			system := coding.System
			m.AddOrStatement(bson.M{"entries.codes.coding.code": code, "entries.codes.coding.system": system})
		}
	}
}

func (m *MatchingStage) AddAgeRange(ageRange models.Range) {
	rangeQuery := bson.M{}
	if ageRange.Low != nil && ageRange.Low.Value != nil {
		lowAgeDate := time.Date(time.Now().Year()-int(*ageRange.Low.Value), time.Now().Month(), time.Now().Day(), 0, 0, 0, 0, time.UTC)
		rangeQuery["$lte"] = lowAgeDate
	}

	if ageRange.High != nil && ageRange.High.Value != nil {
		highAgeDate := time.Date(time.Now().Year()-int(*ageRange.High.Value), time.Now().Month(), time.Now().Day(), 0, 0, 0, 0, time.UTC)
		rangeQuery["$gte"] = highAgeDate
	}
	m.AddAndStatement("birthdate.time", rangeQuery)
}

func (m *MatchingStage) AddValueCheck(e models.Extension) {
	if e.ValueInteger != nil {
		m.AddAndStatement("entries.resultquantity.value", float64(*e.ValueInteger))
	}
	if e.ValueRange != nil {
		if e.ValueRange.High.Value != nil || e.ValueRange.Low.Value != nil {
			rangeQuery := bson.M{}
			if e.ValueRange.High.Value != nil {
				rangeQuery["$lte"] = e.ValueRange.High.Value
			}
			if e.ValueRange.Low.Value != nil {
				rangeQuery["$gte"] = e.ValueRange.Low.Value
			}
			m.AddAndStatement("entries.resultquantity.value", rangeQuery)
		}
	}
}

func (m *MatchingStage) ToBSON() bson.M {
	if len(m.OrStatements) > 1 {
		m.AndStatements["$or"] = m.OrStatements
	}
	return bson.M{"$match": m.AndStatements}
}

type QueryResult struct {
	Total int `json:"total", bson:"total"`
}

type QueryPatientList struct {
	PatientIds []string `json:"patientids", bson:"patientids"`
}

type PipelineProducer func(q *models.Query) (p Pipeline)

func NewPipeline(q *models.Query) Pipeline {
	pipeline := Pipeline{}
	pipeline.MongoPipeline = []bson.M{{"$group": bson.M{"_id": "$patientid", "gender": bson.M{"$max": "$gender"}, "birthdate": bson.M{"$max": "$birthdate"}, "entries": bson.M{"$push": bson.M{"startdate": "$startdate", "enddate": "$enddate", "codes": "$codes", "type": "$type", "resultquantity": "$resultquantity"}}}}}
	for _, extension := range q.Parameter {
		ms := NewMS()
		switch extension.Url {
		case "http://interventionengine.org/patientgender":
			ms.AddAndStatement("gender", extension.ValueString)
		case "http://interventionengine.org/patientage":
			ms.AddAgeRange(*extension.ValueRange)
		case "http://interventionengine.org/conditioncode":
			ms.AddType("Condition")
			ms.AddCodableConecpt(*extension.ValueCodeableConcept)
		case "http://interventionengine.org/encountercode":
			ms.AddType("Encounter")
			ms.AddCodableConecpt(*extension.ValueCodeableConcept)
		case "http://interventionengine.org/observationcode":
			ms.AddType("Observation")
			ms.AddCodableConecpt(*extension.ValueCodeableConcept)
			ms.AddValueCheck(extension)
		}
		pipeline.MongoPipeline = append(pipeline.MongoPipeline, ms.ToBSON())
	}

	return pipeline
}

func IsRangePresent(r models.Range) bool {
	return r.High.Value != nil && r.Low.Value != nil
}

func NewConditionPipeline(q *models.Query) Pipeline {
	pipeline := NewPipeline(q)

	pipeline.MongoPipeline = append(pipeline.MongoPipeline, bson.M{"$unwind": "$entries"})
	pipeline.MongoPipeline = append(pipeline.MongoPipeline, bson.M{"$match": bson.M{"entries.type": "Condition"}})
	return pipeline
}

func NewEncounterPipeline(q *models.Query) Pipeline {
	pipeline := NewPipeline(q)

	pipeline.MongoPipeline = append(pipeline.MongoPipeline, bson.M{"$unwind": "$entries"})
	pipeline.MongoPipeline = append(pipeline.MongoPipeline, bson.M{"$match": bson.M{"entries.type": "Encounter"}})
	return pipeline
}

func (p *Pipeline) MakeCountPipeline() {
	p.MongoPipeline = append(p.MongoPipeline, bson.M{"$group": bson.M{"_id": nil, "total": bson.M{"$sum": 1}}})
}

func (p *Pipeline) MakePatientListPipeline() {
	p.MongoPipeline = append(p.MongoPipeline, bson.M{"$group": bson.M{"_id": nil, "patientids": bson.M{"$push": "$_id"}}})
}

func (p *Pipeline) ExecuteCount(db *mgo.Database) (QueryResult, error) {
	factCollection := db.C("facts")
	qr := QueryResult{}
	p.MakeCountPipeline()
	err := factCollection.Pipe(p.MongoPipeline).One(&qr)
	return qr, err
}

func (p *Pipeline) ExecutePatientList(db *mgo.Database) (QueryPatientList, error) {
	factCollection := db.C("facts")
	qpl := QueryPatientList{}
	p.MakePatientListPipeline()
	err := factCollection.Pipe(p.MongoPipeline).One(&qpl)
	return qpl, err
}
