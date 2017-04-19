package web_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
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
	DB        map[string]ie.Patient
	MembersDB map[string]ie.Membership
}

func TestPatientHandlersSuite(t *testing.T) {
	p := make(map[string]ie.Patient)
	m := make(map[string]ie.Membership)
	suite.Run(t, &patientSuite{DB: p, MembersDB: m})
}

func (suite *patientSuite) SetupTest() {
	for _, patient := range PatientsDB {
		suite.DB[patient.ID] = patient
	}

	for _, mem := range MembershipsDB {
		suite.MembersDB[mem.CareTeamID] = mem
	}
}

func (suite *patientSuite) SetupSuite() {
	api := suite.LoadGin()
	web.RegisterPatientRoutes(api, suite.withTestServices())
}

// TODO: Should make some permanent times so we can test that, too
// When there are patients, should return a slice of patients with 200 OK
func (suite *patientSuite) TestAllPatientsFound() {
	w := suite.AssertGetRequest("/api/patients", http.StatusOK)
	var body = make(map[string][]ie.Patient)
	json.NewDecoder(w.Body).Decode(&body)
	results, ok := body["patients"]
	if !ok {
		suite.T().Errorf("body did not contain an element named patients | body: %#+v\n", w.Body.String())
		return
	}
	for _, patient := range PatientsDB {
		result := suite.findPatient(patient.ID, results)

		expName := patient.Name
		name := result.Name
		suite.Assert().NotNil(result, "body did not contain patient: %s with ID: %s\n", patient.Name.Family, patient.ID)
		suite.Assert().Equal(expName.Family, name.Family)
		suite.Assert().Equal(expName.Given, name.Given)
	}
}

// If a patient with that (correct) id does not exist in the database, should
// return 404 Not Found
func (suite *patientSuite) TestGetPatientNotFound() {
	id := PatientsDB[0].ID
	delete(suite.DB, id)
	suite.AssertGetRequest("/api/patients/"+id, http.StatusNotFound)
}

func (suite *patientSuite) TestListPatientsInCareTeam() {
	suite.AssertGetRequest("/api/care_teams/58c314acb367c1ff54d19e9e/patients", http.StatusOK)
}

// When given an incorrectly formatted id, should return 400 Bad Request
func (suite *patientSuite) TestGetPatientWithBadId() {
	suite.AssertGetRequest("/api/patients/sdf", http.StatusBadRequest)
}

func (suite *patientSuite) TestAddPatientToCareTeam() {
	suite.AssertPutRequest("/api/care_teams/58c314acb367c1ff54d19e9f/patients/58938873bd90ef501e29c919", nil, http.StatusOK)
}

// If a patient with that id exists, should return the patient and 200 OK
func (suite *patientSuite) TestGetPatientFound() {
	w := suite.AssertGetRequest("/api/patients/58938873bd90ef501e29c919", http.StatusOK)
	var body = make(map[string]ie.Patient)
	json.NewDecoder(w.Body).Decode(&body)
	result, ok := body["patient"]

	if !ok {
		suite.T().Errorf("body did not contain an element named patient | body: %#+v\n", w.Body.String())
		return
	}

	patient := PatientsDB[0]
	expName := patient.Name
	name := result.Name

	suite.Assert().NotNil(result, "body did not contain patient: %s with ID: %s\n", patient.Name.Family, patient.ID)
	suite.Assert().Equal(expName.Family, name.Family)
	suite.Assert().Equal(expName.Given, name.Given)
}

// Mock Services

func (suite *patientSuite) Patient(id string) (*ie.Patient, error) {
	if !bson.IsObjectIdHex(id) {
		return nil, errors.New("bad id")
	}
	p, ok := suite.DB[id]
	if !ok {
		return nil, errors.New("not found")
	}

	return &p, nil
}

func (suite *patientSuite) Patients() ([]ie.Patient, error) {
	var pp []ie.Patient
	for _, patient := range suite.DB {
		pp = append(pp, patient)
	}

	return pp, nil
}

func (suite *patientSuite) PatientsByID(ids []string) ([]ie.Patient, error) {
	pp := make([]ie.Patient, len(ids))
	fmt.Println(ids)
	for _, id := range ids {
		pp = append(pp, suite.DB[id])
	}
	return pp, nil
}

