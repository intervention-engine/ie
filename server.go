package main

import (
	"github.com/codegangsta/negroni"
	"github.com/intervention-engine/fhir/server"
	"github.com/intervention-engine/ie/controllers"
	"github.com/intervention-engine/ie/middleware"
	"github.com/intervention-engine/ie/notifications"
	"os"
	"net"
	"fmt"
)

//var Store sessions.Store

func main() {
	// Check for a linked MongoDB container if we are running in Docker
	mongoHost := os.Getenv("MONGO_PORT_27017_TCP_ADDR")
	if mongoHost == "" {
		mongoHost = "localhost"
	}

	riskServiceEndpoint := os.Getenv("RISKSERVICE_PORT_9000_TCP_ADDR")
	if riskServiceEndpoint == "" {
		riskServiceEndpoint = "http://localhost:9000/calculate"
	} else {
		riskServiceEndpoint = "http://" + riskServiceEndpoint + ":9000/calculate"
	}

	var ip net.IP
	var selfUrl string
	host, err := os.Hostname()
	if err != nil {
		panic(err)
	}
	addrs, err := net.LookupIP(host)
	if err != nil {
		fmt.Println("Unable to lookup IP based on hostname, defaulting to localhost.")
		selfUrl = "http://localhost:3001"
	}
	for _, addr := range addrs {
		if ipv4 := addr.To4(); ipv4 != nil {
			ip = ipv4
			selfUrl = "http://" + ip.String() + ":3001"
		}
	}

	s := server.NewServer(mongoHost)

	s.AddMiddleware("GroupCreate", negroni.HandlerFunc(middleware.AuthHandler))
	s.AddMiddleware("GroupCreate", negroni.HandlerFunc(middleware.QueryExecutionHandler))

	s.AddMiddleware("PatientCreate", negroni.HandlerFunc(middleware.FactHandler))
	s.AddMiddleware("PatientUpdate", negroni.HandlerFunc(middleware.FactHandler))
	s.AddMiddleware("PatientDelete", negroni.HandlerFunc(middleware.FactHandler))
	s.AddMiddleware("PatientIndex", negroni.HandlerFunc(middleware.RiskQueryHandler))
	s.AddMiddleware("PatientIndex", negroni.HandlerFunc(middleware.AuthHandler))

	s.AddMiddleware("ConditionCreate", negroni.HandlerFunc(middleware.FactHandler))
	s.AddMiddleware("ConditionCreate", middleware.GenerateRiskHandler(riskServiceEndpoint, selfUrl))
	s.AddMiddleware("ConditionUpdate", negroni.HandlerFunc(middleware.FactHandler))
	s.AddMiddleware("ConditionDelete", negroni.HandlerFunc(middleware.FactHandler))

	s.AddMiddleware("EncounterCreate", negroni.HandlerFunc(middleware.FactHandler))
	s.AddMiddleware("EncounterUpdate", negroni.HandlerFunc(middleware.FactHandler))
	s.AddMiddleware("EncounterDelete", negroni.HandlerFunc(middleware.FactHandler))

	s.AddMiddleware("ObservationCreate", negroni.HandlerFunc(middleware.FactHandler))
	s.AddMiddleware("ObservationUpdate", negroni.HandlerFunc(middleware.FactHandler))
	s.AddMiddleware("ObservationDelete", negroni.HandlerFunc(middleware.FactHandler))

	s.AddMiddleware("MedicationStatementCreate", negroni.HandlerFunc(middleware.FactHandler))
	s.AddMiddleware("MedicationStatementCreate", middleware.GenerateRiskHandler(riskServiceEndpoint, selfUrl))
	s.AddMiddleware("MedicationStatementUpdate", negroni.HandlerFunc(middleware.FactHandler))
	s.AddMiddleware("MedicationStatementDelete", negroni.HandlerFunc(middleware.FactHandler))

	// Setup the notification handler to use the default notification definitions (and then register it)
	notificationHandler := &middleware.NotificationHandler{Registry: notifications.DefaultNotificationDefinitionRegistry}
	s.AddMiddleware("EncounterCreate", negroni.HandlerFunc(notificationHandler.Handle))

	s.Router.HandleFunc("/GroupConditionTotal/{id}", controllers.ConditionTotalHandler)
	s.Router.HandleFunc("/GroupEncounterTotal/{id}", controllers.EncounterTotalHandler)
	s.Router.HandleFunc("/GroupList/{id}", controllers.PatientListHandler)
	s.Router.HandleFunc("/InstaCount/{type}", controllers.InstaCountHandler)
	s.Router.HandleFunc("/InstaCountAll", controllers.InstaCountAllHandler)
	s.Router.HandleFunc("/NotificationCount", controllers.NotificationCountHandler)
	s.Router.HandleFunc("/Pie/{id}", controllers.PieHandler)

	login := s.Router.Path("/login").Subrouter()
	login.Methods("POST").Handler(negroni.New(negroni.HandlerFunc(controllers.LoginHandler)))
	login.Methods("DELETE").Handler(negroni.New(negroni.HandlerFunc(controllers.LogoutHandler)))

	register := s.Router.Path("/register").Subrouter()
	register.Methods("POST").Handler(negroni.New(negroni.HandlerFunc(controllers.RegisterHandler)))

	s.Run()
}
