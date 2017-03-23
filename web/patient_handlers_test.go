package web_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/intervention-engine/fhir/models"
	"github.com/intervention-engine/ie"
	"github.com/intervention-engine/ie/testutil"
	"github.com/intervention-engine/ie/web"
	"github.com/stretchr/testify/suite"
	"gopkg.in/mgo.v2/bson"
)

type patientSuite struct {
	testutil.WebSuite
	DB map[string]ie.Patient
}

func TestPatientHandlersSuite(t *testing.T) {
	m := make(map[string]ie.Patient)
	suite.Run(t, &patientSuite{DB: m})
}

func (suite *patientSuite) SetupTest() {
	for _, patient := range patients {
		suite.DB[patient.Id] = patient
	}
}

func (suite *patientSuite) SetupSuite() {
	api := suite.LoadGin()
	api.GET("/patients", web.Adapt(web.ListAllPatients, suite.withTestService()))
	api.GET("/patients/:id", web.Adapt(web.GetPatient, suite.withTestService()))
}

func generateBirthdate(birthdate string) *models.FHIRDateTime {
	date, err := time.ParseInLocation("2006-01-02", birthdate, time.Local)
	if err != nil {
		panic(err)
	}
	return &models.FHIRDateTime{date, models.Precision("date")}
}

var patients = []ie.Patient{
	{
		Patient: models.Patient{
			DomainResource: models.DomainResource{
				Resource: models.Resource{
					Id: "58938873bd90ef501e29c919",
				},
			},
			Name: []models.HumanName{
				{
					Family: []string{
						"Morgan1765",
					},
					Given: []string{
						"Amy",
					},
				},
			},
			Address: []models.Address{
				{
					Line: []string{
						"30377 Petterle Place",
					},
					City:       "Los Gatos",
					State:      "GA",
					PostalCode: "42586",
				},
			},
			Gender:    "female",
			BirthDate: generateBirthdate("1962-01-01"),
		},
		NextHuddleID: "576c9bbf8bd4a4bdc2ac2038",
		RiskAssessments: []ie.RiskAssessment{
			{
				ID:      "576c9bcf8bd4d4bdc2ac481c",
				GroupID: "576c9bcc8bd4d4bdc2ac429d",
				Value:   3,
			},
		},
	},
	{
		Patient: models.Patient{
			DomainResource: models.DomainResource{
				Resource: models.Resource{
					Id: "58c314acb367c1ff54d19e9e",
				},
			},
			Name: []models.HumanName{
				{
					Family: []string{
						"Taylor Swift",
					},
					Given: []string{
						"TayTay a.k.a Sway Sway",
					},
				},
			},
			Address: []models.Address{
				{
					Line: []string{
						"42 Wallaby Way",
					},
					City:       "Sydney",
					State:      "AUS",
					PostalCode: "xxxxx",
				},
			},
			Gender:    "female",
			BirthDate: generateBirthdate("1954-09-21"),
		},
		NextHuddleID: "576c9bbf8bd4a4bdc2ac2038",
		RiskAssessments: []ie.RiskAssessment{
			{
				ID:      "576c9bcf8bd4d4bdc2ac481c",
				GroupID: "576c9bcc8bd4d4bdc2ac429d",
				Value:   3,
			},
		},
	},
	{
		Patient: models.Patient{
			DomainResource: models.DomainResource{
				Resource: models.Resource{
					Id: "576c9bcc8bd4d4bdc2ac42b5",
				},
			},
			Name: []models.HumanName{
				{
					Family: []string{
						"Banks9849",
					},
					Given: []string{
						"Jacqueline",
					},
				},
			},
			Address: []models.Address{
				{
					Line: []string{
						"4691 7th Way",
					},
					City:       "Loyalton",
					State:      "WY",
					PostalCode: "01409",
				},
			},
			Gender:    "female",
			BirthDate: generateBirthdate("1962-01-01"),
		},
		NextHuddleID: "576c9bbf8bd4a4bdc2ac2038",
		RiskAssessments: []ie.RiskAssessment{
			{
				ID:      "576c9bcf8bd4d4bdc2ac481c",
				GroupID: "576c9bcc8bd4d4bdc2ac429d",
				Value:   3,
			},
		},
	},
}

