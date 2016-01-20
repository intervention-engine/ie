package subscription

import (
	"time"

	fhirmodels "github.com/intervention-engine/fhir/models"
	"github.com/labstack/echo"
)

func GenerateResourceWatch(subUpdateQueue chan<- ResourceUpdateMessage) echo.MiddlewareFunc {
	return func(hf echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			err := hf(c)
			if err != nil {
				return err
			}
			resourceType := c.Get("Resource")
			if resourceType != nil {
				resource := c.Get(resourceType.(string))
				HandleResourceUpdate(subUpdateQueue, resource)
			}
			return nil
		}
	}
}

func HandleResourceUpdate(subUpdateQueue chan<- ResourceUpdateMessage, resource interface{}) {
	var patientID string
	var timestamp time.Time

	switch t := resource.(type) {
	case *fhirmodels.Condition:
		patientID = t.Patient.ReferencedID
		timestamp = t.OnsetDateTime.Time
	case *fhirmodels.MedicationStatement:
		patientID = t.Patient.ReferencedID
		timestamp = t.EffectivePeriod.Start.Time
	case *fhirmodels.Encounter:
		patientID = t.Patient.ReferencedID
		timestamp = t.Period.Start.Time
	case *fhirmodels.Bundle:
		for _, entry := range t.Entry {
			HandleResourceUpdate(subUpdateQueue, entry.Resource)
		}
		return
	default:
		return
	}

	ru := NewResourceUpdateMessage(patientID, timestamp.Format(time.RFC3339))
	subUpdateQueue <- ru
}
