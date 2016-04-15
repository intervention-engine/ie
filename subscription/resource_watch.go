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
	for patientID, timestamp := range triggeredPatients(resource) {
		ru := NewResourceUpdateMessage(patientID, timestamp.Format(time.RFC3339))
		subUpdateQueue <- ru
	}
}

func triggeredPatients(resource interface{}) map[string]time.Time {
	result := make(map[string]time.Time)

	switch t := resource.(type) {
	case *fhirmodels.Condition:
		result[t.Patient.ReferencedID] = t.OnsetDateTime.Time
	case *fhirmodels.MedicationStatement:
		result[t.Patient.ReferencedID] = t.EffectivePeriod.Start.Time
	case *fhirmodels.Encounter:
		result[t.Patient.ReferencedID] = t.Period.Start.Time
	case *fhirmodels.Bundle:
		for _, entry := range t.Entry {
			if entry.Resource != nil {
				subResult := triggeredPatients(entry.Resource)
				for k, v := range subResult {
					if t, ok := result[k]; !ok || v.After(t) {
						result[k] = v
					}
				}

			}
		}
	}
	return result
}
