package huddles

import (
	"fmt"
	"math"
	"strings"
	"testing"
	"time"

	"github.com/intervention-engine/fhir/models"
	"github.com/intervention-engine/fhir/server"
	"github.com/intervention-engine/ie"
	"github.com/intervention-engine/ie/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"gopkg.in/mgo.v2/bson"
)

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestHuddleSchedulerSuite(t *testing.T) {
	suite.Run(t, new(HuddleSchedulerSuite))
}

type HuddleSchedulerSuite struct {
	testutil.MongoSuite
}

func (suite *HuddleSchedulerSuite) SetupTest() {
	server.Database = suite.DB()
}

func (suite *HuddleSchedulerSuite) TearDownTest() {
	suite.TearDownDB()
}

func (suite *HuddleSchedulerSuite) TearDownSuite() {
	suite.TearDownDBServer()
}

func (suite *HuddleSchedulerSuite) TestCreatePopulatedHuddleForNewHuddle() {
	assert := assert.New(suite.T())
	require := require.New(suite.T())

	// PATIENT                                        // Hex | FREQ.   | LAST HUDDLE        | DUE BY
	suite.storePatientAndScores(bsonID(1), 1, 1, 1)   // 01  | never   |                    |
	suite.storePatientAndScores(bsonID(2), 1, 2, 2)   // 02  | never   |                    |
	suite.storePatientAndScores(bsonID(3), 2, 1, 1)   // 03  | never   |                    | * (see encounters below)
	suite.storePatientAndScores(bsonID(4), 1, 2, 1)   // 04  | never   |                    |
	suite.storePatientAndScores(bsonID(5), 2, 3, 2)   // 05  | never   | 5 weeks ago (2/01) |
	suite.storePatientAndScores(bsonID(6), 1, 2, 3)   // 06  | 4 weeks | 1 week ago  (3/14) | 4/11
	suite.storePatientAndScores(bsonID(7), 6, 6, 5)   // 07  | 4 weeks | 5 weeks ago (2/01) | 3/21 (overdue)
	suite.storePatientAndScores(bsonID(8), 5, 5, 4)   // 08  | 4 weeks |                    | 4/11 (at latest)
	suite.storePatientAndScores(bsonID(9), 3, 3, 3)   // 09  | 4 weeks | 2 weeks ago (3/07) | 4/04
	suite.storePatientAndScores(bsonID(10), 4, 5, 4)  // 0a  | 4 weeks | 4 weeks ago (2/22) | 3/21
	suite.storePatientAndScores(bsonID(11), 5, 4, 6)  // 0b  | 2 weeks | 1 week ago  (3/14) | 3/28 *
	suite.storePatientAndScores(bsonID(12), 5, 6, 6)  // 0c  | 2 weeks | 3 weeks ago (2/29) | 3/21 (overdue)
	suite.storePatientAndScores(bsonID(13), 6, 7, 6)  // 0d  | 2 weeks |                    | 3/28 (at latest)
	suite.storePatientAndScores(bsonID(14), 8, 8, 7)  // 0e  | 2 weeks | 1 week ago  (3/14) | 3/28
	suite.storePatientAndScores(bsonID(15), 5, 6, 7)  // 0f  | 2 weeks | 2 weeks ago (3/07) | 3/21
	suite.storePatientAndScores(bsonID(16), 9, 9, 9)  // 10  | 1 week  |                    | 3/21
	suite.storePatientAndScores(bsonID(17), 8, 9, 8)  // 11  | 1 week  | 1 week ago  (3/14) | 3/21
	suite.storePatientAndScores(bsonID(18), 7, 7, 8)  // 12  | 1 week  | 2 weeks ago (3/07) | 3/21
	suite.storePatientAndScores(bsonID(19), 9, 9, 9)  // 13  | 1 week  | 1 week ago  (3/14) | 3/21 *
	suite.storePatientAndScores(bsonID(20), 9, 9, 10) // 14  | 1 week  | 1 week ago  (3/14) | 3/21

	// ENCOUNTERS
	adm := time.Date(2016, time.March, 11, 12, 30, 0, 0, time.UTC)
	dis := time.Date(2016, time.March, 19, 9, 0, 0, 0, time.UTC)
	suite.storeEncounter(bsonID(3), "HOSP", &adm, &dis) // 2 days ago, patient Ox03 should get scheduled
	adm = time.Date(2016, time.January, 8, 9, 30, 0, 0, time.UTC)
	dis = time.Date(2016, time.January, 9, 9, 0, 0, 0, time.UTC)
	suite.storeEncounter(bsonID(6), "HOSP", &adm, &dis) // months ago, patient Ox06 should NOT get scheduled
	adm = time.Date(2016, time.March, 15, 9, 30, 0, 0, time.UTC)
	dis = time.Date(2016, time.March, 18, 9, 0, 0, 0, time.UTC)
	suite.storeEncounter(bsonID(11), "ER", &adm, &dis) // 3 days ago, patient Ox0b should get scheduled
	adm = time.Date(2016, time.March, 14, 9, 30, 0, 0, time.UTC)
	dis = time.Date(2016, time.March, 16, 9, 0, 0, 0, time.UTC)
	suite.storeEncounter(bsonID(19), "ER", &adm, &dis) // 3 days ago, patient Ox13 should get scheduled

	config := createHuddleConfig(true, true, 0, time.Monday)

	suite.storeHuddle(time.Date(2016, time.February, 1, 0, 0, 0, 0, time.UTC), config.LeaderID, riskScoreReason(),
		bsonID(5), bsonID(7))
	suite.storeHuddle(time.Date(2016, time.February, 22, 0, 0, 0, 0, time.UTC), config.LeaderID, riskScoreReason(),
		bsonID(10))
	suite.storeHuddle(time.Date(2016, time.February, 29, 0, 0, 0, 0, time.UTC), config.LeaderID, riskScoreReason(),
		bsonID(12))
	suite.storeHuddle(time.Date(2016, time.March, 7, 0, 0, 0, 0, time.UTC), config.LeaderID, riskScoreReason(),
		bsonID(9), bsonID(15), bsonID(18))
	suite.storeHuddle(time.Date(2016, time.March, 14, 0, 0, 0, 0, time.UTC), config.LeaderID, riskScoreReason(),
		bsonID(6), bsonID(11), bsonID(14), bsonID(17), bsonID(19), bsonID(20))

	group, err := createPopulatedHuddle(time.Date(2016, time.March, 21, 0, 0, 0, 0, time.UTC), config, false)
	require.NoError(err)
	ha := NewHuddleAssertions(group, assert)
	ha.AssertValidHuddleProfile()
	ha.AssertActiveDateTimeEqual(time.Date(2016, time.March, 21, 0, 0, 0, 0, time.UTC))
	ha.AssertLeaderIDEqual("123")
	ha.AssertNameEqual("Test Huddle Config")
	assert.Len(group.Member, 11)

	members := ha.HuddleMembers()
	assert.Len(members, 11)
	for _, id := range []string{bsonID(3), bsonID(7), bsonID(10), bsonID(11), bsonID(12), bsonID(15), bsonID(16),
		bsonID(17), bsonID(18), bsonID(19), bsonID(20)} {
		assert.NotNil(ha.FindHuddleMember(id))
	}
	ha.AssertMember(0, bsonID(11), recentEncounterReason("Emergency Room Visit"))
	ha.AssertMember(1, bsonID(19), recentEncounterReason("Emergency Room Visit"))
	ha.AssertMember(2, bsonID(3), recentEncounterReason("Hospital Discharge"))
	for i := 3; i < 11; i++ {
		assert.Equal(riskScoreReason(), members[i].Reason())
	}
}

func createPopulatedHuddle(date time.Time, config *HuddleConfig, doRollOver bool) (*models.Group, error) {
	oldLookAhead := config.LookAhead
	config.LookAhead = 1
	_nowValueForTestingOnly = &date
	defer func() {
		config.LookAhead = oldLookAhead
		_nowValueForTestingOnly = nil
	}()
	huddles, err := NewHuddleScheduler(config).ScheduleHuddles()
	if err != nil {
		return nil, err
	}
	group := models.Group(*huddles[0])
	return &group, nil
}

func ScheduleHuddles(config *HuddleConfig) ([]*models.Group, error) {
	huddles, err := NewHuddleScheduler(config).ScheduleHuddles()
	if err != nil {
		return nil, err
	}
	groups := make([]*models.Group, len(huddles))
	for i := range huddles {
		group := models.Group(*huddles[i])
		groups[i] = &group
	}
	return groups, nil
}

