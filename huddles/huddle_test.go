package huddles

import (
	"encoding/json"
	"io/ioutil"
	"testing"
	"time"

	"github.com/intervention-engine/fhir/models"
	"github.com/stretchr/testify/suite"
)

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestHuddleSuite(t *testing.T) {
	suite.Run(t, new(HuddleSuite))
}

type HuddleSuite struct {
	suite.Suite
	Huddle Huddle
}

func (suite *HuddleSuite) SetupTest() {
	require := suite.Require()

	data, err := ioutil.ReadFile("../fixtures/huddle.json")
	require.NoError(err)
	group := new(models.Group)
	err = json.Unmarshal(data, group)
	require.NoError(err)
	suite.Huddle = Huddle(*group)
}

func (suite *HuddleSuite) TestIsHuddle() {
	assert := suite.Assert()

	assert.True(suite.Huddle.IsHuddle())

	// Change the code and check
	suite.Huddle.Code.Coding[0].Code = "NOTAHUDDLE"
	assert.False(suite.Huddle.IsHuddle())
}

func (suite *HuddleSuite) TestActiveDateTime() {
	assert := suite.Assert()
	require := suite.Require()

	t := suite.Huddle.ActiveDateTime()
	require.NotNil(t)
	assert.Equal(time.Date(2016, time.February, 2, 9, 0, 0, 0, time.UTC), t.Time)

	// Clear the extensions and check
	suite.Huddle.Extension = []models.Extension{}
	assert.Nil(suite.Huddle.ActiveDateTime())
}

func (suite *HuddleSuite) TestLeader() {
	assert := suite.Assert()

	l := suite.Huddle.Leader()
	assert.Equal("Practitioner/9999999999999999999", l.Reference)
	assert.Equal("9999999999999999999", l.ReferencedID)
	assert.Equal("Practitioner", l.Type)
	assert.False(*l.External)

	// Clear the extensions and check
	suite.Huddle.Extension = []models.Extension{}
	assert.Nil(suite.Huddle.Leader())
}

func (suite *HuddleSuite) TestHuddleMembers() {
	assert := suite.Assert()
	require := suite.Require()

	m := suite.Huddle.HuddleMembers()
	require.Len(m, 5)
	assert.Equal("1111111111111111111", m[0].Entity.ReferencedID)
	assert.Equal("2222222222222222222", m[1].Entity.ReferencedID)
	assert.Equal("3333333333333333333", m[2].Entity.ReferencedID)
	assert.Equal("4444444444444444444", m[3].Entity.ReferencedID)
	assert.Equal("5555555555555555555", m[4].Entity.ReferencedID)

	// Clear the members and check
	suite.Huddle.Member = []models.GroupMemberComponent{}
	assert.Empty(suite.Huddle.HuddleMembers())
}

func (suite *HuddleSuite) TestFindHuddleMember() {
	assert := suite.Assert()
	require := suite.Require()

	m := suite.Huddle.FindHuddleMember("4444444444444444444")
	require.NotNil(m)
	assert.Equal("4444444444444444444", m.Entity.ReferencedID)

	// Try to find one that does not exist
	m = suite.Huddle.FindHuddleMember("6666666666666666666")
	require.Nil(m)
}

func (suite *HuddleSuite) TestReason() {
	assert := suite.Assert()
	require := suite.Require()

	m := suite.Huddle.HuddleMembers()[0]
	r := m.Reason()
	require.NotNil(r)
	assert.Len(r.Coding, 1)
	assert.True(r.MatchesCode("http://interventionengine.org/fhir/cs/huddle-member-reason", "RECENT_ADMISSION"))

	// Clear the extensions and check
	m.Extension = []models.Extension{}
	assert.Nil(m.Reason())
}

func (suite *HuddleSuite) TestReviewed() {
	assert := suite.Assert()
	require := suite.Require()

	m := suite.Huddle.HuddleMembers()[0]
	r := m.Reviewed()
	require.NotNil(r)
	assert.Equal(time.Date(2016, time.February, 2, 9, 8, 15, 0, time.UTC), r.Time)

	// Clear the extensions and check
	m.Extension = []models.Extension{}
	assert.Nil(m.Reviewed())
}
