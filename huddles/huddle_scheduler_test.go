package huddles

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/intervention-engine/fhir/models"
	"github.com/intervention-engine/fhir/server"
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

// For now, this test just uses risk scores, but it should be expanded as we add support for autopopulating
// based on other criteria.
func (suite *HuddleSchedulerSuite) TestCreatePopulatedHuddleForNewHuddle() {
	assert := assert.New(suite.T())
	require := require.New(suite.T())

	// PATIENT                                        // Hex | FREQ.   | LAST HUDDLE        | DUE BY
	suite.storePatientAndScores(bsonID(1), 1, 1, 1)   // 01  | never   |                    |
	suite.storePatientAndScores(bsonID(2), 1, 2, 2)   // 02  | never   |                    |
	suite.storePatientAndScores(bsonID(3), 2, 1, 1)   // 03  | never   |                    |
	suite.storePatientAndScores(bsonID(4), 1, 2, 1)   // 04  | never   |                    |
	suite.storePatientAndScores(bsonID(5), 2, 3, 2)   // 05  | never   | 7 weeks ago (2/01) |
	suite.storePatientAndScores(bsonID(6), 1, 2, 3)   // 06  | 4 weeks | 1 week ago  (3/14) | 4/11
	suite.storePatientAndScores(bsonID(7), 6, 6, 5)   // 07  | 4 weeks | 7 weeks ago (2/01) | 3/21 (overdue)
	suite.storePatientAndScores(bsonID(8), 5, 5, 4)   // 08  | 4 weeks |                    | 3/21 (overdue)
	suite.storePatientAndScores(bsonID(9), 3, 3, 3)   // 09  | 4 weeks | 2 weeks ago (3/07) | 4/04
	suite.storePatientAndScores(bsonID(10), 4, 5, 4)  // 0a  | 4 weeks | 4 weeks ago (2/22) | 3/21
	suite.storePatientAndScores(bsonID(11), 5, 4, 6)  // 0b  | 2 weeks | 1 week ago  (3/14) | 3/28
	suite.storePatientAndScores(bsonID(12), 5, 6, 6)  // 0c  | 2 weeks | 3 weeks ago (2/29) | 3/21 (overdue)
	suite.storePatientAndScores(bsonID(13), 6, 7, 6)  // 0d  | 2 weeks |                    | 3/21 (overdue)
	suite.storePatientAndScores(bsonID(14), 8, 8, 7)  // 0e  | 2 weeks | 1 week ago  (3/14) | 3/28
	suite.storePatientAndScores(bsonID(15), 5, 6, 7)  // 0f  | 2 weeks | 2 weeks ago (3/07) | 3/21
	suite.storePatientAndScores(bsonID(16), 9, 9, 9)  // 10  | 1 week  |                    | 3/21 (overdue)
	suite.storePatientAndScores(bsonID(17), 8, 9, 8)  // 11  | 1 week  | 1 week ago  (3/14) | 3/21
	suite.storePatientAndScores(bsonID(18), 7, 7, 8)  // 12  | 1 week  | 2 weeks ago (3/07) | 3/21
	suite.storePatientAndScores(bsonID(19), 9, 9, 9)  // 13  | 1 week  | 1 week ago  (3/14) | 3/21
	suite.storePatientAndScores(bsonID(20), 9, 9, 10) // 14  | 1 week  | 1 week ago  (3/14) | 3/21

	config := createHuddleConfig()

	suite.storeHuddle(time.Date(2016, time.February, 1, 0, 0, 0, 0, time.UTC), config.LeaderID, &RiskScoreReason,
		bsonID(5), bsonID(7))
	suite.storeHuddle(time.Date(2016, time.February, 22, 0, 0, 0, 0, time.UTC), config.LeaderID, &RiskScoreReason,
		bsonID(10))
	suite.storeHuddle(time.Date(2016, time.February, 29, 0, 0, 0, 0, time.UTC), config.LeaderID, &RiskScoreReason,
		bsonID(12))
	suite.storeHuddle(time.Date(2016, time.March, 7, 0, 0, 0, 0, time.UTC), config.LeaderID, &RiskScoreReason,
		bsonID(9), bsonID(15), bsonID(18))
	suite.storeHuddle(time.Date(2016, time.March, 14, 0, 0, 0, 0, time.UTC), config.LeaderID, &RiskScoreReason,
		bsonID(6), bsonID(11), bsonID(14), bsonID(17), bsonID(19), bsonID(20))

	group, err := CreatePopulatedHuddle(time.Date(2016, time.March, 21, 9, 0, 0, 0, time.UTC), config)
	require.NoError(err)
	ha := NewHuddleAssertions(group, assert)
	ha.AssertValidHuddleProfile()
	ha.AssertActiveDateTimeEqual(time.Date(2016, time.March, 21, 9, 0, 0, 0, time.UTC))
	ha.AssertLeaderIdEqual("123")
	ha.AssertNameEqual("Test Huddle Config")
	assert.Len(group.Member, 11)

	expectedIDs := []string{bsonID(20), bsonID(16), bsonID(19), bsonID(18), bsonID(17), bsonID(15), bsonID(13),
		bsonID(12), bsonID(7), bsonID(8), bsonID(10)}
	for i, expectedID := range expectedIDs {
		ha.AssertMember(i, expectedID, &RiskScoreReason)
	}
}

