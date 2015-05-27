package notifications

import (
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/intervention-engine/fhir/models"
	"github.com/pebbe/util"
	. "gopkg.in/check.v1"
)

type NotificationSuite struct{}

func Test(t *testing.T) { TestingT(t) }

var _ = Suite(&NotificationSuite{})

func (n *NotificationSuite) TestInpatientAdmissionNotifications(c *C) {
	encounter, err := UnmarshallEncounter("../fixtures/encounter-inpatient.json")
	util.CheckErr(err)

	count := 0
	for _, def := range DefaultNotificationDefinitionRegistry.GetAll() {
		if def.Triggers(encounter, "create") {
			count++
			c.Assert(def.Name(), Equals, "Inpatient Admission")
			cr := def.GetNotification(encounter, "create", "http://intervention-engine.org")
			AssertEncounterNotificationContents(cr, &models.Coding{System: "http://snomed.info/sct", Code: "32485007"}, c)
		}
		c.Assert(def.Triggers(encounter, "update"), Equals, false)
		c.Assert(def.GetNotification(encounter, "update", "http://intervention-engine.org"), IsNil)
		c.Assert(def.Triggers(encounter, "delete"), Equals, false)
		c.Assert(def.GetNotification(encounter, "delete", "http://intervention-engine.org"), IsNil)

	}
	c.Assert(count, Equals, 1)
}

func (n *NotificationSuite) TestReadmissionNotifications(c *C) {
	encounter, err := UnmarshallEncounter("../fixtures/encounter-readmission.json")
	util.CheckErr(err)

	count := 0
	for _, def := range DefaultNotificationDefinitionRegistry.GetAll() {
		if def.Triggers(encounter, "create") {
			count++
			c.Assert(def.Name(), Equals, "Readmission")
			cr := def.GetNotification(encounter, "create", "http://intervention-engine.org")
			AssertEncounterNotificationContents(cr, &models.Coding{System: "http://snomed.info/sct", Code: "417005"}, c)
		}
		c.Assert(def.Triggers(encounter, "update"), Equals, false)
		c.Assert(def.GetNotification(encounter, "update", "http://intervention-engine.org"), IsNil)
		c.Assert(def.Triggers(encounter, "delete"), Equals, false)
		c.Assert(def.GetNotification(encounter, "delete", "http://intervention-engine.org"), IsNil)

	}
	c.Assert(count, Equals, 1)
}

func (n *NotificationSuite) TestERVisitNotifications(c *C) {
	encounter, err := UnmarshallEncounter("../fixtures/encounter-er-visit.json")
	util.CheckErr(err)

	count := 0
	for _, def := range DefaultNotificationDefinitionRegistry.GetAll() {
		if def.Triggers(encounter, "create") {
			count++
			c.Assert(def.Name(), Equals, "ER Visit")
			cr := def.GetNotification(encounter, "create", "http://intervention-engine.org")
			AssertEncounterNotificationContents(cr, &models.Coding{System: "http://snomed.info/sct", Code: "4525004"}, c)
		}
		c.Assert(def.Triggers(encounter, "update"), Equals, false)
		c.Assert(def.GetNotification(encounter, "update", "http://intervention-engine.org"), IsNil)
		c.Assert(def.Triggers(encounter, "delete"), Equals, false)
		c.Assert(def.GetNotification(encounter, "delete", "http://intervention-engine.org"), IsNil)

	}
	c.Assert(count, Equals, 1)
}

func (n *NotificationSuite) TestOfficeVisitDoesNotTriggerNoNotifications(c *C) {
	encounter, err := UnmarshallEncounter("../fixtures/encounter-office-visit.json")
	util.CheckErr(err)

	for _, def := range DefaultNotificationDefinitionRegistry.GetAll() {
		c.Assert(def.Triggers(encounter, "create"), Equals, false)
		c.Assert(def.GetNotification(encounter, "create", "http://intervention-engine.org"), IsNil)
		c.Assert(def.Triggers(encounter, "update"), Equals, false)
		c.Assert(def.GetNotification(encounter, "update", "http://intervention-engine.org"), IsNil)
		c.Assert(def.Triggers(encounter, "delete"), Equals, false)
		c.Assert(def.GetNotification(encounter, "delete", "http://intervention-engine.org"), IsNil)

	}
}

func (n *NotificationSuite) TestNotificationDefinitionRegistration(c *C) {
	defs := DefaultNotificationDefinitionRegistry.GetAll()
	c.Assert(defs, HasLen, 3)
	c.Assert(IsRegistered(AdmissionNotificationDefinition), Equals, true)
	c.Assert(IsRegistered(ReadmissionNotificationDefinition), Equals, true)
	c.Assert(IsRegistered(ERVisitNotificationDefinition), Equals, true)
}

func AssertEncounterNotificationContents(cr *models.CommunicationRequest, reason *models.Coding, c *C) {
	c.Assert(cr.Id, NotNil)
	c.Assert(cr.Category.Coding, HasLen, 1)
	c.Assert(cr.Category.Coding[0].System, Equals, "http://snomed.info/sct")
	c.Assert(cr.Category.Coding[0].Code, Equals, "185087000")
	c.Assert(cr.Payload, HasLen, 1)
	c.Assert(cr.Payload[0].ContentReference.Reference, Equals, "http://intervention-engine.org/Encounter/1")
	c.Assert(cr.Status, Equals, "requested")
	c.Assert(cr.Reason, HasLen, 1)
	c.Assert(cr.Reason[0].Coding, HasLen, 1)
	c.Assert(cr.Reason[0].Coding[0].System, Equals, reason.System)
	c.Assert(cr.Reason[0].Coding[0].Code, Equals, reason.Code)
	c.Assert(cr.Subject.Reference, Matches, ".*/Patient/5540f2041cd4623133000001")
	c.Assert(cr.OrderedOn.Precision, Equals, models.Precision(models.Timestamp))
	c.Assert(cr.OrderedOn.Time.Before(time.Now()), Equals, true)
	c.Assert(time.Now().Sub(cr.OrderedOn.Time) < time.Duration(5)*time.Minute, Equals, true)
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
	defer data.Close()
	if err != nil {
		return nil, err
	}

	decoder := json.NewDecoder(data)
	encounter := &models.Encounter{}
	err = decoder.Decode(encounter)
	if err == nil {
		encounter.Id = "1"
	}
	return encounter, err
}