// This test just ensures that manually added patients aren't overwritten by the huddle population
// algorithm.
func (suite *HuddleSchedulerSuite) TestCreatePopulatedHuddleForExistingHuddle() {
	assert := assert.New(suite.T())
	require := require.New(suite.T())

	// PATIENT                                      // Hex | FREQ.   | LAST HUDDLE        | DUE BY
	suite.storePatientAndScores(bsonID(1), 1, 1, 1) // 01  | never   |                    |
	suite.storePatientAndScores(bsonID(2), 5, 5, 4) // 02  | 4 weeks |                    | 4/11 (at latest)
	suite.storePatientAndScores(bsonID(3), 8, 8, 7) // 03  | 2 weeks | 1 week ago  (3/14) | 3/28
	suite.storePatientAndScores(bsonID(4), 8, 9, 8) // 04  | 1 week  | 1 week ago  (3/14) | 3/21

	config := createHuddleConfig(true, true, 0, time.Monday)
	suite.storeHuddle(time.Date(2016, time.March, 14, 0, 0, 0, 0, time.Local), config.LeaderID, riskScoreReason(), bsonID(3), bsonID(4))
	hunchReason := manualAdditionReason("I've got a hunch")
	suite.storeHuddleWithDetails(time.Date(2016, time.March, 21, 0, 0, 0, 0, time.Local), config.LeaderID, riskScoreReason(),
		map[string]*models.CodeableConcept{bsonID(1): hunchReason}, nil, bsonID(1), bsonID(2), bsonID(3), bsonID(4))

	group, err := createPopulatedHuddle(time.Date(2016, time.March, 21, 0, 0, 0, 0, time.Local), config, false)
	require.NoError(err)

	ha := NewHuddleAssertions(group, assert)
	ha.AssertValidHuddleProfile()
	ha.AssertActiveDateTimeEqual(time.Date(2016, time.March, 21, 0, 0, 0, 0, time.Local))
	ha.AssertLeaderIDEqual("123")
	assert.True(strings.HasPrefix(ha.Name, "Test Huddle"))

	require.Len(ha.Member, 2)
	ha.AssertMember(0, bsonID(1), hunchReason)
	ha.AssertMember(1, bsonID(4), riskScoreReason())
}

func (suite *HuddleSchedulerSuite) TestScheduleHuddlesByRiskScore() {
	assert := assert.New(suite.T())
	require := require.New(suite.T())

	suite.storePatientAndScores(bsonID(1), 5, 4, 7)  // every 2 weeks
	suite.storePatientAndScores(bsonID(2), 1, 1, 1)  // never
	suite.storePatientAndScores(bsonID(3), 1, 2, 3)  // every 4 weeks
	suite.storePatientAndScores(bsonID(4), 9, 9, 10) // every week
	suite.storePatientAndScores(bsonID(5), 7, 6, 5)  // every 4 weeks

	config := createHuddleConfig(true, false, 0, time.Monday)
	huddles, err := ScheduleHuddles(config)
	require.NoError(err)
	assert.Len(huddles, 4)
	for i := range huddles {
		ha := NewHuddleAssertions(huddles[i], assert)
		ha.AssertValidHuddleProfile()
		ha.AssertLeaderIDEqual("123")
		ha.AssertNameEqual("Test Huddle Config")
	}

	// Find next Monday to start checking dates
	t := time.Now()
	for ; t.Weekday() != time.Monday; t = t.AddDate(0, 0, 1) {
	}
	t = time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
	// Now check each one individually
	ha := NewHuddleAssertions(huddles[0], assert)
	ha.AssertActiveDateTimeEqual(t)
	ha.AssertMemberIDs(bsonID(4), bsonID(1))

	ha = NewHuddleAssertions(huddles[1], assert)
	ha.AssertActiveDateTimeEqual(t.AddDate(0, 0, 7))
	ha.AssertMemberIDs(bsonID(4), bsonID(5))

	ha = NewHuddleAssertions(huddles[2], assert)
	ha.AssertActiveDateTimeEqual(t.AddDate(0, 0, 14))
	ha.AssertMemberIDs(bsonID(4), bsonID(1))

	ha = NewHuddleAssertions(huddles[3], assert)
	ha.AssertActiveDateTimeEqual(t.AddDate(0, 0, 21))
	// Patient 3 comes first because they're both due, but 3 has never been discussed before
	ha.AssertMemberIDs(bsonID(3), bsonID(4))

	// Now just make sure they were really stored to the db
	var storedHuddles []*models.Group
	server.Database.C("groups").Find(bson.M{}).Sort("extension.activeDateTime").All(&storedHuddles)
	assert.Equal(huddles, storedHuddles, "Stored huddles should match returned huddles")
}

func (suite *HuddleSchedulerSuite) TestScheduleHuddlesByEncounterEvents() {
	assert := assert.New(suite.T())
	require := require.New(suite.T())

	suite.storePatientAndScores(bsonID(1), 7)  // every 2 weeks
	suite.storePatientAndScores(bsonID(2), 1)  // never
	suite.storePatientAndScores(bsonID(3), 3)  // every 4 weeks
	suite.storePatientAndScores(bsonID(4), 10) // every week
	suite.storePatientAndScores(bsonID(5), 5)  // every 4 weeks

	// Add some encounters
	t := time.Now()
	daysAgo := func(days int) *time.Time {
		result := t.AddDate(0, 0, -1*days)
		return &result
	}
	suite.storeEncounter(bsonID(1), "HOSP", daysAgo(15), daysAgo(10)) // 10 days ago -- don't trigger
	suite.storeEncounter(bsonID(2), "HOSP", daysAgo(10), daysAgo(4))  // discharged 4 days ago -- trigger discharge
	suite.storeEncounter(bsonID(3), "HOSP", daysAgo(6), daysAgo(4))   // whole stay in last 7 days -- trigger discharge
	suite.storeEncounter(bsonID(4), "HOSP", daysAgo(6), nil)          // admit 6 days ago, ongoing -- trigger admit
	suite.storeEncounter(bsonID(5), "HOSP", daysAgo(10), nil)         // admit 10 days ago, ongoing -- don't trigger
	suite.storeEncounter(bsonID(6), "ER", daysAgo(1), daysAgo(1))     // ED visit yesterday -- trigger ED
	suite.storeEncounter(bsonID(7), "FOO", daysAgo(2), daysAgo(1))    // Unrecognized code -- don't trigger

	config := createHuddleConfig(false, true, 0, t.Weekday())
	huddles, err := ScheduleHuddles(config)
	require.NoError(err)
	assert.Len(huddles, 4)
	for i := range huddles {
		ha := NewHuddleAssertions(huddles[i], assert)
		ha.AssertValidHuddleProfile()
		ha.AssertLeaderIDEqual("123")
		ha.AssertNameEqual("Test Huddle Config")
	}

	// Set to midnight to start checking dates
	t = time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())

	// Now check each huddle
	ha := NewHuddleAssertions(huddles[0], assert)
	ha.AssertActiveDateTimeEqual(t)
	ha.AssertMemberIDs(bsonID(6), bsonID(3), bsonID(4), bsonID(2))

	ha = NewHuddleAssertions(huddles[1], assert)
	ha.AssertActiveDateTimeEqual(t.AddDate(0, 0, 7))
	assert.Empty(huddles[1].Member)

	ha = NewHuddleAssertions(huddles[2], assert)
	ha.AssertActiveDateTimeEqual(t.AddDate(0, 0, 14))
	assert.Empty(huddles[2].Member)

	ha = NewHuddleAssertions(huddles[3], assert)
	ha.AssertActiveDateTimeEqual(t.AddDate(0, 0, 21))
	assert.Empty(huddles[3].Member)

	// And check the details of the first huddle (note: should be in order of descending start date)
	huddle := Huddle(*huddles[0])
	m := huddle.FindHuddleMember(bsonID(2))
	require.NotNil(m)
	assert.Equal("Hospital Discharge", m.Reason().Text)
	m = huddle.FindHuddleMember(bsonID(3))
	require.NotNil(m)
	assert.Equal("Hospital Discharge", m.Reason().Text)
	m = huddle.FindHuddleMember(bsonID(4))
	require.NotNil(m)
	assert.Equal("Hospital Admission", m.Reason().Text)
	m = huddle.FindHuddleMember(bsonID(6))
	require.NotNil(m)
	assert.Equal("Emergency Room Visit", m.Reason().Text)

	// Now just make sure they were really stored to the db
	var storedHuddles []*models.Group
	server.Database.C("groups").Find(bson.M{}).Sort("extension.activeDateTime").All(&storedHuddles)
	assert.Equal(huddles, storedHuddles, "Stored huddles should match returned huddles")
}

