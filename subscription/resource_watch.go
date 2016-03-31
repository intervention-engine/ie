package subscription

import (
	"time"

	"github.com/gin-gonic/gin"
	fhirmodels "github.com/intervention-engine/fhir/models"
)

func GenerateResourceWatch(subUpdateQueue chan<- ResourceUpdateMessage) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		if c.IsAborted() {
			return
		}
		if c.Request.Method == "GET" {
			return
		}
		if resourceType, ok := c.Get("Resource"); ok {
			if resource, ok := c.Get(resourceType.(string)); ok {
				HandleResourceUpdate(subUpdateQueue, resource)
			}
		}
		return
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
			if entry.Resource != nil {
				HandleResourceUpdate(subUpdateQueue, entry.Resource)
			}
		}
		return
	default:
		return
	}

	ru := NewResourceUpdateMessage(patientID, timestamp.Format(time.RFC3339))
	subUpdateQueue <- ru
}
