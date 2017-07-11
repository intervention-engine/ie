package mongo

import (
	"errors"
	"time"

	"github.com/intervention-engine/ie/app"
	"github.com/intervention-engine/ie/storage"
	"gopkg.in/mgo.v2/bson"
)

// HuddleService provides functions to interact with huddles.
type HuddleService struct {
	Service
}

type Huddle struct {
	ID         string `bson:"_id"`
	app.Huddle `bson:",inline"`
}

// HuddlesFilterBy lists the huddles for that care team with the given filters.
// If no filters are given, all huddles are returned for that CareTeam.
func (s *HuddleService) HuddlesFilterBy(query storage.HuddleFilterQuery) ([]*app.Huddle, error) {
	defer s.S.Close()
	if !bson.IsObjectIdHex(query.CareTeamID) {
		return nil, errors.New("bad care team id")
	}
	mhh := []Huddle{}
	mongoQuery := bson.M{"careteamid": query.CareTeamID}
	if query.PatientID != "" {
		if !bson.IsObjectIdHex(query.PatientID) {
			return nil, errors.New("bad patient id")
		}
		mongoQuery["patients.id"] = query.PatientID
	}
	if !query.Date.IsZero() {
		mongoQuery["date"] = struct {
			Gte time.Time `bson:"$gte"`
			Lt  time.Time `bson:"$lt"`
		}{
			Gte: query.Date,
			Lt:  query.Date.AddDate(0, 0, 1),
		}
	}
	err := s.C.Find(mongoQuery).All(&mhh)
	if err != nil {
		return nil, err
	}
	hh := make([]*app.Huddle, len(mhh), len(mhh))
	for i := range mhh {
		hh[i] = &mhh[i].Huddle
		hh[i].ID = &mhh[i].ID
	}
	return hh, nil
}

func (s *HuddleService) ScheduleHuddle(careTeamID string, patientID string, huddleID string) (*app.Huddle, error) {
	defer s.S.Close()
	exist, err := s.validateCareTeamMembership(careTeamID, patientID)
	if err != nil {
		return nil, err
	}
	if !exist {
		return nil, errors.New("membership does not exist")
	}
	var h Huddle
	reason := ""
	reasonType := "MANUAL_ADDITION"
	reviewed := false
	p := &app.HuddlePatient{
		ID:         &patientID,
		Reason:     &reason,
		ReasonType: &reasonType,
		Reviewed:   &reviewed,
	}
	err = s.C.FindId(huddleID).One(&h)
	if err != nil {
		if err.Error() == "not found" {
			return nil, errors.New("huddle not found")
		}
		return nil, err
	}
	// found huddle, just need to add patient to it
	h.Patients = append(h.Patients, p)
	err = s.C.UpdateId(h.ID, &h)
	if err != nil {
		return nil, err
	}
	return &h.Huddle, nil
}

// DeletePatient removes the patient from a huddle.
func (s *HuddleService) DeletePatient(id string, patientID string) (*app.Huddle, error) {
	defer s.S.Close()
	if !bson.IsObjectIdHex(id) {
		return nil, errors.New("bad huddle id")
	}
	if !bson.IsObjectIdHex(patientID) {
		return nil, errors.New("bad patient id")
	}
	var h Huddle
	err := s.C.FindId(id).One(&h)
	if err != nil {
		if err.Error() == "not found" {
			return nil, errors.New("huddle not found")
		}
		return nil, err
	}
	if h.Patients == nil || len(h.Patients) == 0 {
		return nil, errors.New("patient not found")
	}
	pp, err := s.removePatientFromHuddle(h.Patients, patientID)
	if err != nil {
		return nil, err
	}
	updatedHuddle, err := s.updateHuddlePatientsOrDeleteHuddle(h, pp)
	if err != nil {
		return nil, err
	}
	if updatedHuddle == nil {
		return nil, nil
	}
	return &h.Huddle, nil
}
