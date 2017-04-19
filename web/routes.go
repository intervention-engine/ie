package web

import (
	"errors"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/intervention-engine/fhir/server"
	"github.com/intervention-engine/ie"

	"github.com/intervention-engine/ie/controllers"
	"github.com/intervention-engine/ie/groups"
	"github.com/intervention-engine/ie/middleware"
	"github.com/intervention-engine/ie/notifications"
	"github.com/intervention-engine/ie/subscription"
)

// RegisterRoutes Create IE routes in Gin
func RegisterRoutes(s *server.FHIRServer, selfURL, riskServiceEndpoint string, enableSubscriptions bool) func() {
	returnFunc := func() {}
	if enableSubscriptions {
		workerChannel := make(chan subscription.ResourceUpdateMessage)
		watch := subscription.GenerateResourceWatch(workerChannel)
		var wg sync.WaitGroup
		wg.Add(1)
		go subscription.NotifySubscribers(workerChannel, selfURL, &wg)

		s.AddMiddleware("Condition", watch)

		s.AddMiddleware("MedicationStatement", watch)

		s.AddMiddleware("Encounter", watch)

		s.AddMiddleware("Batch", watch)

		returnFunc = func() { stopNotifier(workerChannel, &wg) }
	}

	// Setup the notification handler to use the default notification definitions (and then register it)
	notificationHandler := &middleware.NotificationHandler{Registry: notifications.DefaultNotificationDefinitionRegistry}
	s.AddMiddleware("Encounter", notificationHandler.Handle())

	s.Engine.POST("/InstaCountAll", groups.InstaCountAllHandler)
	s.Engine.GET("/NotificationCount", controllers.NotificationCountHandler)
	s.Engine.GET("/Pie/:id", controllers.GeneratePieHandler(riskServiceEndpoint))
	s.Engine.POST("/CodeLookup", controllers.CodeLookup)

	return returnFunc
}

func stopNotifier(wc chan subscription.ResourceUpdateMessage, wg *sync.WaitGroup) {
	close(wc)
	wg.Wait()
}

// RegisterAPIRoutes Register new API endpoints
func RegisterAPIRoutes(e *gin.Engine, s ie.Services) {
	api := e.Group("/api")
	RegisterPatientRoutes(api, s.PatientService(), s.MembershipService())
	RegisterCareTeamRoutes(api, s.CareTeamService())
}

func RegisterPatientRoutes(api *gin.RouterGroup, adapters ...ie.Adapter) {
	p := api.Group("/patients")
	p.GET("", ie.Adapt(ListAllPatients, adapters...))
	p.GET("/:id", ie.Adapt(GetPatient, adapters...))
	api.GET("/care_teams/:care_team_id/patients", ie.Adapt(ListAllCareTeamPatients, adapters...))
	api.PUT("/care_teams/:care_team_id/patients/:id", ie.Adapt(AddPatientToCareTeam, adapters...))
}

// func RegisterPatientRoutes(api *gin.RouterGroup, patients ie.Adapter, memberships ie.Adapter) {
// 	p := api.Group("/patients")
// 	p.GET("", ie.Adapt(ListAllPatients, patients))
// 	p.GET("/:id", ie.Adapt(GetPatient, patients, memberships))
// }

func RegisterCareTeamRoutes(api *gin.RouterGroup, careTeams ie.Adapter) {
	ct := api.Group("/care_teams")
	ct.GET("", ie.Adapt(ListAllCareTeams, careTeams))
	ct.POST("", ie.Adapt(CreateCareTeam, careTeams))
	ct.GET("/:id", ie.Adapt(GetCareTeam, careTeams))
	ct.PUT("/:id", ie.Adapt(UpdateCareTeam, careTeams))
	ct.DELETE("/:id", ie.Adapt(DeleteCareTeam, careTeams))
}

func abortNoService(ctx *gin.Context) {
	ctx.AbortWithError(http.StatusInternalServerError, errors.New("context did not contain a valid mongo service"))
}
