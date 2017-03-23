package mongo

import (
	"errors"

	"github.com/intervention-engine/ie"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type PatientService struct {
	C *mgo.Collection
}

func (s *PatientService) Patient(id string) (*ie.RestructedPatient, error) {
	if !bson.IsObjectIdHex(id) {
		return nil, errors.New("bad id")
	}
	var p ie.Patient
	err := s.C.FindId(id).One(&p)
	if err != nil {
		return nil, err
	}

	rp := (&ie.RestructedPatient{}).FromFHIR(&p.Patient)

	return rp, nil
}

func (s *PatientService) Patients() ([]ie.RestructedPatient, error) {
	var pp []ie.Patient
	err := s.C.Find(nil).All(&pp)
	if err != nil {
		return nil, err
	}

	repp := make([]ie.RestructedPatient, len(pp))
	for i, patient := range pp {
		repp[i] = *(&ie.RestructedPatient{}).FromFHIR(&patient.Patient)
	}

	return repp, nil
}