// TestScheduleHuddlesByEncounterEventsWithReschedules tests for a bug we encountered where a patient would be
// scheduled correctly the first time, but if the scheduling algorithm was run again, the patient would be bumped
// one schedule further (and again and again every time the algorithm is run)
func (suite *HuddleSchedulerSuite) TestScheduleHuddlesByEncounterEventsWithReschedules() {
	assert := assert.New(suite.T())
	require := require.New(suite.T())

	suite.storePatientAndScores(bsonID(1), 7)

	// Add an encounter
	t := time.Now()
	yesterday := t.AddDate(0, 0, -1)
	suite.storeEncounter(bsonID(1), "ER", &yesterday, &yesterday) // ED visit yesterday -- trigger ED

	// Set t to midnight (huddle time) to start checking dates
	t = time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
	config := createHuddleConfig(false, true, 0, t.Weekday())
	config.EventConfig.EncounterConfigs[0].LookBackDays = 30

	// Do this five times just to ensure the same results every time
	for i := 0; i < 5; i++ {
		huddles, err := ScheduleHuddles(config)
		require.NoError(err)
		assert.Len(huddles, 4)

		// Now check each huddle
		ha := NewHuddleAssertions(huddles[0], assert)
		ha.AssertActiveDateTimeEqual(t)
		ha.AssertMemberIDs(bsonID(1))

		ha = NewHuddleAssertions(huddles[1], assert)
		ha.AssertActiveDateTimeEqual(t.AddDate(0, 0, 7))
		assert.Empty(huddles[1].Member)

		ha = NewHuddleAssertions(huddles[2], assert)
		ha.AssertActiveDateTimeEqual(t.AddDate(0, 0, 14))
		assert.Empty(huddles[2].Member)

		ha = NewHuddleAssertions(huddles[3], assert)
		ha.AssertActiveDateTimeEqual(t.AddDate(0, 0, 21))
		assert.Empty(huddles[3].Member)

		// And check the details of the first huddle
		huddle := Huddle(*huddles[0])
		m := huddle.FindHuddleMember(bsonID(1))
		require.NotNil(m)
		assert.Equal("Emergency Room Visit", m.Reason().Text)
	}
}

// Test for bug logged in pivotal: https://www.pivotaltracker.com/story/show/117859291
func (suite *HuddleSchedulerSuite) TestManuallyAddPatientToExistingHuddle() {
	assert := assert.New(suite.T())
	require := require.New(suite.T())

	// Create a patient that needs to be discussed every week, to ensure weekly scheduled huddles
	suite.storePatientAndScores(bsonID(1), 8)
	// Create a patient that needs to be discussed every 2 weeks, whom we'll manually schedule on an off-week
	suite.storePatientAndScores(bsonID(2), 7)

	t := time.Now()
	config := createHuddleConfig(true, true, 0, t.Weekday())
	// The bug indicates there must be a huddle TODAY, so set the day of the week to today's
	config.Days[0] = time.Now().Weekday()

	// Schedule the huddles
	huddles, err := ScheduleHuddles(config)
	require.NoError(err)
	assert.Len(huddles, 4)
	for i := range huddles {
		ha := NewHuddleAssertions(huddles[i], assert)
		ha.AssertValidHuddleProfile()
		ha.AssertLeaderIDEqual("123")
		ha.AssertNameEqual("Test Huddle Config")
	}

	// Find today @ midnight to start checking dates
	t = time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())

	// Check each one individually to ensure it's as expected
	ha := NewHuddleAssertions(huddles[0], assert)
	ha.AssertActiveDateTimeEqual(t)
	ha.AssertMemberIDs(bsonID(1), bsonID(2))

	ha = NewHuddleAssertions(huddles[1], assert)
	ha.AssertActiveDateTimeEqual(t.AddDate(0, 0, 7))
	ha.AssertMemberIDs(bsonID(1))

	ha = NewHuddleAssertions(huddles[2], assert)
	ha.AssertActiveDateTimeEqual(t.AddDate(0, 0, 14))
	ha.AssertMemberIDs(bsonID(1), bsonID(2))

	ha = NewHuddleAssertions(huddles[3], assert)
	ha.AssertActiveDateTimeEqual(t.AddDate(0, 0, 21))
	ha.AssertMemberIDs(bsonID(1))

	// Now manually schedule patient 2 to the first huddle (which should be an off-week for the patient)
	huddles[0].Member = append(huddles[0].Member, models.GroupMemberComponent{
		BackboneElement: models.BackboneElement{
			Element: models.Element{
				Extension: []models.Extension{
					{
						Url:                  "http://interventionengine.org/fhir/extension/group/member/reason",
						ValueCodeableConcept: manualAdditionReason("Just Because"),
					},
				},
			},
		},
		Entity: &models.Reference{
			Reference:    "Patient/" + bsonID(2),
			ReferencedID: bsonID(2),
			Type:         "Patient",
			External:     new(bool),
		},
	})
	err = suite.DB().C("groups").UpdateId(huddles[0].Id, huddles[0])
	require.NoError(err)

	// Reschedule the huddles
	huddles, err = ScheduleHuddles(config)
	require.NoError(err)
	assert.Len(huddles, 4)

	// Check each one again to ensure it's as expected
	ha = NewHuddleAssertions(huddles[0], assert)
	ha.AssertActiveDateTimeEqual(t)
	// Should have patient 2 now because patient 2 was manually added for this week
	ha.AssertMemberIDs(bsonID(2), bsonID(1))

	ha = NewHuddleAssertions(huddles[1], assert)
	ha.AssertActiveDateTimeEqual(t.AddDate(0, 0, 7))
	// Should no longer have patient 2 because the manual addition reset the every-2-week cadence
	ha.AssertMemberIDs(bsonID(1))

	ha = NewHuddleAssertions(huddles[2], assert)
	ha.AssertActiveDateTimeEqual(t.AddDate(0, 0, 14))
	// Should have patient 2 now due to the new cadence
	ha.AssertMemberIDs(bsonID(1), bsonID(2))

	ha = NewHuddleAssertions(huddles[3], assert)
	ha.AssertActiveDateTimeEqual(t.AddDate(0, 0, 21))
	// Should no longer have patient 2 because the manual addition reset the every-2-week cadence
	ha.AssertMemberIDs(bsonID(1))

	// Now just make sure they were really stored to the db
	var storedHuddles []*models.Group
	server.Database.C("groups").Find(bson.M{}).Sort("extension.activeDateTime").All(&storedHuddles)
	assert.Equal(huddles, storedHuddles, "Stored huddles should match returned huddles")
}

