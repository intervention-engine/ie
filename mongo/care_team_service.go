package mongo

import (
	"errors"
	"time"

	"github.com/intervention-engine/ie/app"
	"gopkg.in/mgo.v2/bson"
)

// CareTeamService provides functions to interact with care teams in a database.
type CareTeamService struct {
	Service
}

// CareTeam wraps a mongo id around an app.CareTeam.
type CareTeam struct {
	ID           string `bson:"_id"`
	app.CareTeam `bson:",inline"`
}

// CareTeamMembership represents a membership between Patients and CareTeams.
type CareTeamMembership struct {
	ID         string `bson:"_id"`
	CareTeamID string `bson:"care_team_id"`
	PatientID  string `bson:"patient_id"`
}

// CareTeam gets a care team with the given id from the database.
func (s *CareTeamService) CareTeam(id string) (*app.CareTeam, error) {
	defer s.S.Close()
	if !bson.IsObjectIdHex(id) {
		return nil, errors.New("bad id")
	}
	var mc CareTeam
	err := s.C.FindId(id).One(&mc)
	if err != nil {
		return nil, err
	}
	c := mc.CareTeam
	c.ID = &mc.ID
	return &c, nil
}

// CareTeams list the care teams in a database.
func (s *CareTeamService) CareTeams() ([]*app.CareTeam, error) {
	defer s.S.Close()
	mcc := []CareTeam{}
	err := s.C.Find(nil).All(&mcc)
	if err != nil {
		return nil, err
	}
	cc := make([]*app.CareTeam, len(mcc), len(mcc))
	for i := range mcc {
		cc[i] = &mcc[i].CareTeam
		cc[i].ID = &mcc[i].ID
	}
	return cc, nil
}

// CreateCareTeam creates a care team.
func (s *CareTeamService) CreateCareTeam(c *app.CareTeam) error {
	defer s.S.Close()
	id := bson.NewObjectId().Hex()
	c.ID = &id
	t := time.Now()
	c.CreatedAt = &t
	mc := CareTeam{
		ID:       id,
		CareTeam: *c,
	}
	err := s.C.Insert(&mc)
	return err
}

// UpdateCareTeam updates a care team.
func (s *CareTeamService) UpdateCareTeam(c *app.CareTeam) error {
	defer s.S.Close()
	if c.ID == nil {
		return errors.New("nil id")
	}
	mc := CareTeam{ID: *c.ID}
	if !bson.IsObjectIdHex(mc.ID) {
		return errors.New("bad id")
	}
	err := s.C.FindId(mc.ID).One(&mc)
	if err != nil {
		return err
	}
	if c.Leader != nil {
		mc.Leader = c.Leader
	}
	if c.Name != nil {
		mc.Name = c.Name
	}
	err = s.C.UpdateId(mc.ID, &mc)
	return err
}

// DeleteCareTeam deletes a care team.
func (s *CareTeamService) DeleteCareTeam(id string) error {
	defer s.S.Close()
	if !bson.IsObjectIdHex(id) {
		return errors.New("bad id")
	}
	err := s.C.RemoveId(id)
	return err
}

// AddPatient adds a patient with the given id to this care team.
func (s *CareTeamService) AddPatient(careTeamID string, patientID string) error {
	defer s.S.Close()
	exist, err := s.validateCareTeamMembership(careTeamID, patientID)
	if err != nil {
		return err
	}
	if exist {
		return errors.New("patient already belongs to care team")
	}
	mCol := s.S.DB(s.Database).C(careTeamPatientCollection)
	mem := CareTeamMembership{
		ID:         bson.NewObjectId().Hex(),
		CareTeamID: careTeamID,
		PatientID:  patientID,
	}
	err = mCol.Insert(&mem)
	return err
}

// RemovePatient removes a patient with the given id from this care team.
func (s *CareTeamService) RemovePatient(careTeamID string, patientID string) error {
	defer s.S.Close()
	exist, err := s.validateCareTeamMembership(careTeamID, patientID)
	if err != nil {
		return err
	}
	if !exist {
		return errors.New("membership does not exist")
	}
	mCol := s.S.DB(s.Database).C(careTeamPatientCollection)
	var mem CareTeamMembership
	err = mCol.Find(bson.M{"care_team_id": careTeamID, "patient_id": patientID}).One(&mem)
	if err != nil {
		return errors.New("internal error, validated that membership existed and then could not find it to delete it")
	}
	mCol.RemoveId(mem.ID)
	// Now need to remove patient from huddles with this care team.
	hCol := s.S.DB(s.Database).C(huddleCollection)
	var mhh []Huddle
	now := time.Now()
	err = hCol.Find(bson.M{
		"careteamid": careTeamID,
		"date": bson.M{
			"$gt": now,
		},
	}).All(&mhh)
	if err != nil {
		return err
	}
	err = s.removePatientFromHuddles(mhh, patientID)
	return err
}

func (s *CareTeamService) removePatientFromHuddles(huddles []Huddle, patientID string) error {
	for _, h := range huddles {
		if h.Patients == nil {
			continue
		}
		pp, err := s.removePatientFromHuddle(h.Patients, patientID)
		if err != nil {
			if err.Error() == "patient not found" {
				continue
			}
			return err
		}
		_, err = s.updateHuddlePatientsOrDeleteHuddle(h, pp)
		if err != nil {
			return err
		}
	}
	return nil
}
