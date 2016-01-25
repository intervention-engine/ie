package main

import (
	"flag"
	"fmt"
	"net"
	"os"

	"github.com/intervention-engine/fhir/server"
	"github.com/intervention-engine/ie/controllers"
	"github.com/intervention-engine/ie/utilities"
)

//var Store sessions.Store

func main() {
	// Check for a linked MongoDB container if we are running in Docker
	mongoHost := os.Getenv("MONGO_PORT_27017_TCP_ADDR")
	if mongoHost == "" {
		mongoHost = "localhost"
	}

	codeLookupFlag := flag.Bool("loadCodes", false, "flag to enable download of icd-9 code lookup")
	icd9URL := flag.String("icd9URL", "https://www.cms.gov/Medicare/Coding/ICD9ProviderDiagnosticCodes/Downloads/ICD-9-CM-v32-master-descriptions.zip", "url for icd-9 code definition zip")
	flag.Parse()

	parsedLookupFlag := *codeLookupFlag
	parsedICD9LookupURL := *icd9URL

	if parsedLookupFlag {
		utilities.LoadICD9FromCMS(mongoHost, parsedICD9LookupURL)
	}

	riskServiceEndpoint := os.Getenv("RISKSERVICE_PORT_9000_TCP_ADDR")
	if riskServiceEndpoint == "" {
		riskServiceEndpoint = "http://localhost:9000"
	} else {
		riskServiceEndpoint = "http://" + riskServiceEndpoint + ":9000"
	}

	var ip net.IP
	var selfURL string
	host, err := os.Hostname()
	if err != nil {
		panic(err)
	}
	addrs, err := net.LookupIP(host)
	if err != nil {
		fmt.Println("Unable to lookup IP based on hostname, defaulting to localhost.")
		selfURL = "http://localhost:3001"
	}
	for _, addr := range addrs {
		if ipv4 := addr.To4(); ipv4 != nil {
			ip = ipv4
			selfURL = "http://" + ip.String() + ":3001"
		}
	}

	s := server.NewServer(mongoHost)

	controllers.RegisterRoutes(s, selfURL, riskServiceEndpoint)

	s.Run()
}