// Test for bug logged in Jira: https://interventionengine.atlassian.net/browse/IE-48
func (suite *HuddleSchedulerSuite) TestManuallyAddMultiplePatientsToTodaysHuddle() {
	assert := assert.New(suite.T())
	require := require.New(suite.T())

	// Create a patient that needs to be discussed every week, to ensure weekly scheduled huddles
	suite.storePatientAndScores(bsonID(1), 8)
	// Create a patient that needs to be discussed every 2 weeks
	suite.storePatientAndScores(bsonID(2), 7)
	// Create some patients that are low-risk and don't generally need discussion, whom we'll manually schedule
	suite.storePatientAndScores(bsonID(3), 1)
	suite.storePatientAndScores(bsonID(4), 1)
	suite.storePatientAndScores(bsonID(5), 1)

	t := time.Now()
	config := createHuddleConfig(true, true, 0, t.Weekday())
	config.Days[0] = time.Now().Weekday()

	// Schedule the huddles
	huddles, err := ScheduleHuddles(config)
	require.NoError(err)
	assert.Len(huddles, 4)
	for i := range huddles {
		ha := NewHuddleAssertions(huddles[i], assert)
		ha.AssertValidHuddleProfile()
		ha.AssertLeaderIDEqual("123")
		ha.AssertNameEqual("Test Huddle Config")
	}

	// Find today @ midnight to start checking dates
	t = time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())

	// Check each one individually to ensure it's as expected
	ha := NewHuddleAssertions(huddles[0], assert)
	ha.AssertActiveDateTimeEqual(t)
	ha.AssertMemberIDs(bsonID(1), bsonID(2))

	ha = NewHuddleAssertions(huddles[1], assert)
	ha.AssertActiveDateTimeEqual(t.AddDate(0, 0, 7))
	ha.AssertMemberIDs(bsonID(1))

	ha = NewHuddleAssertions(huddles[2], assert)
	ha.AssertActiveDateTimeEqual(t.AddDate(0, 0, 14))
	ha.AssertMemberIDs(bsonID(1), bsonID(2))

	ha = NewHuddleAssertions(huddles[3], assert)
	ha.AssertActiveDateTimeEqual(t.AddDate(0, 0, 21))
	ha.AssertMemberIDs(bsonID(1))

	// Now manually schedule patients to today's huddle
	huddles[0].Member = append(huddles[0].Member, models.GroupMemberComponent{
		BackboneElement: models.BackboneElement{
			Element: models.Element{
				Extension: []models.Extension{
					{
						Url:                  "http://interventionengine.org/fhir/extension/group/member/reason",
						ValueCodeableConcept: manualAdditionReason("I have my reasons"),
					},
				},
			},
		},
		Entity: &models.Reference{
			Reference:    "Patient/" + bsonID(3),
			ReferencedID: bsonID(3),
			Type:         "Patient",
			External:     new(bool),
		},
	})
	huddles[0].Member = append(huddles[0].Member, models.GroupMemberComponent{
		BackboneElement: models.BackboneElement{
			Element: models.Element{
				Extension: []models.Extension{
					{
						Url:                  "http://interventionengine.org/fhir/extension/group/member/reason",
						ValueCodeableConcept: manualAdditionReason("Because I said so"),
					},
				},
			},
		},
		Entity: &models.Reference{
			Reference:    "Patient/" + bsonID(4),
			ReferencedID: bsonID(4),
			Type:         "Patient",
			External:     new(bool),
		},
	})
	huddles[0].Member = append(huddles[0].Member, models.GroupMemberComponent{
		BackboneElement: models.BackboneElement{
			Element: models.Element{
				Extension: []models.Extension{
					{
						Url:                  "http://interventionengine.org/fhir/extension/group/member/reason",
						ValueCodeableConcept: manualAdditionReason("On a hunch"),
					},
				},
			},
		},
		Entity: &models.Reference{
			Reference:    "Patient/" + bsonID(5),
			ReferencedID: bsonID(5),
			Type:         "Patient",
			External:     new(bool),
		},
	})
	err = suite.DB().C("groups").UpdateId(huddles[0].Id, huddles[0])
	require.NoError(err)

	// Reschedule the huddles -- this is where the extra manually added patients seem to get dropped
	huddles, err = ScheduleHuddles(config)
	require.NoError(err)
	assert.Len(huddles, 4)

	// Check each one again to ensure it's as expected
	ha = NewHuddleAssertions(huddles[0], assert)
	ha.AssertActiveDateTimeEqual(t)
	// Should have patients 3, 4, and 5 (in addition to patient 1)
	ha.AssertMemberIDs(bsonID(3), bsonID(4), bsonID(5), bsonID(1))

	ha = NewHuddleAssertions(huddles[1], assert)
	ha.AssertActiveDateTimeEqual(t.AddDate(0, 0, 7))
	// Patent 2 comes first since he never had a huddle before
	ha.AssertMemberIDs(bsonID(2), bsonID(1))

	ha = NewHuddleAssertions(huddles[2], assert)
	ha.AssertActiveDateTimeEqual(t.AddDate(0, 0, 14))
	ha.AssertMemberIDs(bsonID(1))

	ha = NewHuddleAssertions(huddles[3], assert)
	ha.AssertActiveDateTimeEqual(t.AddDate(0, 0, 21))
	ha.AssertMemberIDs(bsonID(1), bsonID(2))

	// Now just make sure they were really stored to the db
	var storedHuddles []*models.Group
	server.Database.C("groups").Find(bson.M{}).Sort("extension.activeDateTime").All(&storedHuddles)
	assert.Equal(huddles, storedHuddles, "Stored huddles should match returned huddles")
}

// Test for bug logged in jira: https://interventionengine.atlassian.net/browse/IE-11
func (suite *HuddleSchedulerSuite) TestManuallyAddPatientToTodaysHuddle() {
	assert := assert.New(suite.T())
	require := require.New(suite.T())

	// Create a patient that needs to be discussed every week, to ensure weekly scheduled huddles
	suite.storePatientAndScores(bsonID(1), 8)
	// Create a patient that needs to be discussed every 2 weeks
	suite.storePatientAndScores(bsonID(2), 7)
	// Create a patient that is low-risk and doesn't generally need discussion, whom we'll manually schedule
	suite.storePatientAndScores(bsonID(3), 1)

	t := time.Now()
	config := createHuddleConfig(true, true, 0, t.Weekday())
	// The bug indicates there must be a huddle TODAY, so set the day of the week to today's
	config.Days[0] = time.Now().Weekday()

	// Schedule the huddles
	huddles, err := ScheduleHuddles(config)
	require.NoError(err)
	assert.Len(huddles, 4)
	for i := range huddles {
		ha := NewHuddleAssertions(huddles[i], assert)
		ha.AssertValidHuddleProfile()
		ha.AssertLeaderIDEqual("123")
		ha.AssertNameEqual("Test Huddle Config")
	}

	// Find today @ midnight to start checking dates
	t = time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())

	// Check each one individually to ensure it's as expected
	ha := NewHuddleAssertions(huddles[0], assert)
	ha.AssertActiveDateTimeEqual(t)
	ha.AssertMemberIDs(bsonID(1), bsonID(2))

	ha = NewHuddleAssertions(huddles[1], assert)
	ha.AssertActiveDateTimeEqual(t.AddDate(0, 0, 7))
	ha.AssertMemberIDs(bsonID(1))

	ha = NewHuddleAssertions(huddles[2], assert)
	ha.AssertActiveDateTimeEqual(t.AddDate(0, 0, 14))
	ha.AssertMemberIDs(bsonID(1), bsonID(2))

	ha = NewHuddleAssertions(huddles[3], assert)
	ha.AssertActiveDateTimeEqual(t.AddDate(0, 0, 21))
	ha.AssertMemberIDs(bsonID(1))

	// Now manually schedule patient 3 to today's huddle
	huddles[0].Member = append(huddles[0].Member, models.GroupMemberComponent{
		BackboneElement: models.BackboneElement{
			Element: models.Element{
				Extension: []models.Extension{
					{
						Url:                  "http://interventionengine.org/fhir/extension/group/member/reason",
						ValueCodeableConcept: manualAdditionReason("Why not?"),
					},
				},
			},
		},
		Entity: &models.Reference{
			Reference:    "Patient/" + bsonID(3),
			ReferencedID: bsonID(3),
			Type:         "Patient",
			External:     new(bool),
		},
	})
	// This bug was actually triggered because the UI sets the huddle time to midnight -- so do that here too!
	for _, ext := range huddles[0].Extension {
		if ext.Url == "http://interventionengine.org/fhir/extension/group/activeDateTime" {
			oldTime := ext.ValueDateTime.Time
			ext.ValueDateTime.Time = time.Date(oldTime.Year(), oldTime.Month(), oldTime.Day(), 0, 0, 0, 0, time.Local)
		}
	}
	err = suite.DB().C("groups").UpdateId(huddles[0].Id, huddles[0])
	require.NoError(err)

	// Reschedule the huddles
	huddles, err = ScheduleHuddles(config)
	require.NoError(err)
	assert.Len(huddles, 4)

	// Check each one again to ensure it's as expected
	ha = NewHuddleAssertions(huddles[0], assert)
	ha.AssertActiveDateTimeEqual(time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.Local))
	// Should have patient 3 (in addition to patient 1)
	ha.AssertMemberIDs(bsonID(3), bsonID(1))

	ha = NewHuddleAssertions(huddles[1], assert)
	ha.AssertActiveDateTimeEqual(t.AddDate(0, 0, 7))
	// Patent 2 comes first since he never had a huddle before
	ha.AssertMemberIDs(bsonID(2), bsonID(1))

	ha = NewHuddleAssertions(huddles[2], assert)
	ha.AssertActiveDateTimeEqual(t.AddDate(0, 0, 14))
	ha.AssertMemberIDs(bsonID(1))

	ha = NewHuddleAssertions(huddles[3], assert)
	ha.AssertActiveDateTimeEqual(t.AddDate(0, 0, 21))
	ha.AssertMemberIDs(bsonID(1), bsonID(2))

	// Now just make sure they were really stored to the db
	var storedHuddles []*models.Group
	server.Database.C("groups").Find(bson.M{}).Sort("extension.activeDateTime").All(&storedHuddles)
	assert.Equal(huddles, storedHuddles, "Stored huddles should match returned huddles")
}

func (suite *HuddleSchedulerSuite) TestRollOverPatientsToTodaysHuddle() {
	require := require.New(suite.T())

	lastHuddle := today().AddDate(0, 0, -3)
	ha := suite.setupAndStartRollOverTest(3, lastHuddle, today())

	// The three unreviewed patients (2, 3, and 5) should roll over to today's huddle -- but since 5 was scheduled
	// anyway, don't call it a rollover.
	require.Len(ha.Member, 4)
	ha.AssertMember(0, bsonID(2), &models.CodeableConcept{
		Coding: []models.Coding{
			models.Coding{System: "http://interventionengine.org/fhir/cs/huddle-member-reason", Code: "ROLLOVER"},
		},
		Text: fmt.Sprintf("Rolled Over from %s (Manually Added - I've got a hunch)", lastHuddle.Format("Jan 2")),
	})
	ha.AssertMember(1, bsonID(3), &models.CodeableConcept{
		Coding: []models.Coding{
			models.Coding{System: "http://interventionengine.org/fhir/cs/huddle-member-reason", Code: "ROLLOVER"},
		},
		Text: fmt.Sprintf("Rolled Over from %s (Risk Score Warrants Discussion)", lastHuddle.Format("Jan 2")),
	})
	ha.AssertMember(2, bsonID(4), riskScoreReason())
	ha.AssertMember(3, bsonID(5), riskScoreReason())
}

