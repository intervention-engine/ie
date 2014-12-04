package main

import (
	"github.com/codegangsta/negroni"
	"gitlab.mitre.org/intervention-engine/fhir/server"
	"gitlab.mitre.org/intervention-engine/ie/middleware"
)

func main() {
	s := server.FHIRServer{DatabaseHost: "localhost", MiddlewareConfig: make(map[string][]negroni.Handler)}
	//s.AddMiddleware("QueryCreate", negroni.HandlerFunc(middleware.FactHandler))
	s.AddMiddleware("QueryCreate", negroni.HandlerFunc(middleware.QueryExecutionHandler))

	s.Run()
}
