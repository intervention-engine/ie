package web

import (
	"errors"
	"net/http"
	"sync"

	mgo "gopkg.in/mgo.v2"

	"github.com/gin-gonic/gin"
	"github.com/intervention-engine/fhir/server"

	"github.com/intervention-engine/ie/controllers"
	"github.com/intervention-engine/ie/groups"
	"github.com/intervention-engine/ie/middleware"
	"github.com/intervention-engine/ie/mongo"
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
func RegisterAPIRoutes(e *gin.Engine, s *mgo.Session) {
	api := e.Group("/api")
	api.GET("/patients", Adapt(ListAllPatients, withMongoPatientService(s)))
	api.GET("/patients/:id", Adapt(GetPatient, withMongoPatientService(s)))

	api.GET("/care_teams", Adapt(ListAllCareTeams, withMongoCareTeamService(s)))
	api.POST("/care_teams", Adapt(CreateCareTeam, withMongoCareTeamService(s)))
	api.GET("/care_teams/:id", Adapt(GetCareTeam, withMongoCareTeamService(s)))
	api.PUT("/care_teams/:id", Adapt(UpdateCareTeam, withMongoCareTeamService(s)))
	api.DELETE("/care_teams/:id", Adapt(DeleteCareTeam, withMongoCareTeamService(s)))
}

type Adapter func(gin.HandlerFunc) gin.HandlerFunc

func Adapt(h gin.HandlerFunc, adapters ...Adapter) gin.HandlerFunc {
	for _, wrapper := range adapters {
		h = wrapper(h)
	}
	return h
}

// withMongo puts a mgo.Collection named "care_teams" into the context for the actual handler to use.
func withMongoCareTeamService(s *mgo.Session) Adapter {
	return func(h gin.HandlerFunc) gin.HandlerFunc {
		return func(ctx *gin.Context) {
			session := s.Copy()
			defer session.Close()
			col := session.DB("fhir").C("care_teams")
			service := &mongo.CareTeamService{C: col}
			ctx.Set("service", service)
			h(ctx)
		}
	}
}

func withMongoPatientService(s *mgo.Session) Adapter {
	return func(h gin.HandlerFunc) gin.HandlerFunc {
		return func(ctx *gin.Context) {
			session := s.Copy()
			defer session.Close()
			col := session.DB("fhir").C("patients")
			service := &mongo.PatientService{C: col}
			ctx.Set("service", service)
			h(ctx)
		}
	}
}

func abortNoService(ctx *gin.Context) {
	ctx.AbortWithError(http.StatusInternalServerError, errors.New("context did not contain a valid mongo service"))
}