func (suite *HuddleSchedulerSuite) TestRollOverPatientsToNextHuddle() {
	require := require.New(suite.T())

	lastHuddle := today().AddDate(0, 0, -3)
	nextHuddle := today().AddDate(0, 0, 1)
	ha := suite.setupAndStartRollOverTest(3, lastHuddle, nextHuddle)

	// The three unreviewed patients (2, 3, and 5) should roll over to tomorrow's huddle -- but since 5 was scheduled
	// anyway, don't call it a rollover.
	require.Len(ha.Member, 4)
	ha.AssertMember(0, bsonID(2), &models.CodeableConcept{
		Coding: []models.Coding{
			models.Coding{System: "http://interventionengine.org/fhir/cs/huddle-member-reason", Code: "ROLLOVER"},
		},
		Text: fmt.Sprintf("Rolled Over from %s (Manually Added - I've got a hunch)", lastHuddle.Format("Jan 2")),
	})
	ha.AssertMember(1, bsonID(3), &models.CodeableConcept{
		Coding: []models.Coding{
			models.Coding{System: "http://interventionengine.org/fhir/cs/huddle-member-reason", Code: "ROLLOVER"},
		},
		Text: fmt.Sprintf("Rolled Over from %s (Risk Score Warrants Discussion)", lastHuddle.Format("Jan 2")),
	})
	ha.AssertMember(2, bsonID(4), riskScoreReason())
	ha.AssertMember(3, bsonID(5), riskScoreReason())
}

func (suite *HuddleSchedulerSuite) TestRollOverPatientsAlreadyInTodaysHuddle() {
	require := require.New(suite.T())

	// First setup an existing huddle today with one risk score patient and one rollover patient
	existingROFromDate := today().AddDate(0, 0, -5)
	suite.storePatientAndScores(bsonID(98), 3, 3, 2)
	suite.storePatientAndScores(bsonID(99), 3, 4, 4)
	reasonMap := map[string]*models.CodeableConcept{
		bsonID(99): rollOverReason(existingROFromDate, riskScoreReason()),
	}
	suite.storeHuddleWithDetails(today(), "123", riskScoreReason(), reasonMap, nil, bsonID(98), bsonID(99))

	// Then setup the rollover stuff
	lastHuddle := today().AddDate(0, 0, -3)
	ha := suite.setupAndStartRollOverTest(3, lastHuddle, today())

	// First, should be the previously rolled over patient (99), then new rollover patients (2, 3), then the risk score
	// patients (4, 5 -- no longer 98 since risk score went down to 2).
	require.Len(ha.Member, 5)
	ha.AssertMember(0, bsonID(99), &models.CodeableConcept{
		Coding: []models.Coding{
			models.Coding{System: "http://interventionengine.org/fhir/cs/huddle-member-reason", Code: "ROLLOVER"},
		},
		Text: fmt.Sprintf("Rolled Over from %s (Risk Score Warrants Discussion)", existingROFromDate.Format("Jan 2")),
	})
	ha.AssertMember(1, bsonID(2), &models.CodeableConcept{
		Coding: []models.Coding{
			models.Coding{System: "http://interventionengine.org/fhir/cs/huddle-member-reason", Code: "ROLLOVER"},
		},
		Text: fmt.Sprintf("Rolled Over from %s (Manually Added - I've got a hunch)", lastHuddle.Format("Jan 2")),
	})
	ha.AssertMember(2, bsonID(3), &models.CodeableConcept{
		Coding: []models.Coding{
			models.Coding{System: "http://interventionengine.org/fhir/cs/huddle-member-reason", Code: "ROLLOVER"},
		},
		Text: fmt.Sprintf("Rolled Over from %s (Risk Score Warrants Discussion)", lastHuddle.Format("Jan 2")),
	})
	ha.AssertMember(3, bsonID(4), riskScoreReason())
	ha.AssertMember(4, bsonID(5), riskScoreReason())
}

func (suite *HuddleSchedulerSuite) TestRollOverPatientsGetNewReasonIfApplicable() {
	require := require.New(suite.T())

	// First setup an existing huddle today with one risk score patient and one rollover patient
	existingROFromDate := today().AddDate(0, 0, -5)
	suite.storePatientAndScores(bsonID(98), 8, 8, 8)
	suite.storePatientAndScores(bsonID(99), 7, 8, 8)
	reasonMap := map[string]*models.CodeableConcept{
		bsonID(99): rollOverReason(existingROFromDate, riskScoreReason()),
	}
	suite.storeHuddleWithDetails(today(), "123", riskScoreReason(), reasonMap, nil, bsonID(98), bsonID(99))

	// Then setup the rollover stuff
	lastHuddle := today().AddDate(0, 0, -3)
	ha := suite.setupAndStartRollOverTest(3, lastHuddle, today())

	// The previously rolled over patient should not show RO as a reason since Risk Score brings him up again anyway
	require.Len(ha.Member, 6)
	ha.AssertMember(0, bsonID(2), &models.CodeableConcept{
		Coding: []models.Coding{
			models.Coding{System: "http://interventionengine.org/fhir/cs/huddle-member-reason", Code: "ROLLOVER"},
		},
		Text: fmt.Sprintf("Rolled Over from %s (Manually Added - I've got a hunch)", lastHuddle.Format("Jan 2")),
	})
	ha.AssertMember(1, bsonID(3), &models.CodeableConcept{
		Coding: []models.Coding{
			models.Coding{System: "http://interventionengine.org/fhir/cs/huddle-member-reason", Code: "ROLLOVER"},
		},
		Text: fmt.Sprintf("Rolled Over from %s (Risk Score Warrants Discussion)", lastHuddle.Format("Jan 2")),
	})
	ha.AssertMember(2, bsonID(98), riskScoreReason())
	ha.AssertMember(3, bsonID(99), riskScoreReason())
	ha.AssertMember(4, bsonID(4), riskScoreReason())
	ha.AssertMember(5, bsonID(5), riskScoreReason())
}

func (suite *HuddleSchedulerSuite) TestRollOverPatientsWithNoRollOverConfigured() {
	require := require.New(suite.T())

	ha := suite.setupAndStartRollOverTest(0, today().AddDate(0, 0, -3), today())

	// Since no rollover is configured, no rollover is expected
	require.Len(ha.Member, 2)
	ha.AssertMember(0, bsonID(4), riskScoreReason())
	ha.AssertMember(1, bsonID(5), riskScoreReason())
}

func (suite *HuddleSchedulerSuite) TestRollOverPatientsNotReadyToRoll() {
	require := require.New(suite.T())

	ha := suite.setupAndStartRollOverTest(4, today().AddDate(0, 0, -3), today())

	// Since configuration says there is a 4-day delay, and the huddle was only 3 days ago, no one is rolled over
	require.Len(ha.Member, 2)
	ha.AssertMember(0, bsonID(4), riskScoreReason())
	ha.AssertMember(1, bsonID(5), riskScoreReason())
}

func (suite *HuddleSchedulerSuite) TestRollOverPatientsTooManyDaysBack() {
	require := require.New(suite.T())

	ha := suite.setupAndStartRollOverTest(2, today().AddDate(0, 0, -3), today())

	// Since configuration says there is a 2-day delay, and the huddle was 4 days ago, no one is rolled over.
	// This is partially an implementation for simplicity, but it works under the assumption that they should
	// have already been rolled over
	require.Len(ha.Member, 2)
	ha.AssertMember(0, bsonID(4), riskScoreReason())
	ha.AssertMember(1, bsonID(5), riskScoreReason())
}

