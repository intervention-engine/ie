package main

import (
	"log"
	"net"
	"os"

	"gopkg.in/mgo.v2"

	"github.com/intervention-engine/fhir/server"
	"github.com/intervention-engine/ie/utilities"
	"github.com/intervention-engine/ie/web"
)

func main() {
	args := getArgs()
	vars := getVars()
	if *args.CodeLookup {
		utilities.LoadICDFromCMS(vars.MongoURL, *args.Icd9URL, "ICD-9")
		utilities.LoadICDFromCMS(vars.MongoURL, *args.Icd10URL, "ICD-10")
	}

	var ip net.IP
	var selfURL string
	host, err := os.Hostname()
	if err != nil {
		log.Fatalln(err)
	}
	addrs, err := net.LookupIP(host)
	if err != nil {
		log.Println("Unable to lookup IP based on hostname, defaulting to localhost.")
		selfURL = "http://localhost:3001"
	}
	for _, addr := range addrs {
		if ipv4 := addr.To4(); ipv4 != nil {
			ip = ipv4
			selfURL = "http://" + ip.String() + ":3001"
		}
	}

	s := server.NewServer(vars.MongoURL)
	setupLogFile(s, vars.LogFilePath)

	ieSession, err := mgo.Dial(vars.MongoURL)
	if err != nil {
		log.Fatalln("dialing mongo failed for ie session at url: %s", vars.MongoURL)
	}
	defer ieSession.Close()

	web.RegisterAPIRoutes(s.Engine, ieSession)

	if *args.ReqLog {
		s.Engine.Use(server.RequestLoggerHandler)
	}

	huddleConfig := resolveHuddleConfig(args.HuddlePath, vars.HuddlePath)
	hc, cronJob := configureHuddles(huddleConfig)

	s.Engine.GET("/ScheduleHuddles", hc.ScheduleHandler)

	if len(huddleConfig) > 0 {
		cronJob.Start()
		defer cronJob.Stop()
	}

	closer := web.RegisterRoutes(s, selfURL, vars.RiskServiceURL, *args.SubFlag)
	defer closer()

	s.Run(server.DefaultConfig)
}
