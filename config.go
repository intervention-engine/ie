package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/intervention-engine/ie/app"
)

type Configuration struct {
	PrintConfiguration bool
	CodeLookup         bool
	Subscriptions      bool
	ReqLog             bool
	Icd9URL            string
	Icd10URL           string
	RiskServicesPath   string
	LogDir             string
	MongoURL           string
	HostURL            string
	huddleString       string
	HuddleConfigFiles  []string
}

const (
	LoadCodesDesc       = "flag to enable download of icd-9 and icd-10 code lookup"
	SubscriptionsDesc   = "enables limited support for resource subscriptions"
	ReqLogDesc          = "Enables request logging -- do NOT use in production"
	LogDirDesc          = "Path to a directory for ie logs to be written to."
	Icd9URLDesc         = "url for icd-9 code definition zip"
	Icd10URLDesc        = "url for icd-10 code definition zip"
	HuddleDesc          = "path to a huddle configuration file (separate multiple paths with a comma)"
	RiskServicePathDesc = "path to list of risk services JSON configuration file"

	Icd9URLDefault  = "https://www.cms.gov/Medicare/Coding/ICD9ProviderDiagnosticCodes/Downloads/ICD-9-CM-v32-master-descriptions.zip"
	Icd10URLDefault = "https://www.cms.gov/Medicare/Coding/ICD10/Downloads/2016-Code-Descriptions-in-Tabular-Order.zip"

	LogDirEnv   = "IE_LOG_DIR"
	MongoURLEnv = "IE_MONGO_URL"
	HuddleEnv   = "IE_HUDDLE"
	HostURLEnv  = "IE_HOST_URL"

	RiskServicePathDefault = "config/risk_services.json"
	MongoURLDefault        = "mongodb://localhost:27017"
	HostURLDefault         = "localhost:3001"
)

var config *Configuration
var setFlags map[string]*flag.Flag

func init() {
	config = &Configuration{}
	flag.BoolVar(&config.PrintConfiguration, "print-config", false, "prints IE's config and then exits")
	flag.BoolVar(&config.CodeLookup, "load-codes", false, LoadCodesDesc)
	flag.BoolVar(&config.Subscriptions, "subscriptions", false, SubscriptionsDesc)
	flag.BoolVar(&config.ReqLog, "req-log", false, ReqLogDesc)
	flag.StringVar(&config.LogDir, "log-dir", "", LogDirDesc)
	flag.StringVar(&config.Icd9URL, "icd9-url", Icd9URLDefault, Icd9URLDesc)
	flag.StringVar(&config.Icd10URL, "icd10-url", Icd10URLDefault, Icd10URLDesc)
	flag.StringVar(&config.huddleString, "huddle", "", HuddleDesc)
	flag.StringVar(&config.RiskServicesPath, "RiskServicesPath", RiskServicePathDefault, RiskServicePathDesc)
	setFlags = make(map[string]*flag.Flag)
}

func ConfigurationInit() *Configuration {
	flag.Parse()
	config.MongoURL = os.Getenv(MongoURLEnv)
	if config.MongoURL == "" {
		config.MongoURL = MongoURLDefault
	}
	config.LogDir = os.Getenv(LogDirEnv)
	config.RiskServicesPath = RiskServicePathDefault
	if config.RiskServicesPath == "" {
		config.RiskServicesPath = RiskServicePathDefault
	}
	config.HostURL = os.Getenv(HostURLEnv)
	if config.HostURL == "" {
		config.HostURL = HostURLDefault
	}
	hh := os.Getenv(HuddleEnv)
	config.HuddleConfigFiles = strings.Split(hh, ",")
	// Since there is no easy way to check if a flag is set and
	// we don't want to overwrite set env variables with default
	// flag values, we need to see if Visit gives us these configurations.
	// If Visit gives us these configurations, they've been set, so we should
	// take that value instead of the env variable value.
	flag.Visit(collectSetFlags)
	if f, ok := setFlags["log-dir"]; ok {
		config.LogDir = f.Value.String()
	}
	if _, ok := setFlags["huddle"]; ok {
		config.HuddleConfigFiles = strings.Split(config.huddleString, ",")
	}
	return config
}

func collectSetFlags(f *flag.Flag) {
	setFlags[f.Name] = f
}

func loadRiskServicesJSON(path string) []*app.RiskService {
	config, err := ioutil.ReadFile(path)
	var services []*app.RiskService

	if err != nil {
		panic("Invalid JSON Configuration for Risk Services")
	} else {
		json.Unmarshal(config, &services)
	}

	return services
}

func setupLogFile(path string) {
	if path != "" {
		err := os.Mkdir(path, 0755)
		if err != nil && !os.IsExist(err) {
			log.Println("Error creating log directory:" + err.Error())
		}

		lf, err := os.OpenFile(path+"/ie.log", os.O_RDWR|os.O_APPEND, 0755)
		if os.IsNotExist(err) {
			lf, err = os.Create(path + "/ie.log")
		}
		if err != nil {
			log.Println("Unable to create ie log file:" + err.Error())
		} else {
			log.SetOutput(lf)
		}
	}
}
