package controllers

import (
	"sync"

	"github.com/intervention-engine/fhir/server"
	"github.com/intervention-engine/ie/groups"
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

	s.Engine.POST("/InstaCountAll", groups.InstaCountAllHandler)
	s.Engine.GET("/NotificationCount", NotificationCountHandler)
	s.Engine.GET("/Pie/:id", GeneratePieHandler(riskServiceEndpoint))
	s.Engine.POST("/CodeLookup", CodeLookup)

	login := s.Engine.Group("/login")
	login.POST("", LoginHandler)
	login.DELETE("", LogoutHandler)

	return func() {
		stopNotifier(workerChannel, &wg)
	}
}

func stopNotifier(wc chan subscription.ResourceUpdateMessage, wg *sync.WaitGroup) {
	close(wc)
	wg.Wait()
}
