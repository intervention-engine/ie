package appt_test

import (
	"testing"

	"github.com/intervention-engine/ie/appt"
	"github.com/intervention-engine/ie/mongo"
	mgo "gopkg.in/mgo.v2"
)

func TestSchedule(t *testing.T) {
	// Get a connection to a test database, then call BatchSchedule.
	// Connect database
	session, err := mgo.Dial("mongodb://localhost:27017")
	if err != nil {
		t.Errorf("dialing mongo failed for session at: %s", "mongodb://localhost:27017")
	}
	defer session.Close()
	svcFactory := mongo.NewServiceFactory(session.Copy(), "fhir")
	appt.ManualSchedule(svcFactory.NewSchedService(), []string{"../config/simple_huddle_config.json"})
}