func (suite *HuddleSchedulerSuite) setupAndStartRollOverTest(rollOverDelay int, lastHuddleDate, huddleToTestDate time.Time) *HuddleAssertions {
	assert := assert.New(suite.T())
	require := require.New(suite.T())

	// PATIENT                                      // Hex | FREQ. |
	suite.storePatientAndScores(bsonID(1), 1, 1, 1) // 01  | never |
	suite.storePatientAndScores(bsonID(2), 5, 5, 4) // 02  | 4     |
	suite.storePatientAndScores(bsonID(3), 8, 8, 7) // 03  | 2     |
	suite.storePatientAndScores(bsonID(4), 8, 9, 8) // 04  | 1     |
	suite.storePatientAndScores(bsonID(5), 8, 9, 8) // 05  | 1     |
	suite.storePatientAndScores(bsonID(6), 8, 8, 7) // 06  | 2     |

	config := createHuddleConfig(true, true, rollOverDelay, lastHuddleDate.Weekday(), huddleToTestDate.Weekday())

	// Store the last huddle with five patients, two of whom were reviewed (it was a tough week)
	reasonMap := map[string]*models.CodeableConcept{
		bsonID(2): manualAdditionReason("I've got a hunch"),
	}
	reviewedMap := map[string]time.Time{
		bsonID(4): lastHuddleDate,
		bsonID(6): lastHuddleDate,
	}
	suite.storeHuddleWithDetails(lastHuddleDate, config.LeaderID, riskScoreReason(), reasonMap, reviewedMap, bsonID(2), bsonID(3), bsonID(4), bsonID(5), bsonID(6))

	groups, err := ScheduleHuddles(config)
	require.NoError(err)

	// First check them all for common huddle stuff
	require.Len(groups, 4)
	for _, g := range groups {
		ha := NewHuddleAssertions(g, assert)
		ha.AssertValidHuddleProfile()
		ha.AssertLeaderIDEqual("123")
		assert.True(strings.HasPrefix(ha.Name, "Test Huddle"))
	}

	// Now check the first one to ensure it's the right date
	ha := NewHuddleAssertions(groups[0], assert)
	ha.AssertActiveDateTimeEqual(huddleToTestDate)

	// And check all the others to make sure they *don't* have rollover patients
	for i := 1; i < 4; i++ {
		for _, member := range groups[i].Member {
			hm := HuddleMember(member)
			assert.False(hm.ReasonIsRollOver())
		}
	}

	return ha
}

func (suite *HuddleSchedulerSuite) TestInProgressHuddleIsntOverwritten() {
	assert := assert.New(suite.T())
	require := require.New(suite.T())

	// PATIENT                                      // Hex | FREQ.   | LAST HUDDLE | DUE BY
	suite.storePatientAndScores(bsonID(1), 1, 1, 1) // 01  | never   |             |
	suite.storePatientAndScores(bsonID(2), 5, 5, 4) // 02  | 4 weeks |             | 3 huddles from today
	suite.storePatientAndScores(bsonID(3), 8, 8, 7) // 03  | 2 weeks | 3 days ago  | 1 huddle from today
	suite.storePatientAndScores(bsonID(4), 8, 9, 8) // 04  | 1 week  | 3 days ago  | today, then 1 huddle from today
	suite.storePatientAndScores(bsonID(5), 8, 9, 8) // 05  | 1 week  | 3 days ago  | today, then 1 huddle from today
	suite.storePatientAndScores(bsonID(6), 8, 8, 5) // 06  | 4 weeks | 3 days ago  | 3 huddles from today

	lastHuddleDate := today().AddDate(0, 0, -3)
	config := createHuddleConfig(true, true, 3, lastHuddleDate.Weekday(), today().Weekday())

	// Store a huddle 3 days ago with three patients, two of whom were reviewed
	reasonMap := map[string]*models.CodeableConcept{
		bsonID(4): manualAdditionReason("I've got a hunch"),
	}
	reviewedMap := map[string]time.Time{
		bsonID(3): lastHuddleDate,
		bsonID(5): lastHuddleDate,
	}
	suite.storeHuddleWithDetails(lastHuddleDate, config.LeaderID, riskScoreReason(), reasonMap, reviewedMap, bsonID(3), bsonID(4), bsonID(5), bsonID(6))

	// Store a huddle today with two patients, one of whom was reviewed -- this is our IN PROGRESS huddle
	reviewedMap = map[string]time.Time{
		bsonID(2): today(),
	}
	suite.storeHuddleWithDetails(today(), config.LeaderID, riskScoreReason(), nil, reviewedMap, bsonID(1), bsonID(2))

	// Schedule the huddles
	groups, err := ScheduleHuddles(config)
	require.NoError(err)

	// First check them all for common huddle stuff
	require.Len(groups, 4)
	for _, g := range groups {
		ha := NewHuddleAssertions(g, assert)
		ha.AssertValidHuddleProfile()
		ha.AssertLeaderIDEqual("123")
		assert.True(strings.HasPrefix(ha.Name, "Test Huddle"))
	}

	// Now check the first one to ensure it's the right date (today) -- but hasn't been modified
	ha := NewHuddleAssertions(groups[0], assert)
	ha.AssertActiveDateTimeEqual(today())
	require.Len(ha.Member, 2)
	ha.AssertMember(0, bsonID(1), riskScoreReason())
	ha.AssertMember(1, bsonID(2), riskScoreReason())
	assert.Equal(today(), ha.FindHuddleMember(bsonID(2)).Reviewed().Time)

	// Now check the one after today to ensure it's as expected
	ha = NewHuddleAssertions(groups[1], assert)
	ha.AssertActiveDateTimeEqual(today().AddDate(0, 0, 4))
	require.Len(ha.Member, 4)
	ha.AssertMember(0, bsonID(6), &models.CodeableConcept{
		Coding: []models.Coding{
			models.Coding{System: "http://interventionengine.org/fhir/cs/huddle-member-reason", Code: "ROLLOVER"},
		},
		Text: fmt.Sprintf("Rolled Over from %s (Risk Score Warrants Discussion)", lastHuddleDate.Format("Jan 2")),
	})
	ha.AssertMember(1, bsonID(4), riskScoreReason())
	ha.AssertMember(2, bsonID(5), riskScoreReason())
	ha.AssertMember(3, bsonID(3), riskScoreReason())

	// And check all the others to make sure they *don't* have rollover patients
	for i := 2; i < 4; i++ {
		for _, member := range groups[i].Member {
			hm := HuddleMember(member)
			assert.False(hm.ReasonIsRollOver())
		}
	}
}

// This test ensures the huddle balancing algorithms work correctly when there are no previous huddles.
// This needs to test a few things:
// - A "strict" huddle adheres to the rules exactly (ideal==min==max)
// - A "flexible" huddle still adheres to the rules (min <= ideal <= max)
// - The "flexible" huddle is well-balanced compared to the "strict" huddle
func (suite *HuddleSchedulerSuite) TestHuddleBalancingWithNoPreviousHuddles() {
	assert := assert.New(suite.T())
	require := require.New(suite.T())

	// First create a bunch of patients to be scheduled into the huddles.
	// Generate 10 w/score 10, 20 w/score 9, 30 w/score 8, etc., for total of 550 patients.
	// Also keep track of patients' scores so we know if the huddle scheduling follows the rules
	scoreMap := make(map[string]int)
	id := 1
	num := 10
	for score := 10; score > 0; score-- {
		for i := 0; i < num; i++ {
			suite.storePatientAndScores(bsonID(id), score)
			scoreMap[bsonID(id)] = score
			id++
		}
		num += 10
	}

	// Schedule the huddles using strict configuration.
	// Since there are no past huddles, and new patients have built-in flexible scheduling, this isn't so bad.
	cfg := createHuddleConfigForBalanceTests(true)
	groups, err := ScheduleHuddles(cfg)
	require.NoError(err)
	suite.checkHuddleRules(groups, scoreMap, cfg)
	strictStdDev := suite.getStdDevForHuddles(groups)

	// Schedule the huddles using a flexible configuration.  Slightly better than the strict one.
	cfg = createHuddleConfigForBalanceTests(false)
	groups, err = ScheduleHuddles(cfg)
	require.NoError(err)
	suite.checkHuddleRules(groups, scoreMap, cfg)
	flexibleStdDev := suite.getStdDevForHuddles(groups)

	assert.True(flexibleStdDev < strictStdDev, "The flexible standard deviation should be smaller than the strict one")
}

// This test ensures the huddle balancing algorithms work correctly with a previous huddle representing a worst-case
// scenario: it contains every patient.  This needs to test a few things:
// - A "strict" huddle adheres to the rules exactly (ideal==min==max)
// - A "flexible" huddle still adheres to the rules (min <= ideal <= max)
// - The "flexible" huddle is well-balanced compared to the "strict" huddle
func (suite *HuddleSchedulerSuite) TestHuddleBalancingWithWorstCasePreviousHuddle() {
	assert := assert.New(suite.T())
	require := require.New(suite.T())

	// First create a bunch of patients to be scheduled into the huddles.
	// Generate 10 w/score 10, 20 w/score 9, 30 w/score 8, etc., for total of 550 patients.
	// Also keep track of patients' scores so we know if the huddle scheduling follows the rules
	var patientIDs []string
	scoreMap := make(map[string]int)
	id := 1
	num := 10
	for score := 10; score > 0; score-- {
		for i := 0; i < num; i++ {
			suite.storePatientAndScores(bsonID(id), score)
			patientIDs = append(patientIDs, bsonID(id))
			scoreMap[bsonID(id)] = score
			id++
		}
		num += 10
	}

	cfg := createHuddleConfigForBalanceTests(true)

	// Create a massive PAST huddle, just to throw off the numbers.  This basically represents the worst case scenario.
	suite.storeHuddle(today().AddDate(0, 0, -2), cfg.LeaderID, riskScoreReason(), patientIDs...)

	// Schedule the huddles using strict configuration.
	// Since there was just a massive huddle with everyone, this will be BAD for strict mode.
	// We will still follow the rules but it will be at expense of huddle distribution.
	groups, err := ScheduleHuddles(cfg)
	require.NoError(err)
	suite.checkHuddleRules(groups, scoreMap, cfg)
	strictStdDev := suite.getStdDevForHuddles(groups)

	// Schedule the huddles using a flexible configuration.  This should be much more evenly distributed.
	cfg = createHuddleConfigForBalanceTests(false)
	groups, err = ScheduleHuddles(cfg)
	require.NoError(err)
	suite.checkHuddleRules(groups, scoreMap, cfg)
	flexibleStdDev := suite.getStdDevForHuddles(groups)

	assert.True(flexibleStdDev < strictStdDev, "The flexible standard deviation should be smaller than the strict one")
}

