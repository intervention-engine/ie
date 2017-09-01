package mongo

import (
	"errors"
	"log"

	"github.com/intervention-engine/ie/app"
	"github.com/intervention-engine/ie/storage"

	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

const (
	huddleCollection          = "huddles"
	patientCollection         = "patients"
	careTeamCollection        = "care_teams"
	careTeamPatientCollection = "care_team_patient"
	riskAssessmentCollection  = "riskassessments"
)

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

type ServiceFactory struct {
	Session  *mgo.Session
	Database string
}

type Service struct {
	S        *mgo.Session
	C        *mgo.Collection
	Database string
}

func NewServiceFactory(session *mgo.Session, db string) *ServiceFactory {
	return &ServiceFactory{
		Session:  session,
		Database: db,
	}
}

func (sf *ServiceFactory) NewCareTeamService() storage.CareTeamService {
	s := sf.Session.Copy()
	c := s.DB(sf.Database).C(careTeamCollection)
	return &CareTeamService{Service: Service{S: s, C: c, Database: sf.Database}}
}

func (sf *ServiceFactory) NewEventService() storage.EventService {
	s := sf.Session.Copy()
	return &EventService{Service: Service{S: s, Database: sf.Database}}
}

func (sf *ServiceFactory) NewRiskAssessmentService() storage.RiskAssessmentService {
	s := sf.Session.Copy()
	c := s.DB(sf.Database).C(riskAssessmentCollection)
	return &RiskAssessmentService{Service: Service{S: s, C: c, Database: sf.Database}}
}

func (sf *ServiceFactory) NewPatientService() storage.PatientService {
	s := sf.Session.Copy()
	c := s.DB(sf.Database).C(patientCollection)
	return &PatientService{Service: Service{S: s, C: c, Database: sf.Database}}
}

func (sf *ServiceFactory) NewHuddleService() storage.HuddleService {
	s := sf.Session.Copy()
	c := s.DB(sf.Database).C(huddleCollection)
	return &HuddleService{Service: Service{S: s, C: c, Database: sf.Database}}
}

func (svc *Service) validateCareTeamMembership(id string, patientID string) (bool, error) {
	if !bson.IsObjectIdHex(id) {
		return false, errors.New("bad care team id")
	}
	if !bson.IsObjectIdHex(patientID) {
		return false, errors.New("bad patient id")
	}
	if !svc.careTeamExists(id) {
		return false, errors.New("care team not found")
	}
	if !svc.patientExists(patientID) {
		return false, errors.New("patient not found")
	}
	if !svc.membershipExists(id, patientID) {
		return false, nil
	}
	return true, nil
}

func (svc *Service) patientExists(id string) bool {
	c := svc.S.DB(svc.Database).C(patientCollection)
	var p app.Patient
	err := c.FindId(id).One(&p)
	if err != nil {
		return false
	}
	return true
}

func (svc *Service) careTeamExists(id string) bool {
	c := svc.S.DB(svc.Database).C(careTeamCollection)
	var ct app.CareTeam
	err := c.FindId(id).One(&ct)
	if err != nil {
		return false
	}
	return true
}

func (svc *Service) membershipExists(careTeamID string, patientID string) bool {
	c := svc.S.DB(svc.Database).C(careTeamPatientCollection)
	var mem CareTeamMembership
	err := c.Find(bson.M{"care_team_id": careTeamID, "patient_id": patientID}).One(&mem)
	if err != nil {
		return false
	}
	return true
}

func (svc *Service) removePatientFromHuddle(data []*app.PatientHuddle, id string) ([]*app.PatientHuddle, error) {
	pos := -1
	for i, p := range data {
		if p.ID == nil {
			continue
		}
		if id == *p.ID {
			pos = i
			break
		}
	}
	if pos < 0 {
		return nil, errors.New("patient not found")
	}
	if len(data) == pos {
		return data[:pos], nil
	}
	return append(data[:pos], data[pos+1:]...), nil
}

func (svc *Service) updateHuddlePatientsOrDeleteHuddle(huddle Huddle, patients []*app.PatientHuddle) (*Huddle, error) {
	hCol := svc.S.DB(svc.Database).C(huddleCollection)
	if len(patients) == 0 {
		err := hCol.RemoveId(huddle.ID)
		if err != nil {
			return nil, err
		}
		return nil, nil
	}
	huddle.Patients = patients
	err := hCol.UpdateId(huddle.ID, &huddle)
	if err != nil {
		return nil, err
	}
	return &huddle, nil
}
