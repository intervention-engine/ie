package mongo

import (
	"errors"
	"strings"
	"time"

	"github.com/intervention-engine/ie"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type CareTeamService struct {
	C *mgo.Collection
}

func (s *CareTeamService) CareTeam(id string) (*ie.CareTeam, error) {
	if !bson.IsObjectIdHex(id) {
		return nil, errors.New("bad id")
	}
	var c ie.CareTeam
	err := s.C.FindId(id).One(&c)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (s *CareTeamService) CareTeams() ([]ie.CareTeam, error) {
	var cc []ie.CareTeam
	err := s.C.Find(nil).All(&cc)
	if err != nil {
		return nil, err
	}
	return cc, nil
}

func (s *CareTeamService) CreateCareTeam(c *ie.CareTeam) error {
	c.ID = compensateForBsonFail(bson.NewObjectId().String())
	c.CreatedAt = time.Now()
	err := s.C.Insert(c)
	return err
}

func (s *CareTeamService) UpdateCareTeam(c *ie.CareTeam) error {
	if !bson.IsObjectIdHex(c.ID) {
		return errors.New("bad id")
	}
	err := s.C.UpdateId(c.ID, c)
	return err
}

func (s *CareTeamService) DeleteCareTeam(id string) error {
	if !bson.IsObjectIdHex(id) {
		return errors.New("bad id")
	}
	err := s.C.RemoveId(id)
	return err
}

func compensateForBsonFail(id string) string {
	result := strings.Split(id, "\"")
	if len(result) != 3 {
		return ""
	}
	return result[1]
}
