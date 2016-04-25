package notifications

import (
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/intervention-engine/fhir/models"
	"github.com/stretchr/testify/suite"
)

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestNotificationSuite(t *testing.T) {
	suite.Run(t, new(NotificationSuite))
}

type NotificationSuite struct {
	suite.Suite
}

func (n *NotificationSuite) TestInpatientAdmissionNotifications() {
	require := n.Require()
	assert := n.Assert()

	encounter, err := UnmarshallEncounter("../fixtures/encounter-inpatient.json")
	require.NoError(err)

	count := 0
	for _, def := range DefaultNotificationDefinitionRegistry.GetAll() {
		if def.Triggers(encounter, "create") {
			count++
			assert.Equal("Inpatient Admission", def.Name())
			cr := def.GetNotification(encounter, "create", "http://intervention-engine.org")
			n.AssertEncounterNotificationContents(cr, &models.Coding{System: "http://snomed.info/sct", Code: "32485007"})
		}
		assert.False(def.Triggers(encounter, "update"))
		assert.Nil(def.GetNotification(encounter, "update", "http://intervention-engine.org"))
		assert.False(def.Triggers(encounter, "delete"))
		assert.Nil(def.GetNotification(encounter, "delete", "http://intervention-engine.org"))
	}
	assert.Equal(1, count)
}

func (n *NotificationSuite) TestReadmissionNotifications() {
	require := n.Require()
	assert := n.Assert()

	encounter, err := UnmarshallEncounter("../fixtures/encounter-readmission.json")
	require.NoError(err)

	count := 0
	for _, def := range DefaultNotificationDefinitionRegistry.GetAll() {
		if def.Triggers(encounter, "create") {
			count++
			assert.Equal("Readmission", def.Name())
			cr := def.GetNotification(encounter, "create", "http://intervention-engine.org")
			n.AssertEncounterNotificationContents(cr, &models.Coding{System: "http://snomed.info/sct", Code: "417005"})
		}
		assert.False(def.Triggers(encounter, "update"))
		assert.Nil(def.GetNotification(encounter, "update", "http://intervention-engine.org"))
		assert.False(def.Triggers(encounter, "delete"))
		assert.Nil(def.GetNotification(encounter, "delete", "http://intervention-engine.org"))
	}
	assert.Equal(1, count)
}

func (n *NotificationSuite) TestERVisitNotifications() {
	require := n.Require()
	assert := n.Assert()

	encounter, err := UnmarshallEncounter("../fixtures/encounter-er-visit.json")
	require.NoError(err)

	count := 0
	for _, def := range DefaultNotificationDefinitionRegistry.GetAll() {
		if def.Triggers(encounter, "create") {
			count++
			assert.Equal("ER Visit", def.Name())
			cr := def.GetNotification(encounter, "create", "http://intervention-engine.org")
			n.AssertEncounterNotificationContents(cr, &models.Coding{System: "http://snomed.info/sct", Code: "4525004"})
		}
		assert.False(def.Triggers(encounter, "update"))
		assert.Nil(def.GetNotification(encounter, "update", "http://intervention-engine.org"))
		assert.False(def.Triggers(encounter, "delete"))
		assert.Nil(def.GetNotification(encounter, "delete", "http://intervention-engine.org"))
	}
	assert.Equal(1, count)
}

func (n *NotificationSuite) TestOfficeVisitDoesNotTriggerNoNotifications() {
	require := n.Require()
	assert := n.Assert()

	encounter, err := UnmarshallEncounter("../fixtures/encounter-office-visit.json")
	require.NoError(err)

	for _, def := range DefaultNotificationDefinitionRegistry.GetAll() {
		assert.False(def.Triggers(encounter, "create"))
		assert.Nil(def.GetNotification(encounter, "create", "http://intervention-engine.org"))
		assert.False(def.Triggers(encounter, "update"))
		assert.Nil(def.GetNotification(encounter, "update", "http://intervention-engine.org"))
		assert.False(def.Triggers(encounter, "delete"))
		assert.Nil(def.GetNotification(encounter, "delete", "http://intervention-engine.org"))
	}
}

func (n *NotificationSuite) TestNotificationDefinitionRegistration() {
	assert := n.Assert()

	defs := DefaultNotificationDefinitionRegistry.GetAll()
	assert.Len(defs, 3)
	assert.True(IsRegistered(AdmissionNotificationDefinition))
	assert.True(IsRegistered(ReadmissionNotificationDefinition))
	assert.True(IsRegistered(ERVisitNotificationDefinition))
}

func (n *NotificationSuite) AssertEncounterNotificationContents(cr *models.CommunicationRequest, reason *models.Coding) {
	assert := n.Assert()
	assert.NotNil(cr.Id)
	assert.Len(cr.Category.Coding, 1)
	assert.Equal("http://snomed.info/sct", cr.Category.Coding[0].System)
	assert.Equal("185087000", cr.Category.Coding[0].Code)
	assert.Len(cr.Payload, 1)
	assert.Equal("http://intervention-engine.org/Encounter/1", cr.Payload[0].ContentReference.Reference)
	assert.Equal("1", cr.Payload[0].ContentReference.ReferencedID)
	assert.Equal("Encounter", cr.Payload[0].ContentReference.Type)
	assert.Equal("requested", cr.Status)
	assert.Len(cr.Reason, 1)
	assert.Len(cr.Reason[0].Coding, 1)
	assert.Equal(reason.System, cr.Reason[0].Coding[0].System)
	assert.Equal(reason.Code, cr.Reason[0].Coding[0].Code)
	assert.Regexp(".*/Patient/5540f2041cd4623133000001", cr.Subject.Reference)
	assert.Equal(models.Precision(models.Timestamp), cr.RequestedOn.Precision)
	assert.True(cr.RequestedOn.Time.Before(time.Now()))
	assert.True(time.Now().Sub(cr.RequestedOn.Time) < time.Duration(5)*time.Minute)
}

func IsRegistered(n NotificationDefinition) bool {
	for _, def := range DefaultNotificationDefinitionRegistry.GetAll() {
		if def == n {
			return true
		}
	}
	return false
}

func UnmarshallEncounter(file string) (*models.Encounter, error) {
	data, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer data.Close()

	decoder := json.NewDecoder(data)
	encounter := &models.Encounter{}
	err = decoder.Decode(encounter)
	if err == nil {
		encounter.Id = "1"
	}
	return encounter, err
}