// This test just ensures that manually added patients aren't overwritten by the huddle population
// algorithm.
func (suite *HuddleSchedulerSuite) TestCreatePopulatedHuddleForExistingHuddle() {
	assert := assert.New(suite.T())
	require := require.New(suite.T())

	// PATIENT                                        // Hex | FREQ.   | LAST HUDDLE        | DUE BY
	suite.storePatientAndScores(bsonID(1), 1, 1, 1) // 01  | never   |                    |
	suite.storePatientAndScores(bsonID(2), 5, 5, 4) // 08  | 4 weeks |                    | 3/21 (overdue)
	suite.storePatientAndScores(bsonID(3), 8, 8, 7) // 0e  | 2 weeks | 1 week ago  (3/14) | 3/28
	suite.storePatientAndScores(bsonID(4), 8, 9, 8) // 11  | 1 week  | 1 week ago  (3/14) | 3/21

	config := createHuddleConfig()
	suite.storeHuddle(time.Date(2016, time.March, 14, 0, 0, 0, 0, time.Local), config.LeaderID, &RiskScoreReason, bsonID(3), bsonID(4))
	suite.storeHuddleWithReasons(time.Date(2016, time.March, 21, 0, 0, 0, 0, time.Local), config.LeaderID, &RiskScoreReason,
		map[string]*models.CodeableConcept{bsonID(1): &ManualAdditionReason}, bsonID(1), bsonID(2), bsonID(3), bsonID(4))

	group, err := CreatePopulatedHuddle(time.Date(2016, time.March, 21, 0, 0, 0, 0, time.Local), config)
	require.NoError(err)

	ha := NewHuddleAssertions(group, assert)
	ha.AssertValidHuddleProfile()
	ha.AssertActiveDateTimeEqual(time.Date(2016, time.March, 21, 0, 0, 0, 0, time.Local))
	ha.AssertLeaderIdEqual("123")
	assert.True(strings.HasPrefix(ha.Name, "Test Huddle"))

	require.Len(ha.Member, 3)
	ha.AssertMember(0, bsonID(1), &ManualAdditionReason)
	ha.AssertMember(1, bsonID(4), &RiskScoreReason)
	ha.AssertMember(2, bsonID(2), &RiskScoreReason)
}

