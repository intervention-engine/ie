package mongo

import (
	"github.com/intervention-engine/fhir/models"
	"github.com/intervention-engine/ie/app"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type CurrentActiveElements struct {
	Conditions  []*app.ActiveElement
	Medications []*app.ActiveElement
	Allergies   []*app.ActiveElement
}

func getActiveForPatient(db *mgo.Database, id string) CurrentActiveElements {
	c := db.C("conditions")
	a := db.C("allergyintolerances")
	m := db.C("medicationstatements")

	var am []*models.AllergyIntolerance
	var cm []*models.Condition
	var mm []*models.MedicationStatement

	activeQuery(a, id, "status", "active", "confirmed").All(&am)
	activeQuery(c, id, "clinicalStatus", "active", "relapsed").All(&cm)
	activeQuery(m, id, "status", "active").All(&mm)

	active := CurrentActiveElements{Conditions: conditionsToActive(cm),
		Allergies:   allergiesToActive(am),
		Medications: medsToActive(mm)}
	return active
}

func medsToActive(meds []*models.MedicationStatement) []*app.ActiveElement {
	active := make([]*app.ActiveElement, len(meds))
	for i, m := range meds {
		active[i] = &app.ActiveElement{Name: &m.MedicationCodeableConcept.Text, StartDate: &m.EffectivePeriod.Start.Time}
	}
	return active
}

func allergiesToActive(all []*models.AllergyIntolerance) []*app.ActiveElement {
	active := make([]*app.ActiveElement, len(all))
	for i, a := range all {
		active[i] = &app.ActiveElement{Name: &a.Substance.Text, StartDate: &a.RecordedDate.Time}
	}
	return active
}

func conditionsToActive(con []*models.Condition) []*app.ActiveElement {
	active := make([]*app.ActiveElement, len(con))
	for i, c := range con {
		active[i] = &app.ActiveElement{Name: &c.Code.Text, StartDate: &c.OnsetDateTime.Time}
	}

	return active
}

func activeQuery(c *mgo.Collection, patientID string, attribute string, status ...string) *mgo.Query {
	statuses := make([]bson.M, len(status))

	for i, s := range status {
		statuses[i] = bson.M{attribute: s}
	}

	return c.Find(bson.M{"patient.reference": "Patient/" + patientID, "$or": statuses})
}
