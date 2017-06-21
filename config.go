package main

import (
	"flag"
	"log"
	"os"
	"strings"
)

type Config struct {
	PrintConfig    bool
	CodeLookup     bool
	Subscriptions  bool
	ReqLog         bool
	Icd9URL        string
	Icd10URL       string
	RiskServiceURL string
	LogDir         string
	MongoURL       string
	HostURL        string
	huddleString   string
	Huddle         []string
}

const (
	LoadCodesDesc     = "flag to enable download of icd-9 and icd-10 code lookup"
	SubscriptionsDesc = "enables limited support for resource subscriptions"
	ReqLogDesc        = "Enables request logging -- do NOT use in production"
	LogDirDesc        = "Path to a directory for ie logs to be written to."
	Icd9URLDesc       = "url for icd-9 code definition zip"
	Icd10URLDesc      = "url for icd-10 code definition zip"
	HuddleDesc        = "path to a huddle configuration file (separate multiple paths with a comma)"

	Icd9URLDefault  = "https://www.cms.gov/Medicare/Coding/ICD9ProviderDiagnosticCodes/Downloads/ICD-9-CM-v32-master-descriptions.zip"
	Icd10URLDefault = "https://www.cms.gov/Medicare/Coding/ICD10/Downloads/2016-Code-Descriptions-in-Tabular-Order.zip"

	RiskServiceURLEnv = "IE_RISK_SERVICE_URL"
	LogDirEnv         = "IE_LOG_DIR"
	MongoURLEnv       = "IE_MONGO_URL"
	HuddleEnv         = "IE_HUDDLE"
	HostURLEnv        = "IE_HOST_URL"

	RiskServiceURLDefault = "http://localhost:9000"
	MongoURLDefault       = "mongodb://localhost:27017"
	HostURLDefault        = "localhost:3001"
)

var flags *Config
var setFlags map[string]bool

func init() {
	flags = &Config{}
	flag.BoolVar(&flags.PrintConfig, "print-config", false, "prints IE's config and then exits")
	flag.BoolVar(&flags.CodeLookup, "load-codes", false, LoadCodesDesc)
	flag.BoolVar(&flags.Subscriptions, "subscriptions", false, SubscriptionsDesc)
	flag.BoolVar(&flags.ReqLog, "req-log", false, ReqLogDesc)
	flag.StringVar(&flags.LogDir, "log-dir", "", LogDirDesc)
	flag.StringVar(&flags.Icd9URL, "icd9-url", Icd9URLDefault, Icd9URLDesc)
	flag.StringVar(&flags.Icd10URL, "icd10-url", Icd10URLDefault, Icd10URLDesc)
	flag.StringVar(&flags.huddleString, "huddle", "", HuddleDesc)

	setFlags = make(map[string]bool)
}

func ConfigInit() *Config {
	flag.Parse()
	flags.MongoURL = os.Getenv(MongoURLEnv)
	if flags.MongoURL == "" {
		flags.MongoURL = MongoURLDefault
	}
	flags.LogDir = os.Getenv(LogDirEnv)
	flags.RiskServiceURL = os.Getenv(RiskServiceURLEnv)
	if flags.RiskServiceURL == "" {
		flags.RiskServiceURL = RiskServiceURLDefault
	}
	flags.HostURL = os.Getenv(HostURLEnv)
	if flags.HostURL == "" {
		flags.HostURL = HostURLDefault
	}
	hh := os.Getenv(HuddleEnv)
	flags.Huddle = strings.Split(hh, ",")
	// Since there is no easy way to check if a flag is set and
	// we don't want to overwrite set env variables with default
	// flag values, we need to see if Visit gives us these flags.
	// If Visit gives us these flags, they've been set, so we should
	// take that value instead of the env variable value.
	flag.Visit(collectSetFlags)
	if setFlags["log-dir"] {
		flags.LogDir = flags.LogDir
	}
	if setFlags["huddle"] {
		flags.Huddle = strings.Split(flags.huddleString, ",")
	}
	return flags
}

func collectSetFlags(flag *flag.Flag) {
	setFlags[flag.Name] = true
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
