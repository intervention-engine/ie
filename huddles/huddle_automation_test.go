package huddles

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/intervention-engine/fhir/models"
	"github.com/intervention-engine/fhir/server"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"gopkg.in/mgo.v2/dbtest"
)

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestHuddleAutomationSuite(t *testing.T) {
	suite.Run(t, new(HuddleAutomationSuite))
}

type HuddleAutomationSuite struct {
	suite.Suite
	DBServer     *dbtest.DBServer
	DBServerPath string
	Session      *mgo.Session
	Database     *mgo.Database
}

func (suite *HuddleAutomationSuite) SetupSuite() {
	// Set up the database
	suite.DBServer = &dbtest.DBServer{}
	var err error
	suite.DBServerPath, err = ioutil.TempDir("", "mongotestdb")
	if err != nil {
		panic(err)
	}
	suite.DBServer.SetPath(suite.DBServerPath)
	RegisterCustomSearchDefinitions()
}

func (suite *HuddleAutomationSuite) SetupTest() {
	suite.Session = suite.DBServer.Session()
	suite.Database = suite.Session.DB("riskservice-test")
	server.Database = suite.Database
}

func (suite *HuddleAutomationSuite) TearDownTest() {
	suite.Session.Close()
	suite.DBServer.Wipe()
}

func (suite *HuddleAutomationSuite) TearDownSuite() {
	suite.DBServer.Stop()
	if err := os.RemoveAll(suite.DBServerPath); err != nil {
		fmt.Fprintf(os.Stderr, "WARNING: Error cleaning up temp directory: %s", err.Error())
	}
}

// For now, this test just uses risk scores, but it should be expanded as we add support for autopopulating
// based on other criteria.
func (suite *HuddleAutomationSuite) TestAutoPopulateNewHuddle() {
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

	group, err := CreateAutoPopulatedHuddle(time.Date(2016, time.March, 21, 9, 0, 0, 0, time.UTC), config)
	require.NoError(err)
	assert.Contains(group.Meta.Profile, "http://interventionengine.org/fhir/profile/huddle")
	assert.Contains(group.Extension, models.Extension{
		Url:           "http://interventionengine.org/fhir/extension/group/activeDateTime",
		ValueDateTime: &models.FHIRDateTime{Time: time.Date(2016, time.March, 21, 9, 0, 0, 0, time.UTC), Precision: models.Precision(models.Date)},
	})
	assert.Contains(group.Extension, models.Extension{
		Url: "http://interventionengine.org/fhir/extension/group/leader",
		ValueReference: &models.Reference{
			Reference:    "Practitioner/123",
			ReferencedID: "123",
			Type:         "Practitioner",
			External:     new(bool),
		},
	})
	assert.Equal("person", group.Type)
	assert.True(*group.Actual)
	assert.Equal(&models.CodeableConcept{
		Coding: []models.Coding{
			{System: "http://interventionengine.org/fhir/cs/huddle", Code: "HUDDLE"},
		},
		Text: "Huddle",
	}, group.Code)
	assert.Equal("Test Huddle Config", group.Name)
	assert.Len(group.Member, 11)

	expectedIDs := []string{bsonID(20), bsonID(16), bsonID(19), bsonID(18), bsonID(17), bsonID(15), bsonID(13),
		bsonID(12), bsonID(7), bsonID(8), bsonID(10)}
	for i, expectedID := range expectedIDs {
		assert.Contains(group.Member[i].Extension, models.Extension{
			Url: "http://interventionengine.org/fhir/extension/group/member/reason",
			ValueCodeableConcept: &models.CodeableConcept{
				Coding: []models.Coding{
					{System: "http://interventionengine.org/fhir/cs/huddle-member-reason", Code: "RISK_SCORE"},
				},
				Text: "Risk Score Warrants Discussion",
			},
		})
		assert.Equal(&models.Reference{
			Reference:    "Patient/" + expectedID,
			ReferencedID: expectedID,
			Type:         "Patient",
			External:     new(bool),
		}, group.Member[i].Entity)
	}
}

