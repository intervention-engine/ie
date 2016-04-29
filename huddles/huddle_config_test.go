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
func TestHuddleConfigSuite(t *testing.T) {
	suite.Run(t, new(HuddleConfigSuite))
}

type HuddleConfigSuite struct {
	suite.Suite
	SimpleConfig *HuddleConfig
}

func (suite *HuddleConfigSuite) SetupSuite() {
	require := suite.Require()

	data, err := ioutil.ReadFile("../fixtures/huddle_config.json")
	require.NoError(err)
	suite.SimpleConfig = new(HuddleConfig)
	err = json.Unmarshal(data, suite.SimpleConfig)
	require.NoError(err)
}

func (suite *HuddleConfigSuite) TestLoadHuddleFromJSON() {
	require := suite.Require()
	assert := suite.Assert()

	config := suite.SimpleConfig
	assert.Equal("Simple Huddle", config.Name)
	assert.Equal("1", config.LeaderID)
	assert.Equal([]time.Weekday{time.Monday}, config.Days)
	assert.Equal(4, config.LookAhead)
	require.NotNil(config.RiskConfig)
	assert.Equal(models.Coding{
		System: "http://interventionengine.org/risk-assessments",
		Code:   "Simple",
	}, config.RiskConfig.RiskMethod)
	assert.Equal([]RiskScoreFrequencyConfig{
		{
			MinScore:              6,
			MaxScore:              10,
			MinDaysBetweenHuddles: 5,
			MaxDaysBetweenHuddles: 7,
		}, {
			MinScore:              4,
			MaxScore:              5,
			MinDaysBetweenHuddles: 12,
			MaxDaysBetweenHuddles: 14,
		}, {
			MinScore:              1,
			MaxScore:              3,
			MinDaysBetweenHuddles: 25,
			MaxDaysBetweenHuddles: 28,
		},
	}, config.RiskConfig.FrequencyConfigs)
	require.NotNil(config.EventConfig)
	assert.Equal([]EncounterEventConfig{
		{
			LookBackDays: 7,
			TypeCodes: []EventCode{
				{
					Name:       "Hospital Discharge",
					System:     "http://snomed.info/sct",
					Code:       "32485007",
					UseEndDate: true,
				},
				{
					Name:       "Hospital Admission",
					System:     "http://snomed.info/sct",
					Code:       "32485007",
					UseEndDate: false,
				},
				{
					Name:       "Hospital Re-Admission Discharge",
					System:     "http://snomed.info/sct",
					Code:       "417005",
					UseEndDate: true,
				},
				{
					Name:       "Hospital Re-Admission",
					System:     "http://snomed.info/sct",
					Code:       "417005",
					UseEndDate: false,
				},
				{
					Name:       "Emergency Room Admission",
					System:     "http://snomed.info/sct",
					Code:       "50849002",
					UseEndDate: false,
				},
			},
		},
	}, config.EventConfig.EncounterConfigs)
	assert.Equal(config.SchedulerCronSpec, "@midnight")
}

func (suite *HuddleConfigSuite) TestIsHuddleDay() {
	assert := suite.Assert()

	monday := time.Date(2016, time.March, 21, 0, 0, 0, 0, time.Local)
	assert.True(suite.SimpleConfig.IsHuddleDay(monday), "March 21 is a Monday, so should be a huddle day")
	for i := 1; i < 7; i++ {
		nonMonday := monday.AddDate(0, 0, i)
		assert.False(suite.SimpleConfig.IsHuddleDay(nonMonday), "Should not be a huddle day: %#v", nonMonday)
	}
}