func (suite *HuddleSchedulerSuite) TestScheduleHuddles() {
	assert := assert.New(suite.T())
	require := require.New(suite.T())

	suite.storePatientAndScores(bsonID(1), 5, 4, 7)  // every 2 weeks
	suite.storePatientAndScores(bsonID(2), 1, 1, 1)  // never
	suite.storePatientAndScores(bsonID(3), 1, 2, 3)  // every 4 weeks
	suite.storePatientAndScores(bsonID(4), 9, 9, 10) // every week
	suite.storePatientAndScores(bsonID(5), 7, 6, 5)  // every 4 weeks

	config := createHuddleConfig()
	huddles, err := ScheduleHuddles(config)
	require.NoError(err)
	assert.Len(huddles, 4)
	for i := range huddles {
		ha := NewHuddleAssertions(huddles[i], assert)
		ha.AssertValidHuddleProfile()
		ha.AssertLeaderIdEqual("123")
		ha.AssertNameEqual("Test Huddle Config")
	}

	// Find next Monday @ 10am to start checking dates
	t := time.Now()
	for ; t.Weekday() != time.Monday; t = t.AddDate(0, 0, 1) {
	}
	t = time.Date(t.Year(), t.Month(), t.Day(), 10, 0, 0, 0, t.Location())
	// Now check each one individually
	ha := NewHuddleAssertions(huddles[0], assert)
	ha.AssertActiveDateTimeEqual(t)
	ha.AssertMemberIDs(bsonID(4), bsonID(1), bsonID(5))

	ha = NewHuddleAssertions(huddles[1], assert)
	ha.AssertActiveDateTimeEqual(t.AddDate(0, 0, 7))
	ha.AssertMemberIDs(bsonID(4), bsonID(3))

	ha = NewHuddleAssertions(huddles[2], assert)
	ha.AssertActiveDateTimeEqual(t.AddDate(0, 0, 14))
	ha.AssertMemberIDs(bsonID(4), bsonID(1))

	ha = NewHuddleAssertions(huddles[3], assert)
	ha.AssertActiveDateTimeEqual(t.AddDate(0, 0, 21))
	ha.AssertMemberIDs(bsonID(4))

	// Now just make sure they were really stored to the db
	var storedHuddles []*models.Group
	server.Database.C("groups").Find(bson.M{}).Sort("extension.activeDateTime").All(&storedHuddles)
	assert.Equal(huddles, storedHuddles, "Stored huddles should match returned huddles")
}