func createHuddleConfigForBalanceTests(strict bool) *HuddleConfig {
	cfg := &HuddleConfig{
		Name:      "Test Huddle Config",
		LeaderID:  "123",
		Days:      []time.Weekday{time.Monday, time.Wednesday, time.Friday},
		LookAhead: 20,
		RiskConfig: &ScheduleByRiskConfig{
			RiskMethod: models.Coding{System: "http://interventionengine.org/risk-assessments", Code: "Test"},
			FrequencyConfigs: []RiskScoreFrequencyConfig{
				{
					MinScore:       10,
					MaxScore:       10,
					IdealFrequency: 1,
					MinFrequency:   1,
					MaxFrequency:   1,
				},
				{
					MinScore:       8,
					MaxScore:       9,
					IdealFrequency: 3,
					MinFrequency:   1,
					MaxFrequency:   4,
				},
				{
					MinScore:       6,
					MaxScore:       7,
					IdealFrequency: 6,
					MinFrequency:   3,
					MaxFrequency:   8,
				},
				{
					MinScore:       3,
					MaxScore:       5,
					IdealFrequency: 12,
					MinFrequency:   8,
					MaxFrequency:   15,
				},
			},
		},
	}

	if strict {
		for i := range cfg.RiskConfig.FrequencyConfigs {
			cfg.RiskConfig.FrequencyConfigs[i].MinFrequency = cfg.RiskConfig.FrequencyConfigs[i].IdealFrequency
			cfg.RiskConfig.FrequencyConfigs[i].MaxFrequency = cfg.RiskConfig.FrequencyConfigs[i].IdealFrequency
		}
	}

	return cfg
}

func (suite *HuddleSchedulerSuite) checkHuddleRules(groups []*models.Group, scoreMap map[string]int, cfg *HuddleConfig) {
	assert := assert.New(suite.T())
	require := require.New(suite.T())

	// Need to keep track of patients' last huddles so we know if the huddle scheduling follows the rules
	lastHuddleMap := make(map[string]*int)

	// Check each huddle to ensure it follows the rules
	require.Len(groups, 20)
	for i := range groups {
		ha := NewHuddleAssertions(groups[i], assert)
		ha.AssertValidHuddleProfile()
		ha.AssertLeaderIDEqual("123")
		assert.True(strings.HasPrefix(ha.Name, "Test Huddle"))
		for _, mem := range ha.HuddleMembers() {
			lastHuddle := lastHuddleMap[mem.ID()]
			frqCfg := cfg.FindRiskScoreFrequencyConfigByScore(float64(scoreMap[mem.ID()]))
			require.NotNil(frqCfg, "Patients with no configured frequency config should NOT be scheduled!")
			if lastHuddle == nil {
				// Special case, always allows for *some* flexibility
				assert.True(i < frqCfg.MaxFrequency, "Detected first huddle more than max frequency")
			} else {
				delta := i - *lastHuddle
				assert.True(delta >= frqCfg.MinFrequency, "Detected frequency less than min frequency")
				assert.True(delta <= frqCfg.MaxFrequency, "Detected frequency more than max frequency")
			}
			// Update the lastHuddle map
			// need to copy over the value of i before we store it, since i memory is re-used
			iPrime := i
			lastHuddleMap[mem.ID()] = &iPrime
		}
	}

	// Now check each patient that should be scheduled to ensure they were
	for id, score := range scoreMap {
		if score >= 3 && score <= 10 {
			assert.NotNil(lastHuddleMap[id])
		} else {
			assert.Nil(lastHuddleMap[id])
		}
	}
}

func (suite *HuddleSchedulerSuite) getStdDevForHuddles(groups []*models.Group) float64 {
	// Make a slice of the group sizes
	total := 0.0
	sizes := make([]int, len(groups))
	for i := range groups {
		sizes[i] = len(groups[i].Member)
		total += float64(sizes[i])
	}

	// Calculate the mean group size
	mean := total / float64(len(groups))

	// Now calculate the stddev
	total = 0.0
	for _, size := range sizes {
		total += math.Pow(float64(size)-mean, 2)
	}
	variance := total / float64(len(groups)-1)
	return math.Sqrt(variance)
}

func createHuddleConfig(inclRisk bool, inclEvents bool, rollOverDelay int, days ...time.Weekday) *HuddleConfig {
	c := new(HuddleConfig)
	c.Name = "Test Huddle Config"
	c.LeaderID = "123"
	c.Days = days
	c.LookAhead = 4
	if inclRisk {
		c.RiskConfig = new(ScheduleByRiskConfig)
		c.RiskConfig.RiskMethod = models.Coding{System: "http://interventionengine.org/risk-assessments", Code: "Test"}
		c.RiskConfig.FrequencyConfigs = []RiskScoreFrequencyConfig{
			{
				MinScore:       8,
				MaxScore:       10,
				IdealFrequency: 1,
				MinFrequency:   1,
				MaxFrequency:   1,
			},
			{
				MinScore:       6,
				MaxScore:       7,
				IdealFrequency: 2,
				MinFrequency:   2,
				MaxFrequency:   2,
			},
			{
				MinScore:       3,
				MaxScore:       5,
				IdealFrequency: 4,
				MinFrequency:   4,
				MaxFrequency:   4,
			},
		}
	}
	if inclEvents {
		c.EventConfig = new(ScheduleByEventConfig)
		c.EventConfig.EncounterConfigs = []EncounterEventConfig{
			{
				LookBackDays: 7,
				TypeCodes: []EventCode{
					{
						Name:       "Hospital Discharge",
						System:     "http://test/cs",
						Code:       "HOSP",
						UseEndDate: true,
					},
					{
						Name:   "Hospital Admission",
						System: "http://test/cs",
						Code:   "HOSP",
					},
					{
						Name:   "Emergency Room Visit",
						System: "http://test/cs",
						Code:   "ER",
					},
				},
			},
		}
	}
	c.RollOverDelayInDays = rollOverDelay
	return c
}

// storePatientAndScores creates a new patient with the given ID and a set of risk
// scores.  The scores are assumed to be in chronological order, so the last score
// passed in is the MOST_RECENT.
func (suite *HuddleSchedulerSuite) storePatientAndScores(id string, scores ...int) {
	require := require.New(suite.T())

	p := new(models.Patient)
	p.Id = id
	require.NoError(suite.DB().C("patients").Insert(p))

	if len(scores) > 0 {
		day := 24 * time.Hour
		date := time.Date(2016, time.March, 21, 0, 0, 0, 0, time.UTC).Add(time.Duration(-7*len(scores)) * day)
		for i, score := range scores {
			ra := new(models.RiskAssessment)
			ra.Id = bson.NewObjectId().Hex()
			ra.Subject = &models.Reference{
				Reference:    "Patient/" + id,
				ReferencedID: id,
				Type:         "Patient",
				External:     new(bool),
			}
			ra.Date = &models.FHIRDateTime{Time: date, Precision: models.Timestamp}
			ra.Method = &models.CodeableConcept{
				Coding: []models.Coding{{System: "http://interventionengine.org/risk-assessments", Code: "Test"}},
				Text:   "Test Risk Assessment",
			}
			ra.Basis = []models.Reference{
				{Reference: "http://foo.org/pie/" + ra.Id},
			}
			scoreFlt := float64(score)
			ra.Prediction = []models.RiskAssessmentPredictionComponent{
				{
					ProbabilityDecimal: &scoreFlt,
					Outcome:            &models.CodeableConcept{Text: "Something Bad"},
				},
			}
			if i == len(scores)-1 {
				ra.Meta = new(models.Meta)
				ra.Meta.Tag = []models.Coding{
					{
						System: "http://interventionengine.org/tags/",
						Code:   "MOST_RECENT",
					},
				}
			}
			require.NoError(suite.DB().C("riskassessments").Insert(ra))
			date = date.Add(7 * 24 * time.Hour)
		}
	}
}

