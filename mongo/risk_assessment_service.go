package mongo

import (
	"time"

	"gopkg.in/mgo.v2/bson"

	"github.com/intervention-engine/fhir/models"
	"github.com/intervention-engine/ie/app"
)

// RiskAssessmentService List and transform FHIR RiskAssessments
type RiskAssessmentService struct {
	Service
}

// RiskAssessments List risk assessments for a patient
func (ras *RiskAssessmentService) RiskAssessments(patientID string, serviceID string, start time.Time, end time.Time) ([]*app.RiskAssessment, error) {
	defer ras.S.Close()
	query := bson.M{"subject.referenceid": patientID, "method.coding.code": serviceID}

	var fhirResults []*models.RiskAssessment
	err := ras.C.Find(query).All(&fhirResults)
	return newAssessments(fhirResults), err
}

// RiskAssessment Find a single risk assessment by ID
func (ras *RiskAssessmentService) RiskAssessment(id string) (*app.RiskAssessment, error) {
	defer ras.S.Close()
	var fhirResult *models.RiskAssessment
	err := ras.C.FindId(id).One(&fhirResult)

	return newAssessment(fhirResult), err
}

func newAssessments(r []*models.RiskAssessment) []*app.RiskAssessment {
	ra := make([]*app.RiskAssessment, len(r))

	for i := 0; i < len(r); i++ {
		ra[i] = newAssessment(r[i])
	}

	return ra
}

func newAssessment(r *models.RiskAssessment) *app.RiskAssessment {
	if r == nil {
		return nil
	}
	ra := &app.RiskAssessment{Date: &r.Date.Time}
	if r.Id != "" {
		ra.ID = &r.Id
	}
	if (r.Method != nil) && (len(r.Method.Coding) > 0) {
		ra.RiskServiceID = &r.Method.Coding[0].Code
	}
	if (r.Prediction != nil) && (len(r.Prediction) > 0) {
		ra.Value = r.Prediction[0].ProbabilityDecimal
	}
	return ra
}
