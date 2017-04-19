package mongo

import (
	"errors"
	"time"

	"github.com/intervention-engine/ie/app"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type CareTeamService struct {
	S *mgo.Session
	C *mgo.Collection
}

func (s *CareTeamService) CareTeam(id string) (*app.Careteam, error) {
	defer s.S.Close()
	if !bson.IsObjectIdHex(id) {
		return nil, errors.New("bad id")
	}
	var c app.Careteam
	err := s.C.FindId(id).One(&c)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (s *CareTeamService) CareTeams() ([]*app.Careteam, error) {
	defer s.S.Close()
	var cc []*app.Careteam
	err := s.C.Find(nil).All(&cc)
	if err != nil {
		return nil, err
	}
	return cc, nil
}

func (s *CareTeamService) CreateCareTeam(c *app.Careteam) error {
	defer s.S.Close()
	id := compensateForBsonFail(bson.NewObjectId().String())
	t := time.Now()
	c.ID = &id
	c.CreatedAt = &t
	err := s.C.Insert(c)
	return err
}

func (s *CareTeamService) UpdateCareTeam(c *app.Careteam) error {
	defer s.S.Close()
	if !bson.IsObjectIdHex(*c.ID) {
		return errors.New("bad id")
	}
	err := s.C.UpdateId(c.ID, c)
	return err
}

func (s *CareTeamService) DeleteCareTeam(id string) error {
	defer s.S.Close()
	if !bson.IsObjectIdHex(id) {
		return errors.New("bad id")
	}
	err := s.C.RemoveId(id)
	return err
}
