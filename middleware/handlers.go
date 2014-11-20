package middleware

import (
	"log"
	"net/http"

	"github.com/gorilla/context"
	"github.com/gorilla/mux"
	fhirmodels "gitlab.mitre.org/intervention-engine/fhir/models"
	"gitlab.mitre.org/intervention-engine/fhir/server"
	"gitlab.mitre.org/intervention-engine/ie/models"
	"gopkg.in/mgo.v2/bson"
)

func PatientHandler(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	log.Println("start of the patient middleware")
	next(rw, r)
	log.Println("end of the patient middleware")
	stuff := context.GetAll(r)
	for k, v := range stuff {
		log.Printf("Found a things of type %s, %s", k, v)
	}
}

func FactHandler(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	next(rw, r)

	f := models.Fact{}

	resourceType := context.Get(r, "Resource")
	resource := context.Get(r, resourceType)

	actionType := context.Get(r, "Action")
	if isFactAction(actionType.(string)) {
		switch t := resource.(type) {
		default:
			log.Printf("type of resource is %T", t)
		case *fhirmodels.Patient:
			f = models.FactFromPatient(resource.(*fhirmodels.Patient))
		case *fhirmodels.Condition:
			f = models.FactFromCondition(resource.(*fhirmodels.Condition))
		case *fhirmodels.Encounter:
			f = models.FactFromEncounter(resource.(*fhirmodels.Encounter))
		case *fhirmodels.Observation:
			f = models.FactFromObservation(resource.(*fhirmodels.Observation))
		}
		ManageFactStorage(f, actionType.(string), rw, r)
	}
	log.Println(isFactAction(actionType.(string)))
	log.Println(f.Id)
}

func isFactAction(actionType string) bool {
	return actionType != "search" && actionType != "read"
}

func ManageFactStorage(f models.Fact, actionType string, rw http.ResponseWriter, r *http.Request) {

	var err error
	factCollection := server.Database.C("facts")

	switch actionType {
	case "create":
		err = factCollection.Insert(f)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
		}
	case "update":
		tempFact := models.Fact{}
		err = factCollection.Find(bson.M{"targetid": f.TargetID}).One(&tempFact)

		f.Id = tempFact.Id

		err = factCollection.Update(bson.M{"targetid": f.TargetID}, f)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
		}
	case "delete":
		err = factCollection.Remove(bson.M{"targetid": mux.Vars(r)["id"]})
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
		}
	}
}
