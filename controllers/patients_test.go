package controllers_test

import (
	"testing"

	"github.com/intervention-engine/ie/controllers"
	"github.com/intervention-engine/ie/models"
	"github.com/stretchr/testify/suite"
)

type PatientsControllerSuite struct {
	IEControllerSuite
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestPatientControllerSuite(t *testing.T) {
	suite.Run(t, new(PatientsControllerSuite))
}

func (c *PatientsControllerSuite) SetupSuite() {
	c.target = &controllers.IEController{DB: c.DB(), CollectionName: "patients", Item: models.Patient{}}
}

func (c *PatientsControllerSuite) TestAll() {
	c.target.All(c.ctx)
}
