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

type CareTeam struct {
	ID           string `bson:"_id"`
	app.CareTeam `bson:",inline"`
	// Name      string    `bson:"name,omitempty" binding:"required"`
	// Leader    string    `bson:"leader,omitempty" binding:"required"`
	// CreatedAt time.Time `bson:"created_at,omitempty"`
}

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

func (s *CareTeamService) CareTeams() ([]*app.CareTeam, error) {
	defer s.S.Close()
	mcc := []CareTeam{}
	err := s.C.Find(nil).All(&mcc)
	if err != nil {
		return nil, err
	}
	cc := make([]*app.CareTeam, len(mcc), len(mcc))
	for i, _ := range mcc {
		cc[i] = &mcc[i].CareTeam
		cc[i].ID = &mcc[i].ID
	}
	return cc, nil
}

func (s *CareTeamService) CreateCareTeam(c *app.CareTeam) error {
	defer s.S.Close()
	id := compensateForBsonFail(bson.NewObjectId().String())
	c.ID = &id
	t := time.Now()
	c.CreatedAt = &t
	mc := CareTeam{
		ID:       id,
		CareTeam: *c,
	}
	// mc := CareTeam{
	// 	ID:        id,
	// 	Name:      *c.Name,
	// 	Leader:    *c.Leader,
	// 	CreatedAt: time.Now(),
	// }
	err := s.C.Insert(&mc)
	return err
}

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

func (s *CareTeamService) DeleteCareTeam(id string) error {
	defer s.S.Close()
	if !bson.IsObjectIdHex(id) {
		return errors.New("bad id")
	}
	err := s.C.RemoveId(id)
	return err
}
