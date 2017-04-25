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

type Careteam struct {
	ID string `bson:"_id"`
	app.Careteam
	// Name      string    `bson:"name,omitempty" binding:"required"`
	// Leader    string    `bson:"leader,omitempty" binding:"required"`
	// CreatedAt time.Time `bson:"created_at,omitempty"`
}

func (s *CareTeamService) CareTeam(id string) (*app.Careteam, error) {
	defer s.S.Close()
	if !bson.IsObjectIdHex(id) {
		return nil, errors.New("bad id")
	}
	var mc Careteam
	err := s.C.FindId(id).One(&mc)
	if err != nil {
		return nil, err
	}
	c := mc.Careteam
	c.ID = &mc.ID
	return &c, nil
}

func (s *CareTeamService) CareTeams() ([]*app.Careteam, error) {
	defer s.S.Close()
	mcc := []Careteam{}
	err := s.C.Find(nil).All(&mcc)
	if err != nil {
		return nil, err
	}
	cc := make([]*app.Careteam, len(mcc), len(mcc))
	for i, _ := range mcc {
		cc[i] = &mcc[i].Careteam
		cc[i].ID = &mcc[i].ID
	}
	return cc, nil
}

func (s *CareTeamService) CreateCareTeam(c *app.Careteam) error {
	defer s.S.Close()
	id := compensateForBsonFail(bson.NewObjectId().String())
	c.ID = &id
	t := time.Now()
	c.CreatedAt = &t
	mc := Careteam{
		ID:       id,
		Careteam: *c,
	}
	// mc := Careteam{
	// 	ID:        id,
	// 	Name:      *c.Name,
	// 	Leader:    *c.Leader,
	// 	CreatedAt: time.Now(),
	// }
	err := s.C.Insert(&mc)
	return err
}

func (s *CareTeamService) UpdateCareTeam(c *app.Careteam) error {
	defer s.S.Close()
	if !bson.IsObjectIdHex(*c.ID) {
		return errors.New("bad id")
	}
	mc := Careteam{
		ID:       *c.ID,
		Careteam: *c,
	}
	err := s.C.UpdateId(*c.ID, &mc)
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
