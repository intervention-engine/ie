package middleware

import (
	"net/http"
	"net/url"
	"time"
	"fmt"

	"github.com/codegangsta/negroni"
	"github.com/gorilla/context"
	fhirmodels "github.com/intervention-engine/fhir/models"
)

func GenerateRiskHandler(riskEndpoint, rootURL string) negroni.HandlerFunc {
	f := func(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
		next(rw, r)
		resourceType, ok := context.GetOk(r, "Resource")
		if ok {
			resource := context.Get(r, resourceType)

			var patientId string
			var timestamp time.Time

			switch resource.(type) {
			case *fhirmodels.Condition:
				patientId = resource.(*fhirmodels.Condition).Patient.ReferencedID
				timestamp = resource.(*fhirmodels.Condition).OnsetDateTime.Time
			case *fhirmodels.MedicationStatement:
				patientId = resource.(*fhirmodels.MedicationStatement).Patient.ReferencedID
				timestamp = resource.(*fhirmodels.MedicationStatement).EffectivePeriod.Start.Time
			}

			calculateEndpoint := riskEndpoint + "/calculate"

			fmt.Println(calculateEndpoint)

			_, err := http.PostForm(calculateEndpoint,
				url.Values{"patientId": {patientId},
					"timestamp":       {timestamp.Format(time.RFC3339)},
					"fhirEndpointUrl": {rootURL}})
			if err != nil {
				http.Error(rw, err.Error(), http.StatusInternalServerError)
			}
		}
	}
	return f
}