func (suite *patientSuite) PatientMemberships(id string) ([]ie.Membership, error) {
	mems := make([]ie.Membership, len(suite.MembersDB))
	for _, mem := range suite.MembersDB {
		if mem.CareTeamID == id {
			mems = append(mems, mem)
		}
	}
	return mems, nil
}

func (suite *patientSuite) CreateMembership(mem ie.Membership) error {
	size := int64(len(suite.MembersDB))
	id := strconv.FormatInt(size, 10)
	suite.MembersDB[id] = mem
	return nil
}

// Fixtures

var MembershipsDB = []ie.Membership{
	ie.Membership{
		ID:         "",
		CareTeamID: "58c314acb367c1ff54d19e9e",
		PatientID:  "58938873bd90ef501e29c919",
		CreatedAt:  time.Now(),
	},
	ie.Membership{
		ID:         "",
		CareTeamID: "58c314acb367c1ff54d19e9f",
		PatientID:  "58c314acb367c1ff54d19e9e",
		CreatedAt:  time.Now(),
	},
	ie.Membership{
		ID:         "",
		CareTeamID: "58c314acb367c1ff54d19e9e",
		PatientID:  "576c9bcc8bd4d4bdc2ac42b5",
		CreatedAt:  time.Now(),
	},
}

var PatientsDB = []ie.Patient{
	{
		ID: "58938873bd90ef501e29c919",
		Name: ie.Name{
			Family: "Morgan1765",
			Given:  "Amy",
		},
		Address: ie.Address{
			Street: []string{
				"30377 Petterle Place",
			},
			City:       "Los Gatos",
			State:      "GA",
			PostalCode: "42586",
		},
		Gender:       "female",
		BirthDate:    generateBirthdate("1962-01-01"),
		NextHuddleID: "576c9bbf8bd4a4bdc2ac2038",
		RecentRiskAssessments: []ie.RiskAssessment{
			{
				ID:      "576c9bcf8bd4d4bdc2ac481c",
				GroupID: "576c9bcc8bd4d4bdc2ac429d",
				Value:   3,
			},
		},
	},
	{
		ID: "58c314acb367c1ff54d19e9e",
		Name: ie.Name{
			Family: "Taylor Swift",
			Given:  "TayTay a.k.a Sway Sway",
		},
		Address: ie.Address{
			Street: []string{
				"42 Wallaby Way",
			},
			City:       "Sydney",
			State:      "AUS",
			PostalCode: "xxxxx",
		},
		Gender:       "female",
		BirthDate:    generateBirthdate("1954-09-21"),
		NextHuddleID: "576c9bbf8bd4a4bdc2ac2038",
		RecentRiskAssessments: []ie.RiskAssessment{
			{
				ID:      "576c9bcf8bd4d4bdc2ac481c",
				GroupID: "576c9bcc8bd4d4bdc2ac429d",
				Value:   3,
			},
		},
	},
	{
		ID: "576c9bcc8bd4d4bdc2ac42b5",
		Name: ie.Name{
			Family: "Banks9849",
			Given:  "Jacqueline",
		},
		Address: ie.Address{
			Street: []string{
				"4691 7th Way",
			},
			City:       "Loyalton",
			State:      "WY",
			PostalCode: "01409",
		},
		Gender:       "female",
		BirthDate:    generateBirthdate("1962-01-01"),
		NextHuddleID: "576c9bbf8bd4a4bdc2ac2038",
		RecentRiskAssessments: []ie.RiskAssessment{
			{
				ID:      "576c9bcf8bd4d4bdc2ac481c",
				GroupID: "576c9bcc8bd4d4bdc2ac429d",
				Value:   3,
			},
		},
	},
}

// Utility Methods

func generateBirthdate(birthdate string) *models.FHIRDateTime {
	date, err := time.ParseInLocation("2006-01-02", birthdate, time.Local)
	if err != nil {
		log.Fatalln(err)
	}
	return &models.FHIRDateTime{Time: date, Precision: models.Precision("date")}
}

func (suite *patientSuite) withTestServices() ie.Adapter {
	return func(h gin.HandlerFunc) gin.HandlerFunc {
		return func(ctx *gin.Context) {
			ctx.Set("patientService", suite)
			ctx.Set("membershipService", suite)
			h(ctx)
		}
	}
}

func (suite *patientSuite) findPatient(ID string, patients []ie.Patient) *ie.Patient {
	for _, p := range patients {
		if ID == p.ID {
			return &p
		}
	}
	return nil
}
