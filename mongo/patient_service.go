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

func (s *PatientService) Patient(id string) (*ie.Patient, error) {
	if !bson.IsObjectIdHex(id) {
		return nil, errors.New("bad id")
	}
	var p ie.Patient
	err := s.C.FindId(id).One(&p)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (s *PatientService) Patients() ([]ie.Patient, error) {
	var pp []ie.Patient
	err := s.C.FindId(nil).All(&pp)
	if err != nil {
		return nil, err
	}
	return pp, nil
}