// Test for bug logged in pivotal: https://www.pivotaltracker.com/story/show/117859291
func (suite *HuddleSchedulerSuite) TestManuallyAddPatientToExistingHuddle() {
	assert := assert.New(suite.T())
	require := require.New(suite.T())

	// Create a patient that needs to be discussed every week, to ensure weekly scheduled huddles
	suite.storePatientAndScores(bsonID(1), 8)
	// Create a patient that needs to be discussed every 2 weeks, whom we'll manually schedule on an off-week
	suite.storePatientAndScores(bsonID(2), 7)

	config := createHuddleConfig()
	// The bug indicates there must be a huddle TODAY, so set the day of the week to today's
	config.Days[0] = time.Now().Weekday()

	// Schedule the huddles
	huddles, err := ScheduleHuddles(config)
	require.NoError(err)
	assert.Len(huddles, 4)
	for i := range huddles {
		ha := NewHuddleAssertions(huddles[i], assert)
		ha.AssertValidHuddleProfile()
		ha.AssertLeaderIdEqual("123")
		ha.AssertNameEqual("Test Huddle Config")
	}

	// Find today @ 10am to start checking dates
	t := time.Now()
	t = time.Date(t.Year(), t.Month(), t.Day(), 10, 0, 0, 0, t.Location())

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

	// Now manually schedule patient 2 to the second huddle (which should be an off-week for the patient)
	huddles[1].Member = append(huddles[1].Member, models.GroupMemberComponent{
		BackboneElement: models.BackboneElement{
			Element: models.Element{
				Extension: []models.Extension{
					{
						Url:                  "http://interventionengine.org/fhir/extension/group/member/reason",
						ValueCodeableConcept: &ManualAdditionReason,
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
	err = suite.DB().C("groups").UpdateId(huddles[1].Id, huddles[1])
	require.NoError(err)

	// Reschedule the huddles
	huddles, err = ScheduleHuddles(config)
	require.NoError(err)
	assert.Len(huddles, 4)

	// Check each one again to ensure it's as expected
	ha = NewHuddleAssertions(huddles[0], assert)
	ha.AssertActiveDateTimeEqual(t)
	// Still should have patient 2 because it's before the new manual addition
	ha.AssertMemberIDs(bsonID(1), bsonID(2))

	ha = NewHuddleAssertions(huddles[1], assert)
	ha.AssertActiveDateTimeEqual(t.AddDate(0, 0, 7))
	// Should have patient 2 now because patient 2 was manually added for this week
	ha.AssertMemberIDs(bsonID(2), bsonID(1))

	ha = NewHuddleAssertions(huddles[2], assert)
	ha.AssertActiveDateTimeEqual(t.AddDate(0, 0, 14))
	// Should no longer have patient 2 because the manual addition reset the every-2-week cadence
	ha.AssertMemberIDs(bsonID(1))

	ha = NewHuddleAssertions(huddles[3], assert)
	ha.AssertActiveDateTimeEqual(t.AddDate(0, 0, 21))
	// Should have patient 2 now due to the new cadence
	ha.AssertMemberIDs(bsonID(1), bsonID(2))

	// Now just make sure they were really stored to the db
	var storedHuddles []*models.Group
	server.Database.C("groups").Find(bson.M{}).Sort("extension.activeDateTime").All(&storedHuddles)
	assert.Equal(huddles, storedHuddles, "Stored huddles should match returned huddles")
}

func (suite *HuddleSchedulerSuite) TestFindEligiblePatientIDsByRiskScoreWithNoPreviousHuddles() {
	assert := assert.New(suite.T())
	require := require.New(suite.T())

	suite.storePatientAndScores(bsonID(1), 5, 4, 7)  // every 2 weeks
	suite.storePatientAndScores(bsonID(2), 1, 1, 1)  // never
	suite.storePatientAndScores(bsonID(3), 1, 2, 3)  // every 4 weeks
	suite.storePatientAndScores(bsonID(4), 9, 9, 10) // every week
	suite.storePatientAndScores(bsonID(5), 7, 6, 5)  // every 4 weeks

	config := createHuddleConfig()

	eligibles, err := findEligiblePatientIDsByRiskScore(time.Date(2016, time.March, 21, 9, 0, 0, 0, time.UTC), config)
	require.NoError(err)
	assert.Len(eligibles, 3)
	assert.Contains(eligibles, bsonID(1))
	assert.NotContains(eligibles, bsonID(2))
	assert.NotContains(eligibles, bsonID(3))
	assert.Contains(eligibles, bsonID(4))
	assert.Contains(eligibles, bsonID(5))
}

func (suite *HuddleSchedulerSuite) TestFindEligiblePatientIDsByRiskScore() {
	assert := assert.New(suite.T())
	require := require.New(suite.T())

	// PATIENT                                        // Hex | FREQ.   | LAST HUDDLE        | DUE BY
	suite.storePatientAndScores(bsonID(1), 1, 1, 1)   // 01  | never   |                    |
	suite.storePatientAndScores(bsonID(2), 1, 2, 2)   // 02  | never   |                    |
	suite.storePatientAndScores(bsonID(3), 2, 1, 1)   // 03  | never   |                    |
	suite.storePatientAndScores(bsonID(4), 1, 2, 1)   // 04  | never   |                    |
	suite.storePatientAndScores(bsonID(5), 2, 3, 2)   // 05  | never   | 7 weeks ago (2/01) |
	suite.storePatientAndScores(bsonID(6), 1, 2, 3)   // 06  | 4 weeks | 1 week ago  (3/14) | 4/11
	suite.storePatientAndScores(bsonID(7), 6, 6, 5)   // 07  | 4 weeks | 7 weeks ago (2/01) | 3/21 (overdue)
	suite.storePatientAndScores(bsonID(8), 5, 5, 4)   // 08  | 4 weeks |                    | 3/21 (overdue)
	suite.storePatientAndScores(bsonID(9), 3, 3, 3)   // 09  | 4 weeks | 2 weeks ago (3/07) | 4/04
	suite.storePatientAndScores(bsonID(10), 4, 5, 4)  // 0a  | 4 weeks | 4 weeks ago (2/22) | 3/21
	suite.storePatientAndScores(bsonID(11), 5, 4, 6)  // 0b  | 2 weeks | 1 week ago  (3/14) | 3/28
	suite.storePatientAndScores(bsonID(12), 5, 6, 6)  // 0c  | 2 weeks | 3 weeks ago (2/29) | 3/21 (overdue)
	suite.storePatientAndScores(bsonID(13), 6, 7, 6)  // 0d  | 2 weeks |                    | 3/21 (overdue)
	suite.storePatientAndScores(bsonID(14), 8, 8, 7)  // 0e  | 2 weeks | 1 week ago  (3/14) | 3/28
	suite.storePatientAndScores(bsonID(15), 5, 6, 7)  // 0f  | 2 weeks | 2 weeks ago (3/07) | 3/21
	suite.storePatientAndScores(bsonID(16), 9, 9, 9)  // 10  | 1 week  |                    | 3/21 (overdue)
	suite.storePatientAndScores(bsonID(17), 8, 9, 8)  // 11  | 1 week  | 1 week ago  (3/14) | 3/21
	suite.storePatientAndScores(bsonID(18), 7, 7, 8)  // 12  | 1 week  | 2 weeks ago (3/07) | 3/21
	suite.storePatientAndScores(bsonID(19), 9, 9, 9)  // 13  | 1 week  | 1 week ago  (3/14) | 3/21
	suite.storePatientAndScores(bsonID(20), 9, 9, 10) // 14  | 1 week  | 1 week ago  (3/14) | 3/21

	config := createHuddleConfig()

	suite.storeHuddle(time.Date(2016, time.February, 1, 0, 0, 0, 0, time.UTC), config.LeaderID, &RiskScoreReason,
		bsonID(5), bsonID(7))
	suite.storeHuddle(time.Date(2016, time.February, 22, 0, 0, 0, 0, time.UTC), config.LeaderID, &RiskScoreReason,
		bsonID(10))
	suite.storeHuddle(time.Date(2016, time.February, 29, 0, 0, 0, 0, time.UTC), config.LeaderID, &RiskScoreReason,
		bsonID(12))
	suite.storeHuddle(time.Date(2016, time.March, 7, 0, 0, 0, 0, time.UTC), config.LeaderID, &RiskScoreReason,
		bsonID(9), bsonID(15), bsonID(18))
	suite.storeHuddle(time.Date(2016, time.March, 14, 0, 0, 0, 0, time.UTC), config.LeaderID, &RiskScoreReason,
		bsonID(6), bsonID(11), bsonID(14), bsonID(17), bsonID(19), bsonID(20))

	eligibles, err := findEligiblePatientIDsByRiskScore(time.Date(2016, time.March, 21, 9, 0, 0, 0, time.UTC), config)
	require.NoError(err)
	assert.Len(eligibles, 11)
	assert.NotContains(eligibles, bsonID(1))
	assert.NotContains(eligibles, bsonID(2))
	assert.NotContains(eligibles, bsonID(3))
	assert.NotContains(eligibles, bsonID(4))
	assert.NotContains(eligibles, bsonID(5))
	assert.NotContains(eligibles, bsonID(6))
	assert.Contains(eligibles, bsonID(7))
	assert.Contains(eligibles, bsonID(8))
	assert.NotContains(eligibles, bsonID(9))
	assert.Contains(eligibles, bsonID(10))
	assert.NotContains(eligibles, bsonID(11))
	assert.Contains(eligibles, bsonID(12))
	assert.Contains(eligibles, bsonID(13))
	assert.NotContains(eligibles, bsonID(14))
	assert.Contains(eligibles, bsonID(15))
	assert.Contains(eligibles, bsonID(16))
	assert.Contains(eligibles, bsonID(17))
	assert.Contains(eligibles, bsonID(18))
	assert.Contains(eligibles, bsonID(19))
	assert.Contains(eligibles, bsonID(20))

	// Go ahead and check the order.  Should be highest scores first -- with the tie breaker being how long its been
	// since the patient's last huddle.  No previous huddle at all gets highest priority in a tie-breaker.
	assert.Equal([]string{bsonID(20), bsonID(16), bsonID(19), bsonID(18), bsonID(17), bsonID(15), bsonID(13),
		bsonID(12), bsonID(7), bsonID(8), bsonID(10)}, eligibles)

}

func createHuddleConfig() *HuddleConfig {
	c := new(HuddleConfig)
	c.Name = "Test Huddle Config"
	c.LeaderID = "123"
	c.Days = []time.Weekday{time.Monday}
	c.LookAhead = 4
	c.RiskConfig = new(ScheduleByRiskConfig)
	c.RiskConfig.RiskMethod = models.Coding{System: "http://interventionengine.org/risk-assessments", Code: "Test"}
	c.RiskConfig.FrequencyConfigs = []RiskScoreFrequencyConfig{
		{
			MinScore:              8,
			MaxScore:              10,
			MinDaysBetweenHuddles: 5,
			MaxDaysBetweenHuddles: 7,
		},
		{
			MinScore:              6,
			MaxScore:              7,
			MinDaysBetweenHuddles: 12,
			MaxDaysBetweenHuddles: 14,
		},
		{
			MinScore:              3,
			MaxScore:              5,
			MinDaysBetweenHuddles: 25,
			MaxDaysBetweenHuddles: 28,
		},
	}
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

func (suite *HuddleSchedulerSuite) storeHuddleWithReasons(date time.Time, leaderID string, defaultReason *models.CodeableConcept, reasonMap map[string]*models.CodeableConcept, patients ...string) {
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
	suite.storeHuddleWithReasons(date, leaderID, reason, nil, patients...)
}

func bsonID(id int) string {
	return fmt.Sprintf("%024x", id)
}

type HuddleAssertions struct {
	*models.Group
	assert *assert.Assertions
}

func NewHuddleAssertions(group *models.Group, assertions *assert.Assertions) *HuddleAssertions {
	return &HuddleAssertions{
		Group:  group,
		assert: assertions,
	}
}

func (h *HuddleAssertions) AssertValidHuddleProfile() bool {
	status := true

	status = status && h.assert.Contains(h.Meta.Profile, "http://interventionengine.org/fhir/profile/huddle")
	var foundActiveDateTime, foundLeader bool
	for _, ext := range h.Extension {
		switch ext.Url {
		case "http://interventionengine.org/fhir/extension/group/activeDateTime":
			status = status && h.assert.NotNil(ext.ValueDateTime)
			foundActiveDateTime = true
		case "http://interventionengine.org/fhir/extension/group/leader":
			status = status && h.assert.NotNil(ext.ValueReference)
			status = status && h.assert.Equal("Practitioner", ext.ValueReference.Type)
			status = status && h.assert.False(*ext.ValueReference.External)
			foundLeader = true
		}
	}
	status = status && h.assert.True(foundActiveDateTime, "activeDateTime extension should be populated")
	status = status && h.assert.True(foundLeader, "leader extension should be populated")
	status = status && h.assert.Equal("person", h.Type)
	status = status && h.assert.True(*h.Actual)
	status = status && h.assert.Equal(&models.CodeableConcept{
		Coding: []models.Coding{
			{System: "http://interventionengine.org/fhir/cs/huddle", Code: "HUDDLE"},
		},
		Text: "Huddle",
	}, h.Code)
	status = status && h.assert.NotEmpty(h.Name)
	for _, mem := range h.Member {
		var foundReason bool
		for _, ext := range mem.Extension {
			if ext.Url == "http://interventionengine.org/fhir/extension/group/member/reason" {
				status = status && h.assert.NotNil(ext.ValueCodeableConcept)
				status = status && h.assert.Len(ext.ValueCodeableConcept.Coding, 1)
				status = status && h.assert.Equal("http://interventionengine.org/fhir/cs/huddle-member-reason", ext.ValueCodeableConcept.Coding[0].System)
				status = status && h.assert.NotEmpty(ext.ValueCodeableConcept.Text)
				foundReason = true
			}
		}
		status = status && h.assert.True(foundReason, "Member reason extension should be populated")
		status = status && h.assert.NotNil(mem.Entity)
		status = status && h.assert.Equal("Patient", mem.Entity.Type)
		status = status && h.assert.False(*mem.Entity.External)
	}

	return status
}

func (h *HuddleAssertions) AssertActiveDateTimeEqual(date time.Time) bool {
	for i := range h.Extension {
		if h.Extension[i].Url == "http://interventionengine.org/fhir/extension/group/activeDateTime" {
			return h.assert.True(h.Extension[i].ValueDateTime.Time.Equal(date), "Expected <", date, ">, Got <", h.Extension[i].ValueDateTime.Time, ">")
		}
	}
	return h.assert.Fail("Expected ActiveDateTime but did not find one")
}

func (h *HuddleAssertions) AssertLeaderIdEqual(leaderID string) bool {
	for i := range h.Extension {
		if h.Extension[i].Url == "http://interventionengine.org/fhir/extension/group/leader" {
			return h.assert.Equal(&models.Reference{
				Reference:    "Practitioner/" + leaderID,
				ReferencedID: leaderID,
				Type:         "Practitioner",
				External:     new(bool),
			}, h.Extension[i].ValueReference)
		}
	}
	return h.assert.Fail("Expected Leader but did not find one")
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
		status = status && h.assert.Len(h.Member[i].Extension, 1)
		status = status && h.assert.Equal(h.Member[i].Extension[0], models.Extension{
			Url:                  "http://interventionengine.org/fhir/extension/group/member/reason",
			ValueCodeableConcept: reason,
		})
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
