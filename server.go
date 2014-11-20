package main

import (
	"github.com/codegangsta/negroni"
	"gitlab.mitre.org/intervention-engine/fhir/server"
	"gitlab.mitre.org/intervention-engine/ie/middleware"
)

func main() {
	s := server.FHIRServer{DatabaseHost: "localhost", Middleware: make([]negroni.Handler, 0)}
	s.AddMiddleware(negroni.HandlerFunc(middleware.PatientHandler))
	s.AddMiddleware(negroni.HandlerFunc(middleware.FactHandler))

	s.Run()
}
