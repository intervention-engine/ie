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

	query := bson.M{"subject.referenceid": patientID, "method.coding.code": serviceID}

	var fhirResults []*models.RiskAssessment
	err := ras.C.Find(query).All(&fhirResults)
	return ras.newAssessments(fhirResults), err
}

// RiskAssessment Find a single risk assessment by ID
func (ras *RiskAssessmentService) RiskAssessment(id string) (*app.RiskAssessment, error) {
	var fhirResult *models.RiskAssessment
	err := ras.C.FindId(id).One(&fhirResult)

	return ras.newAssessment(fhirResult), err
}

func (ras *RiskAssessmentService) newAssessments(r []*models.RiskAssessment) []*app.RiskAssessment {
	ra := make([]*app.RiskAssessment, len(r))

	for i := 0; i < len(r); i++ {
		ra[i] = ras.newAssessment(r[i])
	}

	return ra
}

func (ras *RiskAssessmentService) newAssessment(r *models.RiskAssessment) *app.RiskAssessment {
	return &app.RiskAssessment{
		ID:            &r.Id,
		Date:          &r.Date.Time,
		RiskServiceID: &r.Method.Coding[0].Code,
		Value:         r.Prediction[0].ProbabilityDecimal,
	}
}