func (suite *HuddleSchedulerSuite) storeEncounter(patientID string, code string, startDate *time.Time, endDate *time.Time) *models.Encounter {
	enc := new(models.Encounter)
	enc.Id = bson.NewObjectId().Hex()
	enc.Status = "finished"
	enc.Type = []models.CodeableConcept{
		{
			Coding: []models.Coding{
				{System: "http://test/cs", Code: code},
			},
		},
	}
	enc.Patient = &models.Reference{
		Reference:    "Patient/" + patientID,
		ReferencedID: patientID,
		Type:         "Patient",
		External:     new(bool),
	}
	enc.Period = new(models.Period)
	if startDate != nil {
		enc.Period.Start = &models.FHIRDateTime{Time: *startDate, Precision: models.Timestamp}
	}
	if endDate != nil {
		enc.Period.End = &models.FHIRDateTime{Time: *endDate, Precision: models.Timestamp}
	}
	err := suite.DB().C("encounters").Insert(enc)
	suite.Require().NoError(err)
	return enc
}

func (suite *HuddleSchedulerSuite) storeHuddleWithDetails(date time.Time, leaderID string, defaultReason *models.CodeableConcept, reasonMap map[string]*models.CodeableConcept, reviewedMap map[string]time.Time, patients ...string) {
	require := require.New(suite.T())

	g := new(models.Group)
	g.Id = bson.NewObjectId().Hex()
	g.Meta = &models.Meta{
		Profile: []string{"http://interventionengine.org/fhir/profile/huddle"},
	}
	g.Extension = []models.Extension{
		{
			Url:           "http://interventionengine.org/fhir/extension/group/activeDateTime",
			ValueDateTime: &models.FHIRDateTime{Time: date, Precision: models.Precision(models.Date)},
		},
		{
			Url: "http://interventionengine.org/fhir/extension/group/leader",
			ValueReference: &models.Reference{
				Reference:    "Practitioner/" + leaderID,
				ReferencedID: leaderID,
				Type:         "Practitioner",
				External:     new(bool),
			},
		},
	}
	g.Type = "person"
	tru := true
	g.Actual = &tru
	g.Code = &models.CodeableConcept{
		Coding: []models.Coding{
			{System: "http://interventionengine.org/fhir/cs/huddle", Code: "HUDDLE"},
		},
		Text: "Huddle",
	}
	g.Name = "Test Huddle " + g.Id
	g.Member = make([]models.GroupMemberComponent, len(patients))
	for i := range patients {
		g.Member[i] = models.GroupMemberComponent{}
		reason, found := reasonMap[patients[i]]
		if !found {
			reason = defaultReason
		}
		g.Member[i].Extension = []models.Extension{
			{
				Url:                  "http://interventionengine.org/fhir/extension/group/member/reason",
				ValueCodeableConcept: reason,
			},
		}
		if reviewed, found := reviewedMap[patients[i]]; found {
			g.Member[i].Extension = append(g.Member[i].Extension, models.Extension{
				Url:           "http://interventionengine.org/fhir/extension/group/member/reviewed",
				ValueDateTime: &models.FHIRDateTime{Time: reviewed, Precision: models.Timestamp},
			})
		}
		g.Member[i].Entity = &models.Reference{
			Reference:    "Patient/" + patients[i],
			ReferencedID: patients[i],
			Type:         "Patient",
			External:     new(bool),
		}
	}

	require.NoError(suite.DB().C("groups").Insert(g))
}

func (suite *HuddleSchedulerSuite) storeHuddle(date time.Time, leaderID string, reason *models.CodeableConcept, patients ...string) {
	suite.storeHuddleWithDetails(date, leaderID, reason, nil, nil, patients...)
}

func riskScoreReason() *models.CodeableConcept {
	return &models.CodeableConcept{
		Coding: []models.Coding{
			{System: "http://interventionengine.org/fhir/cs/huddle-member-reason", Code: "RISK_SCORE"},
		},
		Text: "Risk Score Warrants Discussion",
	}
}

func recentEncounterReason(description string) *models.CodeableConcept {
	return &models.CodeableConcept{
		Coding: []models.Coding{
			{System: "http://interventionengine.org/fhir/cs/huddle-member-reason", Code: "RECENT_ENCOUNTER"},
		},
		Text: description,
	}
}

func manualAdditionReason(description string) *models.CodeableConcept {
	return &models.CodeableConcept{
		Coding: []models.Coding{
			{System: "http://interventionengine.org/fhir/cs/huddle-member-reason", Code: "MANUAL_ADDITION"},
		},
		Text: description,
	}
}

func rollOverReason(from time.Time, previousReason *models.CodeableConcept) *models.CodeableConcept {
	var reason string
	if previousReason.MatchesCode("http://interventionengine.org/fhir/cs/huddle-member-reason", "ROLLOVER") {
		reason = previousReason.Text
	} else if previousReason.MatchesCode("http://interventionengine.org/fhir/cs/huddle-member-reason", "MANUAL_ADDITION") {
		reason = fmt.Sprintf("Rolled Over from %s (Manually Added - %s)", from.Format("Jan 2"), previousReason.Text)
	} else {
		reason = fmt.Sprintf("Rolled Over from %s (%s)", from.Format("Jan 2"), previousReason.Text)
	}
	return &models.CodeableConcept{
		Coding: []models.Coding{
			{System: "http://interventionengine.org/fhir/cs/huddle-member-reason", Code: "ROLLOVER"},
		},
		Text: reason,
	}
}

func bsonID(id int) string {
	return fmt.Sprintf("%024x", id)
}

type HuddleAssertions struct {
	*ie.Huddle
	assert *assert.Assertions
}

func NewHuddleAssertions(group *models.Group, assertions *assert.Assertions) *HuddleAssertions {
	h := Huddle(*group)
	return &HuddleAssertions{
		Huddle: &h,
		assert: assertions,
	}
}

func (h *HuddleAssertions) AssertValidHuddleProfile() bool {
	status := true

	status = status && h.assert.Contains(h.Meta.Profile, "http://interventionengine.org/fhir/profile/huddle")
	status = status && h.assert.NotNil(h.ActiveDateTime())
	l := h.Leader()
	status = status && h.assert.NotNil(l)
	status = status && h.assert.Equal("Practitioner", l.Type)
	status = status && h.assert.False(*l.External)
	status = status && h.assert.Equal("person", h.Type)
	status = status && h.assert.True(*h.Actual)
	status = status && h.assert.Equal(&models.CodeableConcept{
		Coding: []models.Coding{
			{System: "http://interventionengine.org/fhir/cs/huddle", Code: "HUDDLE"},
		},
		Text: "Huddle",
	}, h.Code)
	status = status && h.assert.NotEmpty(h.Name)
	for _, mem := range h.HuddleMembers() {
		r := mem.Reason()
		status = status && h.assert.NotNil(r)
		status = status && h.assert.Len(r.Coding, 1)
		status = status && h.assert.Equal("http://interventionengine.org/fhir/cs/huddle-member-reason", r.Coding[0].System)
		status = status && h.assert.NotEmpty(r.Text)
		status = status && h.assert.NotNil(mem.Entity)
		status = status && h.assert.Equal("Patient", mem.Entity.Type)
		status = status && h.assert.False(*mem.Entity.External)
	}

	return status
}

func (h *HuddleAssertions) AssertActiveDateTimeEqual(date time.Time) bool {
	adt := h.ActiveDateTime()
	status := true
	status = status && h.assert.NotNil(adt)
	return status && h.assert.True(adt.Time.Equal(date), "Expected <", date, ">, Got <", adt.Time, ">")
}

func (h *HuddleAssertions) AssertLeaderIDEqual(leaderID string) bool {
	leader := h.Leader()
	status := true
	status = status && h.assert.NotNil(leader)
	return status && h.assert.Equal(&models.Reference{
		Reference:    "Practitioner/" + leaderID,
		ReferencedID: leaderID,
		Type:         "Practitioner",
		External:     new(bool),
	}, leader)
}

func (h *HuddleAssertions) AssertNameEqual(name string) bool {
	return h.assert.Equal(name, h.Name)
}

func (h *HuddleAssertions) AssertMember(i int, id string, reason *models.CodeableConcept) bool {
	status := h.assert.True(i < len(h.Member), "Not enough members: ", len(h.Member))
	if !status {
		return status
	}
	status = status && h.assert.Equal(&models.Reference{
		Reference:    "Patient/" + id,
		ReferencedID: id,
		Type:         "Patient",
		External:     new(bool),
	}, h.Member[i].Entity)
	if reason != nil {
		status = status && h.assert.Equal(reason, h.HuddleMembers()[i].Reason())
	}
	return status
}

func (h *HuddleAssertions) AssertMemberIDs(id ...string) bool {
	h.assert.Len(h.Member, len(id))
	status := true
	for i := range h.Member {
		status = status && h.assert.Equal(&models.Reference{
			Reference:    "Patient/" + id[i],
			ReferencedID: id[i],
			Type:         "Patient",
			External:     new(bool),
		}, h.Member[i].Entity)
	}
	return status
}
