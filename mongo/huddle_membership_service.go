package mongo

import (
	"github.com/intervention-engine/ie"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

// HuddleMembershipService mongodb membership service
type HuddleMembershipService struct {
	C *mgo.Collection
}

// CreateHuddleMembership create membership in mongodb
func (m *HuddleMembershipService) CreateMembership(mem ie.HuddleMembership) error {
	return m.C.Insert(mem)
}

func (m *HuddleMembershipService) DeleteMembership(id string) error {
	return m.C.RemoveId(id)
}

// PatientHuddleMemberships list of memberships a given huddle
func (m *HuddleMembershipService) HuddleMemberships(id string) ([]ie.HuddleMembership, error) {
	return m.findMembersByKey("huddle_id", id)
}

func (m *HuddleMembershipService) PatientMemberships(id string) ([]ie.HuddleMembership, error) {
	return m.findMembersByKey("patient_id", id)
}

func (m *HuddleMembershipService) findMembersByKey(key string, id string) ([]ie.HuddleMembership, error) {
	var mems []ie.HuddleMembership
	err := m.C.Find(bson.M{key: id}).All(&mems)
	return mems, err
}
