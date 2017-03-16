package web_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"testing"
	"time"

	"gopkg.in/mgo.v2/bson"

	"github.com/gin-gonic/gin"
	"github.com/intervention-engine/ie"
	"github.com/intervention-engine/ie/testutil"
	"github.com/intervention-engine/ie/web"
	"github.com/stretchr/testify/suite"
)

type careTeamSuite struct {
	testutil.WebSuite
	DB map[string]ie.CareTeam
}

func TestCareTeamHandlersSuite(t *testing.T) {
	m := make(map[string]ie.CareTeam)
	suite.Run(t, &careTeamSuite{DB: m})
}

func (suite *careTeamSuite) SetupSuite() {
	api := suite.LoadGin()
	api.GET("/care_teams", web.Adapt(web.ListAllCareTeams, suite.withTestService()))
	api.POST("/care_teams", web.Adapt(web.CreateCareTeam, suite.withTestService()))
	api.GET("/care_teams/:id", web.Adapt(web.GetCareTeam, suite.withTestService()))
	api.PUT("/care_teams/:id", web.Adapt(web.UpdateCareTeam, suite.withTestService()))
	api.DELETE("/care_teams/:id", web.Adapt(web.DeleteCareTeam, suite.withTestService()))
}

func (suite *careTeamSuite) SetupTest() {
	for _, team := range careTeams {
		suite.DB[team.ID] = team
	}
}

var careTeams = []ie.CareTeam{
	ie.CareTeam{
		ID:        "58c314acb367c1ff54d19e9e",
		Name:      "The AutoDocs",
		Leader:    "Optometrist Prime",
		CreatedAt: time.Now(),
	},
	ie.CareTeam{
		ID:        "58c314acb367c1ff54d19e9f",
		Name:      "RoboDocs",
		Leader:    "",
		CreatedAt: time.Now(),
	},
}

// TODO: Should make some permanent times so we can test that, too
// When there are care teams, should return a slice of care teams with 200 OK
func (suite *careTeamSuite) TestAllCareTeamsFound() {
	w := suite.AssertGetRequest("/api/care_teams", http.StatusOK)

	var body = make(map[string][]ie.CareTeam)
	json.NewDecoder(w.Body).Decode(&body)
	results, ok := body["care_teams"]
	if !ok {
		suite.T().Errorf("body did not contain an element named care_teams | body: %#+v\n", w.Body.String())
		return
	}
	for _, team := range careTeams {
		result := findCareTeam(team.ID, results)

		suite.Assert().NotNil(result, "body did not contain team: %s with ID: %s\n", team.Name, team.ID)
		suite.Assert().Equal(team.Leader, result.Leader, "team: %s did not contain correct leader | want: %s got: %s", team.Name, team.Leader, result.Leader)
		suite.Assert().Equal(team.Name, result.Name, "team: %s did not contain correct name | want: %s got: %s", team.Name, team.Name, result.Name)
	}
}

// If a care team with that (correct) id does not exist in the database, should
// return 404 Not Found
func (suite *careTeamSuite) TestGetCareTeamNotFound() {
	id := careTeams[0].ID
	delete(suite.DB, id)
	suite.AssertGetRequest("/api/care_teams/"+id, http.StatusNotFound)
}

// When given an incorrectly formatted id, should return 400 Bad Request
func (suite *careTeamSuite) TestGetCareTeamWithBadId() {
	suite.AssertGetRequest("/api/care_teams/sdf", http.StatusBadRequest)
}

// If a care team with that id exists, should return the care team and 200 OK
func (suite *careTeamSuite) TestGetCareTeamFound() {
	team := careTeams[0]
	w := suite.AssertGetRequest("/api/care_teams/58c314acb367c1ff54d19e9e", http.StatusOK)
	var body = make(map[string]ie.CareTeam)
	json.NewDecoder(w.Body).Decode(&body)
	result, ok := body["care_team"]
	if !ok {
		suite.T().Errorf("body did not contain an element named care_team | body: %#+v\n", w.Body.String())
		return
	}
	suite.Assert().Equal(team.Leader, result.Leader, "team: %s did not contain correct leader | want: %s got: %s", team.Name, team.Leader, result.Leader)
	suite.Assert().Equal(team.Name, result.Name, "team: %s did not contain correct name | want: %s got: %s", team.Name, team.Name, result.Name)
}

// When called with no form data, should return 400 Bad Request
func TestCreateCareTeamWithNoFormData(t *testing.T) {

}

// When called with correct form data, should return the created object in
// properly formatted JSON
func TestCreateCareTeamReturnsProperJSONOnSuccess(t *testing.T) {

}

// Mock Services

func (suite *careTeamSuite) CareTeam(id string) (*ie.CareTeam, error) {
	if !bson.IsObjectIdHex(id) {
		return nil, errors.New("bad id")
	}
	c, ok := suite.DB[id]
	if !ok {
		return nil, errors.New("not found")
	}
	return &c, nil
}

func (suite *careTeamSuite) CareTeams() ([]ie.CareTeam, error) {
	var cc []ie.CareTeam
	for _, team := range suite.DB {
		cc = append(cc, team)
	}
	return cc, nil
}

func (suite *careTeamSuite) CreateCareTeam(c *ie.CareTeam) error {
	c.ID = bson.NewObjectId().String()
	c.CreatedAt = time.Now()
	suite.DB[c.ID] = *c
	return nil
}

func (suite *careTeamSuite) UpdateCareTeam(c *ie.CareTeam) error {
	suite.DB[c.ID] = *c
	return nil
}

func (suite *careTeamSuite) DeleteCareTeam(id string) error {
	if !bson.IsObjectIdHex(id) {
		return errors.New("id was not proper object id hex")
	}
	if _, ok := suite.DB[id]; !ok {
		return errors.New("not found")
	}
	delete(suite.DB, id)
	return nil
}

// Utility Methods

func (suite *careTeamSuite) withTestService() web.Adapter {
	return func(h gin.HandlerFunc) gin.HandlerFunc {
		return func(ctx *gin.Context) {
			ctx.Set("service", suite)
			h(ctx)
		}
	}
}

func findCareTeam(ID string, teams []ie.CareTeam) *ie.CareTeam {
	for _, t := range teams {
		if ID == t.ID {
			return &t
		}
	}
	return nil
}