// This test just ensures that manually added patients aren't overwritten by the huddle population
// algorithm.
func (suite *HuddleAutomationSuite) TestAutoPopulateExistingHuddle() {
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

	group, err := CreateAutoPopulatedHuddle(time.Date(2016, time.March, 21, 0, 0, 0, 0, time.Local), config)
	require.NoError(err)
	assert.Contains(group.Meta.Profile, "http://interventionengine.org/fhir/profile/huddle")

	assert.Contains(group.Extension, models.Extension{
		Url:           "http://interventionengine.org/fhir/extension/group/activeDateTime",
		ValueDateTime: &models.FHIRDateTime{Time: time.Date(2016, time.March, 21, 0, 0, 0, 0, time.Local), Precision: models.Precision(models.Date)},
	})
	assert.Contains(group.Extension, models.Extension{
		Url: "http://interventionengine.org/fhir/extension/group/leader",
		ValueReference: &models.Reference{
			Reference:    "Practitioner/123",
			ReferencedID: "123",
			Type:         "Practitioner",
			External:     new(bool),
		},
	})
	assert.Equal("person", group.Type)
	assert.True(*group.Actual)
	assert.Equal(&models.CodeableConcept{
		Coding: []models.Coding{
			{System: "http://interventionengine.org/fhir/cs/huddle", Code: "HUDDLE"},
		},
		Text: "Huddle",
	}, group.Code)
	assert.True(strings.HasPrefix(group.Name, "Test Huddle"))
	require.Len(group.Member, 3)

	expectedIDs := []string{bsonID(1), bsonID(4), bsonID(2)}
	for i, expectedID := range expectedIDs {
		if assert.Len(group.Member[i].Extension, 1) {
			assert.Equal("http://interventionengine.org/fhir/extension/group/member/reason", group.Member[i].Extension[0].Url)
			if assert.NotNil(group.Member[i].Extension[0].ValueCodeableConcept) {
				assert.Equal("http://interventionengine.org/fhir/cs/huddle-member-reason", group.Member[i].Extension[0].ValueCodeableConcept.Coding[0].System)
				if expectedID == bsonID(1) {
					assert.Equal("MANUAL_ADDITION", group.Member[i].Extension[0].ValueCodeableConcept.Coding[0].Code)
					assert.Equal("Manually Added", group.Member[i].Extension[0].ValueCodeableConcept.Text)
				} else {
					assert.Equal("RISK_SCORE", group.Member[i].Extension[0].ValueCodeableConcept.Coding[0].Code)
					assert.Equal("Risk Score Warrants Discussion", group.Member[i].Extension[0].ValueCodeableConcept.Text)
				}
			}
		}
		assert.Equal(&models.Reference{
			Reference:    "Patient/" + expectedID,
			ReferencedID: expectedID,
			Type:         "Patient",
			External:     new(bool),
		}, group.Member[i].Entity)
	}
}

func (suite *HuddleAutomationSuite) TestFindEligiblePatientIDsByRiskScoreWithNoPreviousHuddles() {
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

func (suite *HuddleAutomationSuite) TestFindEligiblePatientIDsByRiskScore() {
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
	c.RiskConfig = new(ScheduleByRiskConfig)
	c.RiskConfig.RiskCode = models.Coding{System: "http://interventionengine.org/risk-assessments", Code: "Test"}
	day := 24 * time.Hour
	c.RiskConfig.FrequencyConfigs = []RiskScoreFrequencyConfig{
		{
			MinScore:              8,
			MaxScore:              10,
			MinTimeBetweenHuddles: 5 * day,
			MaxTimeBetweenHuddles: 7 * day,
		},
		{
			MinScore:              6,
			MaxScore:              7,
			MinTimeBetweenHuddles: 12 * day,
			MaxTimeBetweenHuddles: 14 * day,
		},
		{
			MinScore:              3,
			MaxScore:              5,
			MinTimeBetweenHuddles: 25 * day,
			MaxTimeBetweenHuddles: 28 * day,
		},
	}
	return c
}

// storePatientAndScores creates a new patient with the given ID and a set of risk
// scores.  The scores are assumed to be in chronological order, so the last score
// passed in is the MOST_RECENT.
func (suite *HuddleAutomationSuite) storePatientAndScores(id string, scores ...int) {
	require := require.New(suite.T())

	p := new(models.Patient)
	p.Id = id
	require.NoError(suite.Database.C("patients").Insert(p))

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
			require.NoError(suite.Database.C("riskassessments").Insert(ra))
			date = date.Add(7 * 24 * time.Hour)
		}
	}
}

func (suite *HuddleAutomationSuite) storeHuddleWithReasons(date time.Time, leaderID string, defaultReason *models.CodeableConcept, reasonMap map[string]*models.CodeableConcept, patients ...string) {
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

	require.NoError(suite.Database.C("groups").Insert(g))
}

func (suite *HuddleAutomationSuite) storeHuddle(date time.Time, leaderID string, reason *models.CodeableConcept, patients ...string) {
	suite.storeHuddleWithReasons(date, leaderID, reason, nil, patients...)
}

func bsonID(id int) string {
	return fmt.Sprintf("%024x", id)
}
