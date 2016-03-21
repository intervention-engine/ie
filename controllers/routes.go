package controllers

import (
	"sync"

	"github.com/intervention-engine/fhir/server"
	"github.com/intervention-engine/ie/middleware"
	"github.com/intervention-engine/ie/notifications"
	"github.com/intervention-engine/ie/subscription"
)

func RegisterRoutes(s *server.FHIRServer, selfURL, riskServiceEndpoint string) func() {
	workerChannel := make(chan subscription.ResourceUpdateMessage)
	watch := subscription.GenerateResourceWatch(workerChannel)
	var wg sync.WaitGroup
	wg.Add(1)
	go subscription.NotifySubscribers(workerChannel, selfURL, &wg)

	// Turn off auth because risk services need to connect to Patient endpoint now.
	// This is TEMPORARARY, as there is a story to apply auth to everything and
	// update risk services and other tools to pass credentials through
	//s.AddMiddleware("Group", middleware.AuthHandler())
	//s.AddMiddleware("Patient", middleware.AuthHandler())

	s.AddMiddleware("Condition", watch)

	s.AddMiddleware("MedicationStatement", watch)

	s.AddMiddleware("Encounter", watch)

	s.AddMiddleware("Batch", watch)

	// Setup the notification handler to use the default notification definitions (and then register it)
	notificationHandler := &middleware.NotificationHandler{Registry: notifications.DefaultNotificationDefinitionRegistry}
	s.AddMiddleware("Encounter", notificationHandler.Handle())

	s.Echo.Get("/GroupList/:id", PatientListHandler)
	s.Echo.Post("/InstaCountAll", InstaCountAllHandler)
	s.Echo.Get("/NotificationCount", NotificationCountHandler)
	s.Echo.Get("/Pie/:id", GeneratePieHandler(riskServiceEndpoint))
	s.Echo.Post("/CodeLookup", CodeLookup)

	login := s.Echo.Group("/login")
	login.Post("", LoginHandler)
	login.Delete("", LogoutHandler)

	return func() {
		stopNotifier(workerChannel, &wg)
	}
}

func stopNotifier(wc chan subscription.ResourceUpdateMessage, wg *sync.WaitGroup) {
	close(wc)
	wg.Wait()
}
