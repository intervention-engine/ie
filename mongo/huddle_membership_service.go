package mongo

import (
	"github.com/intervention-engine/ie/app"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

// HuddleMembershipService mongodb membership service
type HuddleMembershipService struct {
	C *mgo.Collection
}

// CreateHuddleMembership create membership in mongodb
func (m *HuddleMembershipService) CreateMembership(mem app.HuddleMembership) error {
	return m.C.Insert(mem)
}

func (m *HuddleMembershipService) DeleteMembership(id string) error {
	return m.C.RemoveId(id)
}

// PatientHuddleMemberships list of memberships a given huddle
func (m *HuddleMembershipService) HuddleMemberships(id string) ([]app.HuddleMembership, error) {
	return m.findMembersByKey("huddle_id", id)
}

func (m *HuddleMembershipService) PatientMemberships(id string) ([]app.HuddleMembership, error) {
	return m.findMembersByKey("patient_id", id)
}

func (m *HuddleMembershipService) findMembersByKey(key string, id string) ([]app.HuddleMembership, error) {
	var mems []app.HuddleMembership
	err := m.C.Find(bson.M{key: id}).All(&mems)
	return mems, err
}
