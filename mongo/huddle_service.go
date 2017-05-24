package mongo

import (
	"errors"
	"time"

	"github.com/intervention-engine/ie/app"
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

// Huddles lists all huddles.
func (s *HuddleService) Huddles() ([]*app.Huddle, error) {
	defer s.S.Close()
	mhh := []Huddle{}
	err := s.C.Find(nil).All(&mhh)
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

// HuddlesForCareTeam lists the huddles for that care team.
func (s *HuddleService) HuddlesForCareTeam(id string) ([]*app.Huddle, error) {
	defer s.S.Close()
	if !bson.IsObjectIdHex(id) {
		return nil, errors.New("bad care team id")
	}
	mhh := []Huddle{}
	// Then work on filtering by date and by patient ID
	err := s.C.Find(bson.M{"careteamid": id}).All(&mhh)
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

func (s *HuddleService) ScheduleHuddle(careTeamID string, patientID string, date time.Time) (*app.Huddle, bool, error) {
	defer s.S.Close()
	exist, err := s.validateCareTeamMembership(careTeamID, patientID)
	if err != nil {
		return nil, false, err
	}
	if !exist {
		return nil, false, errors.New("membership does not exist")
	}
	var h Huddle
	reason := ""
	reasonType := "MANUAL_ADDITION"
	reviewed := false
	p := &app.PatientHuddle{
		ID:         &patientID,
		Reason:     &reason,
		ReasonType: &reasonType,
		Reviewed:   &reviewed,
	}
	// check to see if we need to make the huddle for that date
	err = s.C.Find(bson.M{"careteamid": careTeamID, "date": date}).One(&h)
	if (err != nil) && (err.Error() == "not found") {
		id := bson.NewObjectId().Hex()
		h = Huddle{
			ID: id,
			Huddle: app.Huddle{
				ID:         &id,
				CareTeamID: &careTeamID,
				Date:       &date,
				Patients:   []*app.PatientHuddle{p},
			},
		}
		err = s.C.Insert(&h)
		if err != nil {
			return nil, false, err
		}
		return &h.Huddle, true, nil
	} else if err != nil {
		return nil, false, err
	}
	// found huddle, just need to add patient to it
	h.Patients = append(h.Patients, p)
	err = s.C.UpdateId(h.ID, &h)
	if err != nil {
		return nil, false, err
	}
	return &h.Huddle, false, nil
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
