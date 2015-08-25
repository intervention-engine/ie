package subscription

import (
	"net/http"
	"net/url"
	"sync"

	"github.com/intervention-engine/fhir/models"
	"github.com/intervention-engine/fhir/server"
	"gopkg.in/mgo.v2/bson"
)

func NotifySubscribers(subUpdateQueue <-chan ResourceUpdateMessage, rootURL string, wg *sync.WaitGroup) {
	for {
		rum, ok := <-subUpdateQueue
		if !ok {
			wg.Done()
			return
		}
		subsCollection := server.Database.C("subscriptions")
		query := subsCollection.Find(bson.M{"channel.type": "rest-hook"})
		var subscriptions []models.Subscription
		query.All(&subscriptions)
		for _, s := range subscriptions {
			_, err := http.PostForm(s.Channel.Endpoint,
				url.Values{"patientId": {rum.PatientID},
					"timestamp":       {rum.Timestamp},
					"fhirEndpointUrl": {rootURL}})
			if err != nil {
				s.Error = err.Error()
				subsCollection.Update(bson.M{"_id": s.Id}, s)
			}
		}
	}
}
