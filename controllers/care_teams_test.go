package controllers_test

import (
	"testing"

	"github.com/intervention-engine/ie/controllers"
	"github.com/intervention-engine/ie/models"
	"github.com/stretchr/testify/suite"
)

var care_teams *controllers.IEController

type CareTeamsControllerSuite struct {
	IEControllerSuite
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestIEControllerSuite(t *testing.T) {
	suite.Run(t, new(CareTeamsControllerSuite))
}

func (c *CareTeamsControllerSuite) SetupSuite() {
	c.target = &controllers.IEController{DB: c.DB(), CollectionName: "care_teams", Item: models.CareTeam{}}
}

func (c *CareTeamsControllerSuite) TestAll() {
	c.target.All(c.ctx)
}