// TODO: Should make some permanent times so we can test that, too
// When there are patients, should return a slice of patients with 200 OK
func (suite *patientSuite) TestAllPatientsFound() {
	w := suite.AssertGetRequest("/api/patients", http.StatusOK)
	var body = make(map[string][]ie.RestructedPatient)
	json.NewDecoder(w.Body).Decode(&body)
	results, ok := body["patients"]
	if !ok {
		suite.T().Errorf("body did not contain an element named patients | body: %#+v\n", w.Body.String())
		return
	}
	for _, patient := range patients {
		result := suite.findPatient(patient.Id, results)

		expName := patient.Name[0]
		name := result.Name
		suite.Assert().NotNil(result, "body did not contain patient: %s with ID: %s\n", patient.Name[0].Family[0], patient.Id)
		suite.Assert().Equal(expName.Family[0], name.Family)
		suite.Assert().Equal(expName.Given[0], name.Given)
	}
}

// If a patient with that (correct) id does not exist in the database, should
// return 404 Not Found
func (suite *patientSuite) TestGetPatientNotFound() {
	id := patients[0].Id
	delete(suite.DB, id)
	suite.AssertGetRequest("/api/patients/"+id, http.StatusNotFound)
}

// When given an incorrectly formatted id, should return 400 Bad Request
func (suite *patientSuite) TestGetPatientWithBadId() {
	suite.AssertGetRequest("/api/patients/sdf", http.StatusBadRequest)
}

// If a patient with that id exists, should return the patient and 200 OK
func (suite *patientSuite) TestGetPatientFound() {
	w := suite.AssertGetRequest("/api/patients/58938873bd90ef501e29c919", http.StatusOK)
	var body = make(map[string]ie.RestructedPatient)
	json.NewDecoder(w.Body).Decode(&body)
	result, ok := body["patient"]

	if !ok {
		suite.T().Errorf("body did not contain an element named patient | body: %#+v\n", w.Body.String())
		return
	}

	patient := patients[0]
	expName := patient.Name[0]
	name := result.Name

	suite.Assert().NotNil(result, "body did not contain patient: %s with ID: %s\n", patient.Name[0].Family[0], patient.Id)
	suite.Assert().Equal(expName.Family[0], name.Family)
	suite.Assert().Equal(expName.Given[0], name.Given)
}

// Mock Services

func (suite *patientSuite) Patient(id string) (*ie.RestructedPatient, error) {
	if !bson.IsObjectIdHex(id) {
		return nil, errors.New("bad id")
	}
	c, ok := suite.DB[id]
	if !ok {
		return nil, errors.New("not found")
	}

	rep := (&ie.RestructedPatient{}).FromFHIR(&c.Patient)

	return rep, nil
}

func (suite *patientSuite) Patients() ([]ie.RestructedPatient, error) {
	var cc []ie.Patient
	for _, patient := range suite.DB {
		cc = append(cc, patient)
	}

	repp := make([]ie.RestructedPatient, len(cc))
	for i, patient := range cc {
		repp[i] = *(&ie.RestructedPatient{}).FromFHIR(&patient.Patient)
	}

	return repp, nil
}

// Utility Methods

func (suite *patientSuite) withTestService() web.Adapter {
	return func(h gin.HandlerFunc) gin.HandlerFunc {
		return func(ctx *gin.Context) {
			ctx.Set("service", suite)
			h(ctx)
		}
	}
}

func (suite *patientSuite) findPatient(ID string, patients []ie.RestructedPatient) *ie.RestructedPatient {
	for _, p := range patients {
		if ID == p.ID {
			return &p
		}
	}
	return nil
}
