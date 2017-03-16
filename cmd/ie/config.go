package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/intervention-engine/fhir/server"
	"github.com/intervention-engine/ie/huddles"
	"github.com/robfig/cron"
)

// HuddlePath is a type that implements flag.Var by having a Set and String func
type huddlePath []string

func (h *huddlePath) String() string {
	return fmt.Sprint(*h)
}

func (h *huddlePath) Set(value string) error {
	*h = append(*h, value)
	return nil
}

type args struct {
	CodeLookup *bool
	SubFlag    *bool
	ReqLog     *bool
	LogFile    *string
	Icd9URL    *string
	Icd10URL   *string
	HuddlePath huddlePath
}

type vars struct {
	MongoURL       string
	RiskServiceURL string
	LogFilePath    string
	HuddlePath     huddlePath
}

var huddleFlag huddlePath

func init() {
	flag.Var(&huddleFlag, "huddle", "path to a huddle configuration file (repeat flag for multiple paths)")
}

func getArgs() args {
	a := args{}
	a.CodeLookup = flag.Bool("loadCodes", false, "flag to enable download of icd-9 and icd-10 code lookup")
	a.Icd9URL = flag.String("icd9URL", "https://www.cms.gov/Medicare/Coding/ICD9ProviderDiagnosticCodes/Downloads/ICD-9-CM-v32-master-descriptions.zip", "url for icd-9 code definition zip")
	a.Icd10URL = flag.String("icd10URL", "https://www.cms.gov/Medicare/Coding/ICD10/Downloads/2016-Code-Descriptions-in-Tabular-Order.zip", "url for icd-10 code definition zip")
	a.SubFlag = flag.Bool("subscriptions", false, "enables limited support for resource subscriptions (default: false)")
	a.ReqLog = flag.Bool("reqlog", false, "Enables request logging -- do NOT use in production")
	a.LogFile = flag.String("logdir", "", "Path to a directory for ie and gin logs to be written to.")
	flag.Parse()
	a.HuddlePath = huddleFlag
	return a
}

func getVars() vars {
	v := vars{}

	v.MongoURL = os.Getenv("MONGO_URL")
	if v.MongoURL == "" {
		v.MongoURL = "mongodb://localhost:27017"
	}

	v.RiskServiceURL = os.Getenv("RISK_SERVICE_URL")
	if v.RiskServiceURL == "" {
		v.RiskServiceURL = "http://localhost:9000"
	}

	v.LogFilePath = os.Getenv("IE_LOG_DIR")

	paths := os.Getenv("HUDDLE_CONFIGS")
	if len(paths) > 0 {
		v.HuddlePath = []string(strings.Split(paths, ","))
	}

	return v
}

func setupLogFile(s *server.FHIRServer, path string) {
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

	if path != "" {
		glf, err := os.OpenFile(path+"/gin.log", os.O_RDWR|os.O_APPEND, 0755)
		if os.IsNotExist(err) {
			glf, err = os.Create(path + "/gin.log")
		}
		if err != nil {
			log.Println("Unable to create ie log file:" + err.Error())
		} else {
			s.Engine.Use(gin.LoggerWithWriter(glf))
		}
	}
}

func configureHuddles(huddleConfigs []string) (*huddles.HuddleSchedulerController, *cron.Cron) {
	// Since the huddle controller needs info from the command line, set it up here.  When IE is refactored
	// to take out globals (and other stuff), this should be rethought.
	c := cron.New()
	huddleController := new(huddles.HuddleSchedulerController)
	for _, huddlePath := range huddleConfigs {
		data, err := ioutil.ReadFile(huddlePath)
		if err != nil {
			log.Fatalln(err)
		}
		config := new(huddles.HuddleConfig)
		if err := json.Unmarshal(data, config); err != nil {
			log.Fatalln(err)
		}
		huddleController.AddConfig(config)
		scheduleIt := func() {
			_, err := huddles.NewHuddleScheduler(config).ScheduleHuddles()
			if err != nil {
				log.Printf("ERROR: Could not schedule huddles for huddle with name %s: %v", config.Name, err)
			}
		}

		// Wait 1 minute before doing initial runs or setting up cron jobs.  This allows the server to get
		// started (since it needs to initiate the db connection, etc).
		time.AfterFunc(1*time.Minute, func() {
			// Do an initial run
			log.Println("Initial scheduling for huddle with name ", config.Name)
			scheduleIt()

			// Set up the cron job for future runs
			if config.SchedulerCronSpec != "" {
				err := c.AddFunc(config.SchedulerCronSpec, scheduleIt)
				if err != nil {
					log.Fatalln(err)
				}
				log.Printf("Huddle with name %s scheduled with cron spec: %s\n", config.Name, config.SchedulerCronSpec)
			} else {
				log.Printf("Warning: Huddle with name %s is not configured with a scheduler cron job.\n", config.Name)
			}
		})
	}
	return huddleController, c
}

func resolveHuddleConfig(argPath []string, varPath []string) []string {
	var p []string
	if len(varPath) != 0 {
		p = varPath
	}
	if len(argPath) != 0 {
		p = argPath
	}
	return p
}
