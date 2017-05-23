package mongo

import (
	"log"
	"strings"

	"github.com/intervention-engine/ie/storage"

	mgo "gopkg.in/mgo.v2"
)

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

type MongoService struct {
	S *mgo.Session
}

func NewMongoService(session *mgo.Session) *MongoService {
	return &MongoService{S: session}
}

func (m *MongoService) NewCareTeamService() storage.CareTeamService {
	s := m.S.Copy()
	c := s.DB("fhir").C("care_teams")
	return &CareTeamService{S: s, C: c}
}

func (m *MongoService) NewPatientService() storage.PatientService {
	s := m.S.Copy()
	c := s.DB("fhir").C("patients")
	return &PatientService{S: s, C: c}
}

func compensateForBsonFail(id string) string {
	result := strings.Split(id, "\"")
	if len(result) != 3 {
		return ""
	}
	return result[1]
}
