package models

import (
	"gitlab.mitre.org/intervention-engine/fhir/models"
	"gopkg.in/mgo.v2/bson"
)

type Fact struct {
	Id                    string                   `json:"-" bson:"_id"`
	TargetID              string                   `json:"targetid" bson:"targetid"`
	StartDate             models.FHIRDateTime      `json:"startdate" bson:"startdate"`
	EndDate               models.FHIRDateTime      `json:"enddate" bson:"enddate"`
	BirthDate             models.FHIRDateTime      `json:"birthdate" bson:"birthdate"`
	Codes                 []models.CodeableConcept `json:"codes" bson:"codes"`
	ResultQuantity        models.Quantity          `json:"resultquantity" bson:"resultquantity"`
	ResultCodeableConcept models.CodeableConcept   `json:"resultcodeableconcept" bson:"resultcodeableconcept"`
	PatientID             string                   `json:"patientid" bson:"patientid"`
	Type                  string                   `json:"type" bson:"type"`
	Gender                string                   `json:"gender" bson:"gender"`
}

func FactFromPatient(p *models.Patient) Fact {
	f := Fact{}
	f.Type = "Patient"
	f.BirthDate = p.BirthDate
	f.PatientID = p.Id
	f.TargetID = p.Id
	f.Gender = p.Gender.Coding[0].Code
	i := bson.NewObjectId()
	f.Id = i.Hex()
	return f
}

func FactFromCondition(c *models.Condition) Fact {
	f := Fact{}
	f.Type = "Condition"
	f.StartDate = c.OnsetDate
	f.EndDate = c.AbatementDate
	f.Codes = []models.CodeableConcept{c.Code}
	f.PatientID = c.Subject.ReferencedID
	f.TargetID = c.Id
	i := bson.NewObjectId()
	f.Id = i.Hex()
	return f
}

func FactFromEncounter(e *models.Encounter) Fact {
	f := Fact{}
	f.Type = "Encounter"
	f.StartDate = e.Period.Start
	f.EndDate = e.Period.End
	f.Codes = e.Type
	f.PatientID = e.Subject.ReferencedID
	f.TargetID = e.Id
	i := bson.NewObjectId()
	f.Id = i.Hex()
	return f
}

func FactFromObservation(o *models.Observation) Fact {
	f := Fact{}
	f.Type = "Observation"
	f.StartDate = o.AppliesPeriod.Start
	f.EndDate = o.AppliesPeriod.End
	f.ResultQuantity = o.ValueQuantity
	f.ResultCodeableConcept = o.ValueCodeableConcept
	f.Codes = []models.CodeableConcept{o.Name}
	f.PatientID = o.Subject.ReferencedID
	f.TargetID = o.Id
	i := bson.NewObjectId()
	f.Id = i.Hex()
	return f
}

func CreatePersonPipeline(q *models.Query) []bson.M {
	pipeline := startPipeline(q)

	pipeline = append(pipeline, bson.M{"$group": bson.M{"_id": nil, "total": bson.M{"$sum": 1}}})
	return pipeline
}

func CreateConditionPipeline(q *models.Query) []bson.M {
	pipeline := startPipeline(q)

	pipeline = append(pipeline, bson.M{"$unwind": "$entries"})
	pipeline = append(pipeline, bson.M{"$match": bson.M{"entries.type": "Condition"}})
	pipeline = append(pipeline, bson.M{"$group": bson.M{"_id": nil, "total": bson.M{"$sum": 1}}})
	return pipeline
}

func CreateEncounterPipeline(q *models.Query) []bson.M {
	pipeline := startPipeline(q)

	pipeline = append(pipeline, bson.M{"$unwind": "$entries"})
	pipeline = append(pipeline, bson.M{"$match": bson.M{"entries.type": "Encounter"}})
	pipeline = append(pipeline, bson.M{"$group": bson.M{"_id": nil, "total": bson.M{"$sum": 1}}})
	return pipeline
}

func startPipeline(q *models.Query) []bson.M {
	pipeline := []bson.M{{"$group": bson.M{"_id": "$patientid", "gender": bson.M{"$max": "$gender"}, "birthdate": bson.M{"$max": "$birthdate"}, "entries": bson.M{"$push": bson.M{"startdate": "$startdate", "enddate": "$enddate", "codes": "$codes", "type": "$type"}}}}}
	for _, extension := range q.Parameter {
		switch extension.Url {
		case "http://interventionengine.org/patientgender":
			pipeline = append(pipeline, bson.M{"$match": bson.M{"gender": extension.ValueString}})
		case "http://interventionengine.org/patientage":

			pipeline = append(pipeline, bson.M{"$match": bson.M{"birthdate": extension.ValueString}})
		case "http://interventionengine.org/conditioncode":
			// Hack for now assuming that all codable concepts contain a single code
			conditionCode := extension.ValueCodeableConcept.Coding[0].Code
			conditionSystem := extension.ValueCodeableConcept.Coding[0].System
			pipeline = append(pipeline, bson.M{"$match": bson.M{"entries.type": "Condition", "entries.codes.coding.code": conditionCode, "entries.codes.coding.system": conditionSystem}})
		}
	}

	return pipeline
}
