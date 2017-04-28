package mongo

import (
	"github.com/intervention-engine/ie/app"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type HuddleService struct {
	C *mgo.Collection
}

func (h *HuddleService) CareTeamHuddles(id string) ([]app.Huddle, error) {
	var result []app.Huddle
	err := h.C.Find(bson.M{"care_team_id": id}).All(&result)
	return result, err
}

func (h *HuddleService) UpsertHuddles(huddles []app.Huddle) error {
	for _, huddle := range huddles {
		if _, err := h.C.UpsertId(huddle.ID, huddle); err != nil {
			return err
		}
	}
	return nil
}
